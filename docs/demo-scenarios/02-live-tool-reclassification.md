# Scenario 2: Live Tool Risk Reclassification (Dynamic Governance)

## 1. What This Proves (What to Say)
> "In traditional setups, changing how risky an API call is means editing code or config and redeploying. With AgentGate, a security team changes a tool's risk level in the console, and the very next call from the agent is judged by the new rule. Nothing restarts."

---

## 2. Pre-Conditions
- The stack is running, and you did the preparation in `docs/DEMO_START_TO_FINISH.md` Part 4.0: traffic generator stopped, Inspector connected as the **reader** agent, the admin console open on **Tools & resources** (`https://localhost:8444/tools`).
- Nothing else is needed. The whole scenario uses the reader token from Scenario 1.

---

## 3. Step-by-Step Execution

**Why this works with a reader:** the baseline policy lets a reader use only `read` tools. `list_tickets` is classified `read`, so it works today. If we reclassify the same tool to `write`, the very same call is no longer something a reader may do. One tool, one agent, a before and an after.

### Step 1: The call works today
In Inspector, click `list_tickets` and **Execute Tool** (no input). You get the three tickets.

### Step 2: Reclassify the tool in the console
1. In the admin console, open **Tools & resources** (`/tools`).
2. Find `list_tickets` under `mcp-probe` (risk level `read`) and click **Edit**.
3. Change **Risk level** to **`write`** and click **Save**.
   (Moving a tool to `destructive` also asks you to tick *"I understand this changes authorization for all agents"*; `write` does not.)
4. The table shows `list_tickets` with the new `write` level.

### Step 3: The same call is now blocked
Back in Inspector, click **Execute Tool** on `list_tickets` again. You get a JSON-RPC error instead of tickets:

```json
{"jsonrpc":"2.0","id":3,"error":{"code":-40301,"data":{"execution_id":"exec-…","reason":"policy_deny"},"message":"AgentGate Authorization Denied: policy_deny (an explicit policy denied this request)"}}
```

### Step 4: Put it back
Edit `list_tickets` back to **`read`**, Save, and run it in Inspector once more. It works again.

---

## 4. What to Look For (Expected Result)
- **Inspector:** tickets, then an error, then tickets again, with no restart of anything in between.
- **Audit page** (`/audit`): the middle call is a **`DENY`** row for `mcp-probe/list_tickets`, risk level `write`, reason `policy_deny`. The calls before and after are `ALLOW` rows with risk level `read`.
- **Say:** "We did not restart the gateway, redeploy the backend or touch any code."

---

## 5. Clean-Up
Make sure `list_tickets` is back to `read`. If you are unsure of the state, run `.\deploy\prod\scripts\demo-mode.ps1 -Reset` (it restores every tool's bootstrap level and re-activates policy `v1.0.0`; it also restarts the traffic generator, so `docker stop prod-demo-agent` again if you are continuing a manual demo).

## 6. Fallback without the console
The same change through the Admin API (the field is `risk`; the request replaces the whole tool record, so it repeats the schema):
```powershell
$opsPass = (Get-Content deploy\prod\secrets\ops_cli_client_secret).Trim()
$tokenResp = curl.exe -k -s -X POST https://localhost:8443/realms/agentgate/protocol/openid-connect/token -d "grant_type=client_credentials&client_id=agentgate-ops-cli&client_secret=$opsPass"
$adminToken = ($tokenResp | ConvertFrom-Json).access_token
curl.exe -k -s -X PUT https://localhost:9443/api/v1/tools/mcp-probe/list_tickets -H "Authorization: Bearer $adminToken" -H "Content-Type: application/json" -d '{\"risk\":\"write\",\"schema\":{\"type\":\"object\",\"properties\":{\"status_filter\":{\"type\":\"string\"}}},\"arguments\":[{\"name\":\"status_filter\",\"type\":\"string\",\"required\":false}],\"description\":\"\"}'
```
Use the literal string exactly as shown; building the JSON from a PowerShell variable makes curl report a misleading "malformed request body". To undo, send the same command with `\"risk\":\"read\"`.
