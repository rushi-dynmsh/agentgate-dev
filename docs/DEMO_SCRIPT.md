# AgentGate Management Demo Script

**Estimated Duration:** ~15 minutes (the optional automatic-traffic part adds ~2)
**Target Audience:** Engineering Leadership, Security Architects, Management
**Presenter Goal:** Show that AgentGate is a real-time checkpoint in the only path between AI agents and tools, using a real MCP client that you drive by hand.

How this demo works: **you** make each call from MCP Inspector (the official MCP testing tool) and each one is logged. You change the rules live from the admin console and call again, so the audience sees behaviour change. An automatic traffic client comes last, only to make the dashboard come alive.

---

## Pre-Flight Checklist (the night before, and again 5 minutes before)

All setup lives in one place: **[`docs/DEMO_START_TO_FINISH.md`](./DEMO_START_TO_FINISH.md)**. Do its Parts 1 to 3 (one-time setup, start everything, health check), then confirm:
- [ ] The health check passes (seven containers up, then `200`, `401`, `200`).
- [ ] You are logged in to the admin console (`https://localhost:8444`), and the documentation site is open (`http://localhost:3000`).
- [ ] You have run MCP Inspector once on this machine, so `npx` has downloaded it (it needs internet the first time).
- [ ] The fallback video opens: `website/static/media/walkthrough.webm`.
- [ ] **You have run this whole script once, start to finish, before the real thing.** That is the only way to catch anything specific to your machine.

---

## Demo Timeline & Steps

### Step 0: Context & The Problem (00:00 – 01:30)
- **Talking Point:** "AI agents now take real actions in production, but nothing checks what each individual tool call is asking for. AgentGate is the checkpoint in the only path between an agent and its tools."
- **Action:** Show the architecture diagram (`brand/diagrams/system-flow.svg`).
- **Expected Result:** Everyone sees that agents reach tools only through AgentGate.
- **Fallback:** The diagram on the live site at `http://localhost:3000/docs/concepts`.

---

### Step 1: The Landing Page & Attack Player (01:30 – 03:00)
- **Talking Point:** "First, how AgentGate treats real attacks."
- **Action:**
  1. On `http://localhost:3000`, show the animated request flow at the top: allowed requests pass, rogue ones are stopped.
  2. Scroll to **Watch attacks get stopped** and play two scenarios, for example **Unauthenticated Request** and **Unknown Tool Access**. Use **Next** to step through.
- **Expected Result:** Each scenario ends with a red **BLOCKED** verdict and the reason.
- **Fallback:** Step through manually with the **Next** button.

---

### Step 2: A Real Client, Allowed and Blocked (03:00 – 06:30)
Exact steps and inputs: [`docs/DEMO_START_TO_FINISH.md`](./DEMO_START_TO_FINISH.md) Part 4.0 to 4.2, and [`docs/demo-scenarios/01-real-client-inspector.md`](./demo-scenarios/01-real-client-inspector.md).
- **Talking Point:** "This is MCP Inspector, the tool the MCP community uses to test servers. To it, the gateway is just an MCP server. I'll make the calls myself."
- **Action:**
  1. Stop the background traffic: `docker stop prod-demo-agent`. Open the admin console's **Audit log** in a second tab.
  2. Run the token and Inspector command from Part 4.0 (the token lasts 5 minutes, so do this right before).
  3. In Inspector: **Tools**, run `list_tickets` (no input). Three tickets come back. Refresh the Audit tab: one **ALLOW** row.
  4. Run `delete_ticket` with `ticket_id` = `TCK-1001`. A JSON-RPC error comes back (`policy_deny`). Refresh the Audit tab: one **DENY** row.
- **Expected Result:** Tickets for the read, an error for the destructive call, and two matching audit rows (principal `agent-reader`).
- **Fallback:** Scenario 1, Step 6 gives the same calls as plain `curl`. If both fail, say so and go to Step 3; later steps need the console but not Inspector.

---

### Step 3: Change a Tool's Risk Level, Live (06:30 – 08:00)
Exact steps: [`docs/demo-scenarios/02-live-tool-reclassification.md`](./demo-scenarios/02-live-tool-reclassification.md).
- **Talking Point:** "A security team can change how risky a tool is, and the very next call is judged by the new rule, with no restart."
- **Action:**
  1. In the console, **Tools & resources**: edit `list_tickets`, set Risk level to `write`, Save.
  2. In Inspector, run `list_tickets` again: it is now **blocked** (a reader may only use `read` tools).
  3. Edit `list_tickets` back to `read`, Save, run it again: allowed.
- **Expected Result:** Tickets, error, tickets, with nothing restarted. The Audit page shows the middle call as a DENY with risk level `write`.
- **Fallback:** The scenario's Step 6 does the same through the Admin API.

---

### Step 4: Policy Lifecycle: Dry-Run, Activate & Rollback (08:00 – 11:00)
Exact steps: [`docs/demo-scenarios/03-policy-dryrun-and-rollback.md`](./demo-scenarios/03-policy-dryrun-and-rollback.md).
- **Talking Point:** "AgentGate treats policy like code: drafted, tested, activated with a reason, and instantly reversible."
- **Action:**
  1. **Policies**, **Create candidate**. Description `Demo: block reader reads`. On **Rule 1** (Reader) change **Decision** to **Deny**. **Validate policy**, **Save candidate**. You never type policy syntax; the generated Cedar is shown read-only.
  2. The dry run opens. **Say exactly what it does:** it replays the candidate against 5 fixed sample requests, not live history, and shows which would flip. The `agent-reader` read row flips from ALLOW to DENY.
  3. **Proceed to activation**, enter a reason, tick the confirmation, **Activate policy**.
  4. In Inspector, run `list_tickets`: **blocked**, because the new policy is live.
  5. **Policies**, `v1.0.0`, **Rollback to this**, **Rollback**. Run `list_tickets`: allowed again.
- **Expected Result:** The agent's behaviour changes the moment the policy does, and returns on rollback.
- **Fallback:** Explain the lifecycle from `http://localhost:3000/docs/policies`.

---

### Step 5: The Evidence: Audit Trail & Tamper Check (11:00 – 13:00)
- **Talking Point:** "Every decision, allowed or denied, is written to a cryptographically chained log. You've just watched every call we made appear in it."
- **Action:**
  1. **Audit log**: filter **Decision** to `DENY`. Expand a row: principal, tool, risk level, policy version, reason.
  2. Switch the filter back to `All` to show the full sequence: ALLOW, DENY, the risk change, the policy change.
  3. In a terminal: `docker exec prod-agentgate /bin/agentgate verify-audit -workspace default`
- **Expected Result:** Instant filtering, and `OK <N> records` (the number grows through the demo; its exact value doesn't matter).
- **Fallback:** Expand any row and show the **Audit Chain** panel.

---

### Step 6 (optional): The Automatic Client & Live Dashboard (13:00 – 15:00)
- **Talking Point:** "Real deployments are not one person clicking. Here is an automated agent sending a steady stream of mixed requests."
- **Action:**
  1. `docker start prod-demo-agent`
  2. Open the **Dashboard**: the allowed and denied counters and the per-hour chart fill up (it takes a minute); the Audit page streams new rows.
- **Expected Result:** Live counters, a visible deny rate, and top denied reasons (`policy_deny`, `unknown_tool`).
- **Fallback:** The dashboard and walkthrough video are on the landing page's **Production Admin Console** section.

---

### Step 7: Summary & What's Next (15:00)
- **Talking Point:**
  - "Zero-trust identity, per-call authorization, short-lived credentials, and a tamper-evident audit trail, working with a real MCP client."
  - "Current milestone: v1.0.0-rc feature complete."
  - "Next: connecting third-party MCP servers with per-backend credential settings, and registering a server by URL with automatic tool discovery."
- **Honest limits if asked:** the demo's tool server is our own; the tool catalog is entered by hand today; the tool list shows every tool, but forbidden ones are blocked when called.
- **Post-demo clean-up:**
  ```powershell
  .\deploy\prod\scripts\demo-mode.ps1 -Reset
  ```
