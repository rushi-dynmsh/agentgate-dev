# Scenario 3: Policy Lifecycle (Dry-Run, Activate & Instant Rollback)

## 1. What This Proves (What to Say)
> "A bad authorization rule can stop a business process. AgentGate treats policy like code: you draft a candidate, check what it would change with a dry run, activate it with a recorded reason, and if anything looks wrong you roll back in one click. And we'll see the agent's behaviour change the moment the policy changes."

---

## 2. Pre-Conditions
- The stack is running, and you did the preparation in `docs/DEMO_START_TO_FINISH.md` Part 4.0: traffic generator stopped, Inspector connected as the **reader** agent, the admin console open at `https://localhost:8444/policies`.
- `list_tickets` is classified `read` (it is, unless you left Scenario 2 half-done).

---

## 3. Step-by-Step Execution

### Step 1: See today's behaviour
In Inspector, run `list_tickets`: it works. In the console, **Policies** shows the active policy `v1.0.0` (green **active**).

### Step 2: Draft a candidate (no Cedar typing)
1. Click **Create candidate**.
2. The page is a no-code rule builder. **Description:** `Demo: block reader reads`.
3. **Rule 1** reads "Allow reader to use read tools". Change its **Decision** from Allow to **Deny**. Leave Role = Reader and Tool risk = `read`. (Rule 2, admin, stays as is.)
4. Below the form, the **Generated Cedar** preview updates, read-only. You never type policy syntax.
5. Click **Validate policy**, then **Save candidate**.

### Step 3: Dry run (what would change?)
Saving opens the dry-run page. It replays the candidate against **5 fixed sample requests** (a reader and an admin each trying a read, a write and a destructive action). It is **not** live historical traffic: it shows whether the *shape* of access changes as intended. Say that out loud if asked.

Read the table: the `agent-reader` read row goes from **ALLOW** to **DENY** (**Changed**). The admin rows show **No change**. (A reader's destructive row can also be flagged, because the reason for its denial changes while the answer stays DENY.)

### Step 4: Activate, and watch the agent change
1. Click **Proceed to activation**.
2. Enter a **Reason for activation** (for example `demo: block reader reads`), tick *"I confirm I have reviewed the dry run results…"*, and click **Activate policy**.
3. In Inspector, run `list_tickets` again. It is now **blocked** with `policy_deny`.

### Step 5: Roll back
1. On **Policies**, find the row for `v1.0.0` and click **Rollback to this**, then **Rollback**.
2. Run `list_tickets` in Inspector once more: **allowed** again.

---

## 4. What to Look For (Expected Result)

| Moment | Inspector | Audit page |
|---|---|---|
| Before | tickets | `ALLOW` |
| After activation | `policy_deny` error | `DENY` |
| After rollback | tickets | `ALLOW` |

Point out: no restart, no redeploy, no dropped connections. Each activation and rollback is itself recorded, and the audit rows show which policy version made each decision.

Note: a new candidate's version is a long hash, not `v1.1.0`. Read the version from the row you just created.

---

## 5. Clean-Up
Confirm `v1.0.0` is the active policy. The rejected candidate stays in the list as history. `demo-mode.ps1 -Reset` re-activates `v1.0.0`; `-Down -Volumes` removes everything.
