# Scenario 1: A Real MCP Client (MCP Inspector) — Allowed, Blocked, Invented Tool

## 1. What This Proves (What to Say)
> "This is MCP Inspector, the official tool the MCP community uses to test servers. It connects to our gateway the way any real AI agent would. An allowed call goes through. A call the agent isn't permitted to make is blocked, with a clear reason. And a tool the agent simply made up never reaches anything. The agent doesn't need to know a security layer is in the path: to it, the gateway is just an MCP server."

---

## 2. Pre-Conditions
- The stack is running (`.\deploy\prod\scripts\demo-mode.ps1 -Reset`). Setup and health checks: `docs/DEMO_START_TO_FINISH.md`.
- Node.js 22.19 or newer (`node --version`). Inspector runs through `npx`; nothing to install.
- You are in the repo root in PowerShell, with the admin console open at `https://localhost:8444/audit` in a browser tab.
- Run Inspector once beforehand so `npx` has downloaded it. It needs internet the first time; the curl fallback in Step 6 needs nothing.

---

## 3. Step-by-Step Execution

### Step 1: Get the agent's token and open Inspector
An agent is a machine, so it logs in with its own client ID and secret, not a browser login. The token lasts **5 minutes**, so run this block right before you demo and re-run it if Inspector starts failing.

```powershell
$agentSecret = (Get-Content deploy\prod\secrets\agent_reader_secret).Trim()
$tokenResp = curl.exe -k -s -X POST https://localhost:8443/realms/agentgate/protocol/openid-connect/token -H "Content-Type: application/x-www-form-urlencoded" -d "grant_type=client_credentials&client_id=agent-reader&client_secret=$agentSecret"
$agentToken = ($tokenResp | ConvertFrom-Json).access_token
Write-Host "token length: $($agentToken.Length) (must be about 1100, not 0)"

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
1. The header shows a green **Connected** and the server name `support-desk-mcp`.
2. Click **Tools** at the top. The list on the left reads **Tools (3)**: `list_tickets`, `update_ticket_status`, `delete_ticket`. Click a tool to open its form in the middle.
3. The **Messages** panel on the right (marked MODERN) logs every request and response. Keep it visible: it shows exactly what was sent and what came back.

**How to run any tool:** click it in the left list, fill in the fields in the middle panel, click **Execute Tool**. Fields marked with a red `*` are required, and **Execute Tool** stays greyed out until they are filled. (The **Edit as JSON** switch lets you type the arguments as JSON instead of filling the boxes.)

| Tool | Fields to enter | Safe values for the demo |
|---|---|---|
| `list_tickets` | `status_filter` (optional) | Leave empty for all tickets, or type `open`, `in_progress` or `resolved` to filter |
| `update_ticket_status` | `ticket_id` *, `new_status` * | `TCK-1002` and `resolved` |
| `delete_ticket` | `ticket_id` * | `TCK-1001` |

The three demo tickets are `TCK-1001` (open), `TCK-1002` (in progress) and `TCK-1003` (open).

### Step 3: Run an allowed tool
1. Click `list_tickets`, leave `status_filter` empty, and click **Execute Tool**.
2. The **Response** box on the right shows the three demo tickets (`"text": "Found 3 ticket(s)"`, then a `tickets` list).
3. In the admin console Audit page, the top row shows **`ALLOW`**, tool `mcp-probe/list_tickets`, principal `agent-reader`. Expand it: risk level `read`, policy version `v1.0.0`, and an **Audit Chain** hash that links this record to the one before it.

### Step 4: Run a forbidden tool
1. Click `delete_ticket`, type `TCK-1001` into the `ticket_id` box, and click **Execute Tool**. This agent only has read access.
2. You get a JSON-RPC error instead of a result (shown in the Response box and as a red error in Inspector):
   ```json
   {"jsonrpc":"2.0","id":2,"error":{"code":-40301,"data":{"execution_id":"exec-…","reason":"policy_deny"},"message":"AgentGate Authorization Denied: policy_deny (an explicit policy denied this request)"}}
   ```
3. No ticket was deleted; the backend never received the call.
4. In the Audit page, filter by **`DENY`**. The newest row shows tool `mcp-probe/delete_ticket`, principal `agent-reader`, risk level `destructive`, and the reason `policy_deny`.

### Step 5: Try a tool that doesn't exist (an AI "hallucinating" or being tricked)
Inspector can't do this one: it only lets you run tools the server listed, and refuses to send anything else (`Tool 'dump_customer_database' not found on server`). That is a courtesy of a well-behaved client. A misbehaving or hijacked agent won't be so polite, so this step uses a plain request to show what the **gateway** does:

```powershell
curl.exe -k -s -i -X POST https://localhost:3443/mcp -H "Authorization: Bearer $agentToken" -H "Content-Type: application/json" -d '{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"dump_customer_database\",\"arguments\":{\"format\":\"csv\"}}}'
```
Expected: `403 Forbidden` with a JSON-RPC error whose `reason` is `unknown_tool`:
```json
{"jsonrpc":"2.0","id":4,"error":{"code":-40301,"data":{"execution_id":"exec-…","reason":"unknown_tool"},"message":"AgentGate Authorization Denied: unknown_tool (tool mcp-probe/dump_customer_database is not registered in governance catalog)"}}
```
In the Audit page (filter `DENY`) the row shows tool `dump_customer_database`, reason `unknown_tool`, and the request's arguments (`{"format":"csv"}`) are recorded, so a security analyst can see exactly what was attempted.

### Optional: the admin agent
Repeat Step 1 with `agent_admin_secret` and `client_id=agent-admin`. `update_ticket_status` with `ticket_id` = `TCK-1002` and `new_status` = `resolved` is now allowed, and `delete_ticket` is still denied (nobody may use destructive tools).

### Step 6 (fallback): the same thing without Inspector
If Node or internet is a problem, nothing above needs Inspector. After the token lines from Step 1:

```powershell
# Allowed: expect 200 and the three tickets
curl.exe -k -s -i -X POST https://localhost:3443/mcp -H "Authorization: Bearer $agentToken" -H "Content-Type: application/json" -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"list_tickets\",\"arguments\":{}}}'
# Forbidden: expect 403 with reason policy_deny
curl.exe -k -s -i -X POST https://localhost:3443/mcp -H "Authorization: Bearer $agentToken" -H "Content-Type: application/json" -d '{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"delete_ticket\",\"arguments\":{\"ticket_id\":\"TCK-1001\"}}}'
```
Step 5's command is already a plain request. Inspector's own command-line mode (verified live on 2026-10-15) does the same as the browser:

```powershell
$base = @("-y","@modelcontextprotocol/inspector","--cli","--server-url","https://localhost:3443/mcp","--transport","http","--protocol-era","modern","--header","Authorization: Bearer $agentToken")
npx @base --method tools/list
npx @base --method tools/call --tool-name list_tickets
npx @base --method tools/call --tool-name delete_ticket --tool-arg ticket_id=TCK-1001
```
(A harmless `Assertion failed … async.c` line can print when Node exits on Windows; ignore it.)

---

## 4. What to Look For (Expected Result)

| Call | Result | Audit page |
|---|---|---|
| `tools/list`, `server/discover` | Tools listed | Not an audit row (counted in metrics) |
| `list_tickets` | Three tickets | `ALLOW`, policy `v1.0.0` |
| `delete_ticket` (reader) | JSON-RPC error, `policy_deny` | `DENY`, risk `destructive` |
| `dump_customer_database` | JSON-RPC error, `unknown_tool` | `DENY`, arguments recorded |

---

## 5. What Happened Behind the Scenes?
1. The agent sent each request to the gateway with its token. The gateway checked the token's signature against Keycloak.
2. The gateway asked AgentGate about **every** request.
3. For `server/discover` and `tools/list` (questions that only describe the server), AgentGate says yes immediately. It needs no tool decision, and it strips the agent's token so the backend never sees it. These are counted in metrics (`reason="discovery_allow"`), not written to the audit chain.
4. For `tools/call`, AgentGate does the full check:
   - It looks the tool up in the catalog. An unregistered tool is refused here, before any policy runs (this is Step 5). It fails closed: unknown means denied.
   - It evaluates the active policy. The real rule that blocks Step 4 is this one, which applies to every role with no exception (`agentgate/internal/fixturepolicy/fixturepolicy.go`):
     ```cedar
     forbid(
       principal,
       action == AgentGate::Action::"InvokeTool",
       resource
     ) when {
       resource.risk == "destructive"
     };
     ```
   - The rule that allows Step 3 is `permit(principal in AgentGate::Role::"reader", action == AgentGate::Action::"InvokeTool", resource) when { resource.risk == "read" };`.
   - For an allowed call, AgentGate mints a fresh short-lived credential (a signed token valid for 60 seconds), swaps it in for the agent's own token, and the call is forwarded. A record is written to the tamper-evident audit chain **before** anything is allowed.
5. Everything else (`resources/read`, `prompts/get`, `subscriptions/listen`, `initialize`, any unknown method) is denied. The method is read from the request body, so a request that merely **labels** itself as a harmless list call but contains a delete is still treated as a delete. The live security suite checks this (`TestV1_19`–`TestV1_22`).

---

## 6. Honest Limits to Mention if Asked
- The tool list shows **every** tool, not only the ones this agent may call. Forbidden ones are still blocked when called.
- The backend here is our own demo server. Pointing the gateway at a third-party MCP server needs per-backend credential settings that are planned, not built.
- The tool catalog is still filled in by hand; automatic discovery with "unclassified until approved" tools is planned.
- Only agents speaking the newer MCP style (2026-07-28) can connect.
