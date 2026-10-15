# Scenario 5: Policy Lifecycle (Dry-Run & Instant Rollback)

## 1. What This Proves (What to Say)
> "Deploying a broken authorization policy in production can take down critical business workflows. AgentGate treats policies with the rigor of code: policies are versioned, candidates can be tested against historical audit logs using 'Dry-Run Simulation' to measure counterfactual impact, and if an issue occurs, operators can atomically roll back with a single click."

---

## 2. Pre-Conditions
- Admin Console is open at **`https://localhost:8444/policies`**.

---

## 3. Step-by-Step Execution

### Step 1: Inspect the Active Policy
1. Navigate to **`https://localhost:8444/policies`** in your browser.
2. Under the **Active Policy** card, notice:
   - Version: **`v1.0.0`**
   - Active status with a green badge.
3. Review the declarative Cedar policy syntax shown in the editor/viewer.

---

### Step 2: Create a Candidate Policy
1. Click **"Create candidate"** (`/policies/new`).
2. **You do not write Cedar by hand.** The page is a no-code rule builder: pick a role (reader/admin/payer/all), allow or deny, and tick which risk levels (read/write/destructive) it applies to. AgentGate generates the actual Cedar text from your choices and shows it read-only below the form — you can glance at it, but you build the rule with the controls, not by typing policy syntax.
3. Give it a description, e.g. `Stricter read-only lockdown for weekend maintenance`.
4. Find the **Admin** rule row and **uncheck "write"**, leaving only "read" checked. This proposes making admins read-only.
5. Click **Validate policy**, then **Save**. The candidate appears in the table as the next version (e.g. `v1.1.0`).

---

### Step 3: Run Dry-Run Simulation (The Safety Guard)
1. Next to your candidate policy, click **Dry-Run Analysis** (or **Run dry run** on its page).
2. **What this does, exactly — don't overstate it:** it replays the candidate against **5 fixed, representative sample requests** (a reader and an admin each trying a read, a write, and a destructive action) — **not** live historical traffic from the database. This is a deliberate, honest limitation worth saying out loud if asked: it tells you whether the *shape* of access changes as intended, not a statistical measure over real past requests.
3. **What the screen displays:**
   - **Total compared:** `5`
   - **Changed decisions:** how many of those 5 would flip from ALLOW to DENY. Expect the admin-write sample to flip — that's the lockdown working.
   - This gives the security engineer concrete, verifiable confidence *before* activating the change — just over 5 known cases, not a statistical sample.

---

### Step 4: Activate & Atomic Rollback
1. Click **Activate Policy**.
2. The UI confirms that `v1.1.0` is now the active policy serving live traffic.
3. To demonstrate disaster recovery, immediately click the **Rollback** button (or select `v1.0.0` and click Rollback).
4. The system issues an atomic database transaction restoring `v1.0.0` as the active engine baseline.

---

## 4. What to Point Out to the Audience
- **Audit continuity:** Both the activation and the rollback events are themselves recorded in the audit log.
- **Zero service restarts:** No Envoy configuration reloads, no downtime, and no drop in running gRPC connections.
