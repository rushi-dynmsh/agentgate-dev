# Scenario 8: A Real MCP Client (MCP Inspector) Through the Gateway

## 1. What This Proves (What to Say)
> "So far we sent hand-built requests. This time we use MCP Inspector, the official reference tool for testing MCP servers. It connects to our gateway like any real AI agent would: it asks what tools exist, calls an allowed one, and gets blocked on a forbidden one. The agent doesn't need to know a security layer is in the path. To it, the gateway is just an MCP server."

---

## 2. Pre-Conditions
- The stack is running (`.\deploy\prod\scripts\demo-mode.ps1 -Reset`).
- Node.js 22.19 or newer (`node --version`). Inspector runs through `npx`; nothing to install.
- You are in the repo root in PowerShell.

---

## 3. Step-by-Step Execution

### Step 1: Get the agent's token and open Inspector
An agent is a machine, so it logs in with its own client ID and secret, not a browser login. The token lasts **5 minutes**, so run this block right before you demo and re-run it if Inspector starts failing.

```powershell
$agentSecret = (Get-Content deploy\prod\secrets\agent_reader_secret).Trim()
$tokenResp = curl.exe -k -s -X POST https://localhost:8443/realms/agentgate/protocol/openid-connect/token -H "Content-Type: application/x-www-form-urlencoded" -d "grant_type=client_credentials&client_id=agent-reader&client_secret=$agentSecret"
$agentToken = ($tokenResp | ConvertFrom-Json).access_token

$env:NODE_TLS_REJECT_UNAUTHORIZED = "0"
npx @modelcontextprotocol/inspector --server-url https://localhost:3443/mcp --transport http --protocol-era modern --header "Authorization: Bearer $agentToken"
```

What each part does:
- `https://localhost:3443/mcp` is the gateway. (`/mcp` is the conventional MCP path; the gateway also accepts the bare `/`.)
- `--header "Authorization: Bearer …"` is the agent's token. Inspector's own browser-login (OAuth) settings are **not** used; leave them off.
- `--protocol-era modern` matters. Inspector defaults to an older style that opens with an `initialize` handshake. The MCP revision AgentGate speaks (2026-07-28) has no handshake, so the default would be refused.
- `NODE_TLS_REJECT_UNAUTHORIZED=0` accepts the demo's self-signed certificate for **this terminal only**. Don't set it permanently.

Inspector opens in your browser. The server card is read-only because it was launched from the command line; that's expected.

### Step 2: See the tools
1. The server card shows **Connected**.
2. Open the **Tools** tab and list tools. You should see `list_tickets`, `update_ticket_status` and `delete_ticket`.

### Step 3: Run an allowed tool
1. Pick `list_tickets` and run it with no arguments.
2. You get the three demo tickets back.

### Step 4: Run a forbidden tool
1. Pick `delete_ticket` and run it with `{"ticket_id": "TCK-1001"}`.
2. You get a JSON-RPC error instead of a result:
   ```json
   {"jsonrpc":"2.0","id":2,"error":{"code":-40301,"data":{"execution_id":"exec-…","reason":"policy_deny"},"message":"AgentGate Authorization Denied: policy_deny (an explicit policy denied this request)"}}
   ```
3. No ticket was deleted.

### Step 5: See it in the audit log
Open `https://localhost:8444/audit`. There is an `ALLOW` row for `list_tickets` and a `DENY` row for `delete_ticket`, both for `agent-reader`.

### Optional: the admin agent
Repeat Step 1 with `agent_admin_secret` and `client_id=agent-admin`. `update_ticket_status` with `{"ticket_id":"TCK-1002","new_status":"resolved"}` is now allowed, and `delete_ticket` is still denied (nobody may use destructive tools).

### Same thing without the browser (Inspector's command-line mode)
These are the exact calls verified against the live stack on 2026-10-15; use them if the browser misbehaves. Run Step 1's first three lines first, then:

```powershell
$base = @("-y","@modelcontextprotocol/inspector","--cli","--server-url","https://localhost:3443/mcp","--transport","http","--protocol-era","modern","--header","Authorization: Bearer $agentToken")
npx @base --method tools/list
npx @base --method tools/call --tool-name list_tickets
npx @base --method tools/call --tool-name delete_ticket --tool-arg ticket_id=TCK-1001
```
(A harmless `Assertion failed … async.c` line can print when Node exits on Windows; ignore it.)

---

## 4. What to Look For (Expected Result)
- `tools/list` and `list_tickets` succeed; `delete_ticket` returns the JSON-RPC error above.
- The same `ALLOW` and `DENY` appear in the audit log.

---

## 5. What Happened Behind the Scenes?
1. Inspector sent each request to the gateway with the agent's token. The gateway checked the token's signature against Keycloak.
2. The gateway asked AgentGate about **every** request.
3. For `server/discover` and `tools/list` (questions that only describe the server), AgentGate says yes immediately. It does not need a tool decision, and it strips the agent's token so the backend never sees it. These are counted in metrics (`reason="discovery_allow"`), not written to the audit chain.
4. For `tools/call`, AgentGate does the full check exactly as in Scenarios 1 and 2: who is the agent, is the tool known, what do the policy rules say, and an audit record is written before anything is allowed.
5. Everything else (`resources/read`, `prompts/get`, `subscriptions/listen`, `initialize`, any unknown method) is denied. The method is read from the request body, so a request that merely **labels** itself as a harmless list call but contains a delete is still treated as a delete. The live security suite checks this (`TestV1_19`–`TestV1_22`).

---

## 6. Honest Limits to Mention if Asked
- The tool list shows **every** tool, not only the ones this agent may call. Forbidden ones are still blocked when called.
- The backend here is our own demo server. Pointing the gateway at a third-party MCP server needs per-backend credential settings that are planned, not built.
- The tool catalog is still filled in by hand; automatic discovery with "unclassified until approved" tools is planned.
