# AgentGate Demo — Start to Finish

Follow this top to bottom on demo day. Every command is Windows PowerShell, run from the repo root. For what to *say*, use `docs/DEMO_SCRIPT.md`; for exact commands per scenario, use `docs/demo-scenarios/`.

## What is the demo MCP server?

`support-desk-mcp` is a small MCP server **we built** for the demo (`deploy/demo/support-desk-mcp/`). It pretends to be a customer-support ticket system with three tools:

| Tool | Risk | What it does |
|---|---|---|
| `list_tickets` | read | Lists three fake tickets |
| `update_ticket_status` | write | Changes a ticket's status |
| `delete_ticket` | destructive | Deletes a ticket |

It starts automatically with everything else and keeps its data in memory (`-Reset` restores the three tickets). It runs only inside Docker, so you never connect to it directly. Everything goes through the gateway:

```
Inspector / agent  ->  gateway (https://localhost:3443/mcp)  ->  AgentGate checks the call  ->  support-desk-mcp
```

## Part 1 — One-time setup (first run on a machine only)

1. **Docker Desktop** installed. **Node.js 22.19 or newer** (`node --version`).
2. **Repo on the right branch** (the branch with the Inspector support):
   ```powershell
   git checkout feat/mcp-client-compat
   ```
3. **Browser login fix for the admin console** (needed only for the admin console's "Sign in with SSO"; Inspector doesn't need it). Open Notepad **as Administrator**, open `C:\Windows\System32\drivers\etc\hosts`, add this line at the end, save, then run `ipconfig /flushdns`:
   ```
   127.0.0.1 keycloak
   ```
4. **Build the documentation site once:**
   ```powershell
   cd website
   npm ci
   npm run build
   cd ..
   ```

## Part 2 — Start everything (every demo day, ~10 minutes the first time, ~2 minutes after)

1. Open **Docker Desktop** and wait until it says the engine is running.
2. Start the stack and put it in a clean state:
   ```powershell
   .\deploy\prod\scripts\demo-mode.ps1 -Reset
   ```
   Wait for `AgentGate Demo Environment is Running`. The first run builds images, so it is slow. Red text like `NativeCommandError` between the Docker lines is harmless **if** the final banner appears.
3. Start the documentation site in a **second** PowerShell window and leave it running:
   ```powershell
   cd website
   npm run serve -- --port 3000
   ```

## Part 3 — Check it is healthy (1 minute)

```powershell
docker ps --format "table {{.Names}}\t{{.Status}}"
curl.exe -k -s -o NUL -w "admin console: %{http_code}`n" https://localhost:8444/
curl.exe -k -s -o NUL -w "gateway (no token, expect 401): %{http_code}`n" https://localhost:3443/mcp
curl.exe -s -o NUL -w "docs site: %{http_code}`n" http://localhost:3000/
```
Expected: seven `prod-…` containers `Up` (`prod-keycloak` and `prod-agentgate` say `healthy`), then `200`, `401`, `200`. The gateway answering 401 is correct: it means it is up and refusing callers with no token.

Also log in to the admin console once now: open `https://localhost:8444`, accept the certificate warning (Advanced, then Proceed), click **Sign in with SSO**, user `admin`, password from:
```powershell
Get-Content deploy\prod\secrets\admin_initial_password
```

## Part 4 — The demo, in order

1. **Admin console** (`https://localhost:8444`): show the Dashboard and the Audit page. The background demo agent creates live traffic.
2. **Real MCP client** (this is the main proof). First pause the background traffic generator, so the Audit page shows only your own calls:
   ```powershell
   docker stop prod-demo-agent
   ```
   Then, in a PowerShell window (the token lasts **5 minutes**, so run it right before you show it):
   ```powershell
   $agentSecret = (Get-Content deploy\prod\secrets\agent_reader_secret).Trim()
   $tokenResp = curl.exe -k -s -X POST https://localhost:8443/realms/agentgate/protocol/openid-connect/token -H "Content-Type: application/x-www-form-urlencoded" -d "grant_type=client_credentials&client_id=agent-reader&client_secret=$agentSecret"
   $agentToken = ($tokenResp | ConvertFrom-Json).access_token
   Write-Host "token length: $($agentToken.Length) (must be about 1100, not 0)"
   $env:NODE_TLS_REJECT_UNAUTHORIZED = "0"
   npx @modelcontextprotocol/inspector --server-url https://localhost:3443/mcp --transport http --protocol-era modern --header "Authorization: Bearer $agentToken"
   ```
   In the browser tab that opens: confirm the green **Connected**, click **Tools** (three tools listed on the left). Click `list_tickets`, leave the optional `status_filter` empty, click **Execute Tool** (works: three tickets). Then click `delete_ticket`, type `TCK-1001` in `ticket_id`, click **Execute Tool** (blocked with a JSON-RPC error). Required fields are marked `*`, and the button stays grey until they are filled. Do **not** restart the traffic generator yet. Then go to the admin console's Audit page and point at the matching ALLOW and DENY rows. Details and the command-line fallback: `docs/demo-scenarios/01-real-client-inspector.md`.
3. **Show your calls in the audit trail:** open the Audit page and point at your ALLOW and DENY rows (they are the only recent ones). Then restart the traffic generator: `docker start prod-demo-agent`.
4. **Change a rule live:** `docs/DEMO_SCRIPT.md` Step 4 (reclassify a tool) and Step 5 (policy dry-run and rollback).
5. **Prove the log can't be tampered with:** `docker exec prod-agentgate /bin/agentgate verify-audit -workspace default` prints `OK <N> records`.

## Part 5 — Between rehearsals, and shutting down

- **Back to a clean state** (restores tickets, tool risk levels and the active policy):
  ```powershell
  .\deploy\prod\scripts\demo-mode.ps1 -Reset
  ```
- **Stop everything, keep data:** `.\deploy\prod\scripts\demo-mode.ps1 -Down`
- **Stop and wipe all data:** `.\deploy\prod\scripts\demo-mode.ps1 -Down -Volumes` (the next start takes the slow path again).
- Close the documentation-site window (`Ctrl+C`).

## If something goes wrong

| What you see | Why | Fix |
|---|---|---|
| Docker commands hang or time out | Docker Desktop's engine is stuck | Quit Docker Desktop, run `wsl --shutdown`, start Docker Desktop again, rerun Part 2 |
| Admin console login: `DNS_PROBE_FINISHED_NXDOMAIN` | The hosts-file line is missing | Part 1, step 3 |
| Admin console login: `HTTP ERROR 431` | Old cookies piled up | Clear site data for `localhost`, or use an Incognito window |
| Inspector: `401` or "Connection failed" | The 5-minute token expired | Re-run the token lines and relaunch Inspector |
| Inspector shows **Failed** with many 401s on `/register` and `.well-known` | The token was empty, so Inspector found no bearer token. Usually PowerShell is not in the repo root (the secret path is relative), or the stack is down | `cd` to the repo root, re-run the token lines, check `token length` prints about 1100, relaunch Inspector |
| Inspector: "Version negotiation failed" or an `initialize` error | `--protocol-era modern` is missing | Add it; the server speaks the newer MCP style only |
| Inspector prints a TLS warning, or `Assertion failed … async.c` on exit | Expected: self-signed certificate, and a Node-on-Windows exit message | Ignore both |
| `delete_ticket` was actually allowed, or tickets are missing | An earlier rehearsal changed the data or rules | Run `demo-mode.ps1 -Reset` |
| Gateway check returns something other than 401 with no token | A service isn't ready | Wait a minute, rerun Part 3; if it persists, rerun Part 2 |

## Know the limits (say them before someone asks)

- The tool server is our own demo, not a third-party product. Connecting other MCP servers is the next milestone.
- The tool list shows every tool, even ones an agent can't use; the forbidden ones are blocked when called.
- Only agents using the newer MCP style (2026-07-28) can connect.
