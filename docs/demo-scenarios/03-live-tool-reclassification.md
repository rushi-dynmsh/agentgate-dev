# Scenario 3: Live Tool Risk Reclassification (Dynamic Governance)

## 1. What This Proves (What to Say)
> "In traditional architectures, changing an API endpoint's security classification requires editing code, changing config maps, and redeploying or restarting microservices. With AgentGate, security teams can reclassify tool risk in real time through the console. Changes take effect on the very next incoming packet without interrupting running services."

---

## 2. Pre-Conditions
- Admin Console is open at `https://localhost:8444/tools`.
- Stack is running.

---

## 3. Step-by-Step Execution

**Correction (2026-10-14): the tool used here is `update_ticket_status`, not `close_ticket` — there is no `close_ticket` tool anywhere in the real system (checked against `deploy/demo/support-desk-mcp/main.go` — note the tool *server's* folder is named `support-desk-mcp`, but it registers itself in the live catalog under the backend id `mcp-probe`, which is the name that actually appears in the Tools and Audit pages), so the original version of this scenario would fail with "unknown tool" instead of demonstrating reclassification. Also use the `agent-admin` identity, not `agent-reader`: under the baseline policy, `update_ticket_status` is risk `write`, and a reader is never allowed `write` — so a reader's call would already be denied before you touch anything, with nothing visibly changing when you reclassify. Admin is allowed `read` and `write` but explicitly forbidden from `destructive`, which is what makes the before/after actually visible.**

### Step 1: Observe the Initial State (Allowed Call)
In terminal, call `update_ticket_status` as `agent-admin`. Under the baseline policy it is classified `write`, which admin is allowed:

```powershell
$agentSecret = (Get-Content deploy/prod/secrets/agent_admin_secret).Trim()
$tokenResp = curl.exe -k -s -X POST https://localhost:8443/realms/agentgate/protocol/openid-connect/token `
    -H "Content-Type: application/x-www-form-urlencoded" `
    -d "grant_type=client_credentials&client_id=agent-admin&client_secret=$agentSecret"
$agentToken = ($tokenResp | ConvertFrom-Json).access_token

$payload = '{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"update_ticket_status\",\"arguments\":{\"ticket_id\":\"TCK-1002\",\"new_status\":\"resolved\"}}}'
curl.exe -k -s -i -X POST https://localhost:3443/ `
    -H "Authorization: Bearer $agentToken" `
    -H "Content-Type: application/json" `
    -d $payload
```
*(Expected: `200 OK` — admin is allowed both read and write tools under the baseline policy.)*

---

### Step 2: Reclassify the Tool Live in the UI
Now switch to your browser:
1. Go to **`https://localhost:8444/tools`** (Tools & Resources).
2. Look at the `mcp-probe` section.
3. Locate the tool named **`update_ticket_status`**.
4. Click the **Edit** button next to it.
5. In the dialog:
   - Change the **Risk level** dropdown to **`destructive`**.
   - Check the confirmation box: **"I understand this changes authorization for all agents"** (only appears because you're moving *to* destructive).
   - Click **Save**.
6. The table refreshes immediately showing `update_ticket_status` with a red **`destructive`** badge.

---

### Step 3: Trigger the Tool Call Again (Immediate Denial)
Return to your PowerShell terminal and run the exact same command from Step 1 again (reuse the `$payload` and `$agentToken` variables already in your session):

```powershell
curl.exe -k -s -i -X POST https://localhost:3443/ `
    -H "Authorization: Bearer $agentToken" `
    -H "Content-Type: application/json" `
    -d $payload
```

---

## 4. What to Look For (Expected Result)

### In the Terminal:
Instead of `200 OK`, you now immediately receive (same verified response shape as Scenario 2):
```http
HTTP/1.1 403 Forbidden
x-agentgate-decision: deny
x-agentgate-reason: policy_deny

AgentGate Authorization Denied: policy_deny (an explicit policy denied this request)
```

### In the Admin Console UI:
1. Navigate to **Audit** (`https://localhost:8444/audit`).
2. The latest event shows:
   - **Decision:** **`DENY`**.
   - **Tool:** `mcp-probe/update_ticket_status`.
   - **Evaluated Risk Level:** **`destructive`**.
3. **Point out to audience:** We did not restart the gateway, we did not redeploy the backend MCP container, and we did not touch any code. The control plane immediately enforced the new risk boundary.

---

## 5. Clean-Up for Next Scenarios
You can either leave it as destructive, or click **Edit** on `update_ticket_status` in `https://localhost:8444/tools` and restore its risk level back to `write`, or run `.\deploy\prod\scripts\demo-mode.ps1 -Reset` later (this also restores every other tool's bootstrap classification and reactivates policy `v1.0.0`).
