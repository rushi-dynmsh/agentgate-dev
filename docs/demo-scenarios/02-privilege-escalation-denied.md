# Scenario 2: Privilege Escalation Denied (Policy Block)

## 1. What This Proves (What to Say)
> "Here we demonstrate zero-trust least privilege. An AI agent compromised by prompt injection or rogue instructions attempts to delete a customer ticket (`delete_ticket`). Because this agent only possesses read-level authorization, AgentGate intercepts and rejects the request at the perimeter. The backend tool never receives the packet."

---

## 2. Pre-Conditions
- The stack is running (`.\deploy\prod\scripts\demo-mode.ps1 -Reset`).
- Admin Console is open at `https://localhost:8444/audit`.

---

## 3. Step-by-Step Execution

Open PowerShell in the repo root and run this command:

```powershell
# 1. Obtain token for 'agent-reader'
$agentSecret = (Get-Content deploy/prod/secrets/agent_reader_secret).Trim()
$tokenResp = curl.exe -k -s -X POST https://localhost:8443/realms/agentgate/protocol/openid-connect/token `
    -H "Content-Type: application/x-www-form-urlencoded" `
    -d "grant_type=client_credentials&client_id=agent-reader&client_secret=$agentSecret"
$agentToken = ($tokenResp | ConvertFrom-Json).access_token

# 2. Attempt to call destructive tool 'delete_ticket' as agent-reader
curl.exe -k -s -i -X POST https://localhost:3443/ `
    -H "Authorization: Bearer $agentToken" `
    -H "Content-Type: application/json" `
    -d '{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"delete_ticket\",\"arguments\":{\"ticket_id\":\"TCK-1001\"}}}'
```

---

## 4. What to Look For (Expected Result)

### In the Terminal:
You will immediately receive an HTTP `403 Forbidden` response (verified live on 2026-10-14, this is the real, exact response — not an example):
```http
HTTP/1.1 403 Forbidden
x-agentgate-decision: deny
x-agentgate-reason: policy_deny
content-type: application/json

{"jsonrpc":"2.0","id":2,"error":{"code":-40301,"data":{"execution_id":"exec-…","reason":"policy_deny"},"message":"AgentGate Authorization Denied: policy_deny (an explicit policy denied this request)"}}
```
*(Notice: The response comes directly from the gateway. The target mcp-probe server was never contacted).*

### In the Admin Console UI:
1. Go to **`https://localhost:8444/audit`**.
2. Click the **Decision** dropdown filter and select **`DENY`**.
3. Locate the latest event row:
   - **Decision:** Red badge displaying **`DENY`**.
   - **Tool:** `mcp-probe/delete_ticket`.
   - **Principal:** `agent-reader`.
4. Click the row to expand the details:
   - **Risk Level:** `destructive`.
   - **Policy Reason:** Displays the Cedar evaluation denial: `forbid-wins` / `unauthorized_role`.
   - **Payload Hash:** Confirms the exact request argument hash that was rejected.

---

## 5. What Happened Behind the Scenes?
1. The gateway received the `delete_ticket` call from `agent-reader`.
2. AgentGate looked up `delete_ticket` in the Tool Catalog and found `risk: destructive`.
3. Cedar evaluated the active policy (`v1.0.0`). Here is the real policy text (`agentgate/internal/fixturepolicy/fixturepolicy.go`), not a paraphrase:
   ```cedar
   forbid(
     principal,
     action == AgentGate::Action::"InvokeTool",
     resource
   ) when {
     resource.risk == "destructive"
   };
   ```
4. This explicit `forbid` applies to every role, with no exception — unlike the `permit` rules for `reader`/`admin`/`payer`, there is no role that is allowed a `destructive` tool. Cedar evaluated to **FORBID**.
5. AgentGate denied the request immediately, logged the denial to the immutable PostgreSQL audit chain, and told the gateway to terminate the connection with `403 Forbidden`.
