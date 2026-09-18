# AgentGate v1 — 3-Team Parallel Execution Plan

**Purpose:** Replace the 5-workstream, tightly-gated execution model with 3 strictly independent
teams now that the critical-path enforcement work (G1–G6) is closed, and bring an
out-of-band-developed Admin UI into the same governance the rest of the codebase follows.

**Planning basis:** Supersedes
[`AGENTGATE_V1_10_DAY_PARALLEL_TEAM_EXECUTION_PLAN.md`](AGENTGATE_V1_10_DAY_PARALLEL_TEAM_EXECUTION_PLAN.md)
(see the banner added to that file). That plan's 5-workstream model, checkpoint-gate discipline,
and non-negotiable security invariants are preserved in spirit; only the team shape, the
independence model between teams, and the current milestone change.

**Why now:**
1. G1–G6 are PASS/CLOSED/FROZEN (`docs/DEVELOPMENT/CURRENT_STATUS.md`). The tight critical path
   that justified 5 synchronized workstreams (`Go Backend → Gateway/MCP → Frontend → QA →
   DevOps`, gated at every step) is done — real MCP traffic is enforced end-to-end, durably
   audited, and policy-governed.
2. A teammate independently built a substantial Admin UI (`admin-ui/`, merged 2026-09-18) against
   a G1–G4 understanding of the backend. Backend has since shipped G5 (durable audit) and G6 (real
   enforcement); the UI's own `FLOW_AND_ARCHITECTURE.md` documents mock-vs-real status as of G4 and
   is now stale in its reasoning (see §3).
3. QA/Security and DevOps as *separate* workstreams made sense while every stream had daily
   critical-path dependencies on the others. They no longer do — verification and deployment work
   now belongs inside whichever team owns the change (a backend change carries its own tests and
   deploy config; there is no longer a 10-day calendar forcing a dedicated stream per concern).
4. The product direction has grown two new, explicit goals beyond the original enforcement scope:
   a realistic demonstration deployment (to prove AgentGate's value in something more convincing
   than a toy `probe-mcp` counter), and — later, explicitly deferred — a public documentation site
   and landing page.

---

## 1. Shared Architecture and Non-Negotiable Boundaries (carried forward, updated)

The production path, now proven real end-to-end (G6):

```text
MCP Client / AI Agent
        |
        v
agentgateway (v1.4.0, pinned)
  - MCP transport/routing, JWT validation (strict), tool discovery
  - Envoy v3 ext_authz gRPC callout to AgentGate:9001
        |
        v
AgentGate (cmd/agentgate)
  - trusted identity (gateway-verified JWT claims only, never client headers)
  - tool/argument governance (internal/toolregistry, internal/argdecl)
  - Cedar authorization (internal/decision, internal/policy)
  - fail-closed enforcement + audit-before-ALLOW (internal/audit)
  - policy versioning/lifecycle (internal/policystore, internal/policymanager, internal/govapi)
        |
        v
Real MCP Backend
        |
        v
Downstream credential boundary — NOT YET BUILT (O-001, this plan's G7)
```

Ownership table, collapsed from 5 rows to 3 (QA/Security and DevOps concerns move into whichever
team owns the change, per §4):

| Team | Owns | Must not own |
|---|---|---|
| **Backend** | `agentgate/` Go module: identity, tool/argument governance, Cedar decision core, policy lifecycle, durable audit, downstream credentials, its own tests (unit/race/integration/security) and deploy config for what it ships | MCP protocol/proxy implementation, UI |
| **AI / Gateway** | `gateway/`, `deploy/`, agentgateway configuration, MCP transport correctness, the realistic demonstration system, integration-point discovery, its own E2E/black-box proof suites | Cedar policy content, durable audit schema, UI |
| **Frontend** | `admin-ui/`, `frontend/` (framework-agnostic contract layer), operator-facing workflow correctness, eventual docs/landing sites (deferred, §7) | Direct authorization enforcement, backend contract changes without Backend sign-off |

The non-negotiable invariants from `docs/SECURITY/PRODUCTION-INVARIANTS.md` are unchanged and
binding on all 3 teams without exception:

- Missing/invalid identity → DENY. Unknown/unclassified tool → DENY. Missing policy → DENY. Policy
  evaluation error → DENY. Malformed input cannot produce ALLOW.
- Every decision carries its exact policy version/hash and `execution_id`.
- Denied calls never reach the MCP backend (proven, G6).
- No ALLOW is returned unless its audit record is durably persisted first (proven, G5/G6).
- The gateway cannot bypass AgentGate for governed traffic; identity is trusted only from the
  gateway's cryptographic JWT verification, never from client-supplied headers (proven, G6).
- Downstream credentials must not be the raw inbound bearer token, once built (O-001, this plan's G7).
- Policy activation is atomic and rollbackable (proven, G3/G4). Policy mutation requires
  authenticated administration (proven, G3).
- No critical/high unresolved security defect may ship.

---

## 2. Current State (read before starting any team's work)

Do not re-derive this — read it:

- `docs/DEVELOPMENT/CURRENT_STATUS.md` — what exists and runs today, package by package.
- `docs/DECISIONS/OPEN_DECISIONS.md` — O-001 (downstream identity) and O-004 (MCP revision
  support) are open; everything else through G6 is resolved.
- `docs/PHASES/G6_WORKSTREAMS/CLOSURE_SUMMARY.md` — the most recent checkpoint's full walkthrough,
  including the exact diagram this plan's G7 closes the dotted line on.
- `admin-ui/README.md` and `admin-ui/FLOW_AND_ARCHITECTURE.md` — the UI's own honest
  real-vs-mock accounting, written against G4. §3 below states exactly what's changed since.

---

## 3. Admin UI ↔ Backend Consistency — what's actually stale, and what to do about it

The UI's self-assessment (`admin-ui/FLOW_AND_ARCHITECTURE.md` §3) is unusually rigorous — it
already labels every screen 🟢 real / 🟡 mock-grounded / 🔴 reframed and explains why. It is not
"vibecoded slop" that needs a rewrite. It needs a **targeted update**, because two of its stated
reasons for "mock" have changed:

| Screen | UI's stated reason (as of G4) | Actual state now (G6) | Action |
|---|---|---|---|
| Audit / Logs | "Durable, queryable audit storage is Gate G5 — not started" | G5 shipped durable, tamper-evident audit (`internal/audit`, Postgres `audit_events`). **But there is still no REST read endpoint for it** — `internal/govapi` exposes only policy CRUD/lifecycle routes (verified: `grep HandleFunc agentgate/internal/govapi/handler.go` — 8 routes, all under `/policies`). | Backend team builds a read-only audit query API (§5, G8). Until it exists, the UI's mock is still the *correct* choice — but its reasoning must say "no read API yet," not "audit doesn't exist yet." |
| Tools & Resources | "No REST API exposes `internal/toolregistry` yet" | Still true — `internal/toolregistry` has no HTTP surface. | Backend team builds a read-only tool inventory/classification API (§5, G8). Same "still correct conclusion, update the reasoning" note. |
| Decision Tester | Posts to `cmd/g1-mock-authz` (a non-production binary wrapping the same `internal/decision.Engine`) | `cmd/agentgate` now enforces for real, but only via gRPC `ext_authz` on `:9001` — there is no HTTP JSON decision-test surface on the real production binary. `g1-mock-authz` also bypasses the tool-governance/argument-whitelist/audit layers G2–G6 added around the core engine. | Document this gap explicitly in the UI (it already should not claim more than it does); Backend team decides in G8 whether a lightweight HTTP decision-test endpoint on `cmd/agentgate` is worth adding for demo/ops purposes, separate from the real gRPC enforcement path. |
| Everything else (Policies, Dashboard active-policy card, Settings, Login, Error states) | — | Unaffected — these were already real or correctly reframed and nothing about their backing changed. | No action. |

**Immediate task (Frontend team, before any new feature work):** update `FLOW_AND_ARCHITECTURE.md`
§3 rows 5 and 8/9 and the README's status table to say *why* each is still mock in **current**
terms (no read API), not G4-era terms (feature doesn't exist). This is a documentation-accuracy
fix, not a UI rebuild — the actual screens don't need to change until the backend read APIs exist.

**No other consistency defects were found.** The admin-ui's mock data deliberately lives only
under `admin-ui/src/mock/`, never mixed into the frozen `frontend/src/models` contract layer that
other workstreams read from — confirmed by direct inspection of `admin-ui/src/lib/contract.ts`,
which re-exports (not reimplements) every type/client from `frontend/src`. The merge that brought
`admin-ui/` into `development` also merged current `development` into the PR branch first, so
`frontend/src` as consumed by `admin-ui` is already at the G6 tip — there is no version-skew bug to
fix, only the documentation-accuracy gap above.

---

## 4. The Independence Model (what changes from v1)

The old plan's 5 workstreams were gated: every stream waited for the same numbered checkpoint
(G1...G9) at the same time. That discipline was correct for building one shared enforcement path
where every piece had to agree on one contract simultaneously.

Now:

- **Checkpoints keep the same shared numbering** (G7, G8, G9, ...) so the project's history stays
  one continuous ledger — `docs/README.md`'s checkpoint-folder convention
  (`docs/PHASES/G{N}_WORKSTREAMS/`) is unchanged.
- **A checkpoint no longer requires all 3 teams to participate or close together.** Each
  checkpoint's `00_G{N}_CHECKPOINT_REFERENCE.md` states which teams are active in it. A team not
  listed simply continues its own prior-checkpoint work or starts the next one — it does not wait.
- **Each team's workstream file within a checkpoint closes on its own schedule**, with its own
  report/digest in that checkpoint's `results/`, and its own PASS/CLOSED verdict from the Lead
  Architect. One team's checkpoint N+1 can start before another team's checkpoint N has closed.
- **Only two genuine cross-team dependencies exist right now** (everything else is parallelizable):
  1. Frontend's Audit/Tools screens cannot go from 🟡 to 🟢 until Backend ships the read APIs
     (§5, G8). Frontend does not block on this — it keeps building other screens and defers only
     those two.
  2. AI/Gateway's realistic demonstration system (§6) cannot exercise a *real* downstream credential
     until Backend ships G7's token-exchange mechanism. AI/Gateway does not block on this either —
     it builds the demonstration topology and picks the real backend first, using a stub credential
     the same way Gateway/MCP used a mock AgentGate during G1.
- **Use the same synchronization protocol as before** (`§6` of the superseded plan): each team
  reports DONE / CONTRACT / BLOCKED / RISK / NEXT. A team is blocked only when its next task
  genuinely needs a contract the other team hasn't frozen yet — not merely because the other team
  hasn't finished implementing.

---

## 5. Team A — Backend (`agentgate/`)

**Owns:** identity, tool/argument governance, Cedar decision core, policy lifecycle, durable
audit, downstream credentials, and all of `agentgate/`'s own tests and deploy config.

### G7 — Downstream Scoped Identity & Token Exchange (already named as "next" in `CURRENT_STATUS.md`)

Resolves **O-001**. Design and implement the production mechanism by which AgentGate (or the
component it delegates to) obtains a downstream credential for the real MCP backend that:

- Represents the correct caller/on-behalf-of identity, audience, and lifetime.
- Is never the raw inbound bearer token (no passthrough).
- Fails closed on credential-issuance failure (does not silently proceed uncredentialed).
- Keeps the caller/on-behalf-of identity auditable end-to-end (ties into `internal/audit`'s
  existing `execution_id`/provenance fields — no new audit schema should be needed, only a new
  populated field for the downstream credential reference/audience).

**Coordinate with AI/Gateway team:** this mechanism should target whatever real backend AI/Gateway
selects for the demonstration system (§6) — a scoped GitHub token, a scoped API key, or an
OAuth token-exchange flow, depending on what that backend actually requires. Do not design this in
a vacuum against a hypothetical backend; use AI/Gateway's G7 backend selection (their first task,
§6) as the concrete target before finalizing the mechanism.

**DoD:**
1. Correct downstream audience, scoped/short-lived credential where the backend supports it.
2. No raw inbound bearer-token passthrough — verified by an automated test asserting the token
   the backend receives differs from the token the client sent.
3. Credential-issuance failure denies the call rather than proceeding without a credential.
4. Caller/on-behalf-of identity remains present in the audit record for the call.
5. Replay/confusion risks addressed (e.g. credential scoped to this specific call/execution_id
   where the backend's credential model supports it).

### G8 — Governance Read-Surface Expansion

Unblocks Frontend's §3 remediation. Build two new read-only, authenticated REST endpoints in
`internal/govapi` (same auth pattern as the existing policy routes — constant-time admin-token
comparison):

1. **Audit query API** — `GET /api/v1/workspaces/{workspace_id}/audit-events` with pagination and
   basic filtering (by decision, by tool, by time range). Reads from the existing `internal/audit`
   Postgres store; no new write path, no new audit fields — this is read-only exposure of data
   that already durably exists.
2. **Tool registry read API** — `GET /api/v1/workspaces/{workspace_id}/tools` listing tools known
   to `internal/toolregistry` with their current classification/fingerprint status. Classification
   *mutation* (if the UI is to support operator classification of unclassified tools) is a
   separate, explicit decision — do not silently add a write path here; if the Frontend team wants
   one, that's a new ticket with its own authz review, not bundled into this read API.

**DoD:**
1. Both endpoints follow the existing `govapi` auth/error-model conventions exactly (no new
   auth mechanism).
2. Both are read-only — no new mutation surface introduced under this ticket.
3. Contract is documented and typed (mirroring how the G1 contract and G3 governance API were
   documented) before Frontend starts consuming it, so this doesn't repeat the "built ahead of an
   agreed contract" pattern that produced the §3 gap.
4. Frontend team reviews and signs off on the response shape before Backend freezes it.

### Ongoing backend hardening backlog (from the superseded plan's deferred Days 8–10 list, still valid, not yet scheduled to a specific G-number)

Carry forward, unscheduled until a team has bandwidth or a concrete need surfaces one:
TLS/mTLS configuration, request size/time limits, rate limiting, OpenTelemetry tracing, metrics,
concurrency correctness under load, and **O-004** (supported MCP revision matrix — currently
single-pinned to `2026-07-28`/`tools/call`).

---

## 6. Team B — AI / Gateway (`gateway/`, `deploy/`, MCP, the demonstration system)

**Owns:** agentgateway configuration and MCP transport correctness (unchanged from v1's WS-B), plus
two new mandates that didn't exist in the original 10-day plan.

### G7 — Realistic Demonstration System (new mandate)

The current proof (`deploy/g6/`) uses `probe-mcp`, a toy backend that only counts calls. That was
correct for proving the enforcement boundary in isolation, but it does not demonstrate AgentGate's
value to anyone evaluating the product. Build a second, additive deployment
(`deploy/demo/`, does not replace or modify `deploy/g6/`) wired to a **real or realistic** MCP
backend — pick one concrete option and justify it before building:

- A well-known open-source MCP server (filesystem, GitHub, or similar) run against real or
  realistic data, or
- A small custom MCP server exposing 2–3 tools with actual business meaning (e.g. a mock
  ticketing/CRM/database tool) with deliberately mixed risk classifications (read/write/destructive)
  so the governance story (deny-by-default, policy activation, rollback) is visibly meaningful
  rather than abstract.

**First task, and Backend's G7 dependency:** decide and document which backend and which
credential model it needs (API key? OAuth token? scoped PAT?) — hand this to Backend team before
they finalize the token-exchange mechanism (§5). Use a stub/test credential in the meantime so this
work is not blocked on Backend's G7 completing first.

**DoD:**
1. A real (or realistic, non-toy) MCP backend is reachable through the same
   `agentgateway → AgentGate → backend` path proven in G6, in an additive `deploy/demo/` topology.
2. At least one tool is classified `destructive` or `write` and demonstrably denied/allowed based
   on an actual policy change (candidate → validate → activate → observe), not just a fixture.
3. A short runbook/README explains the scenario in plain language — this is a demonstration
   artifact, its documentation IS part of the deliverable.
4. Once Backend's G7 credential mechanism exists, this system is updated to use it for real
   (replacing the interim stub) — tracked as a follow-up task in this same checkpoint, not a new one.

### G9 — Integration Point Expansion (new mandate)

Once the demonstration system works end-to-end, use it to identify and document additional places
AgentGate could plausibly integrate — e.g. multiple concurrent MCP backends behind one gateway,
a second agent framework, or a second identity provider pattern. This is explicitly a
**discovery and documentation task first** — write findings to
`docs/DECISIONS/OPEN_DECISIONS.md` or a new `docs/PHASES/G9_WORKSTREAMS/INTEGRATION_POINTS.md` as
candidate future scope, not silent architecture expansion. Per `CLAUDE.md`: identify assumptions
and report them; do not silently build speculative integrations without recording the decision.

---

## 7. Team C — Frontend (`admin-ui/`, `frontend/`, docs/landing sites later)

**Owns:** the admin UI, the framework-agnostic contract layer, and — later — public-facing
documentation and landing sites.

### Immediate (before new feature work): Consistency Remediation

Execute §3's table exactly: update `admin-ui/FLOW_AND_ARCHITECTURE.md` §3 rows 5 (Audit) and 8/9
(Tools) and the README status table to state the *current* reason each is still mock (no read API
yet), not the *G4-era* reason (feature not built). This is a documentation edit, not a UI rebuild.

### G8 (paired with Backend's G8): Wire the new read APIs

Once Backend ships the audit-query and tool-registry read endpoints:

1. Add typed models/client methods to `frontend/src/models` and `frontend/src/api` for both
   endpoints (matching the existing pattern `GovernanceClient` already establishes — a documented
   interface plus `Http*` and `Mock*` implementations, so `admin-ui` can toggle mock/live the same
   way Policies already does).
2. Wire `admin-ui`'s Audit Logs and Tools & Resources pages to the real client, keeping the mock
   client as the default/fallback exactly as Policies does today.
3. Flip the 🟡 status to 🟢 in `FLOW_AND_ARCHITECTURE.md` only once this is actually live and
   tested against the real backend — not preemptively.

### Ongoing: continued admin-ui development

Beyond the consistency fix, continue building out the UI per its own existing, already-good
architecture doc (dashboard aggregation, dry-run-over-real-audit-history once G8's audit API
exists, tool classification write-path if and when Backend agrees to build one, etc.) — this is
normal incremental product work, not a special checkpoint; scope individual tickets as needed
using the `new-checkpoint-scaffold`/task-ticket convention when a body of work is large enough to
warrant one.

### Deferred (explicitly not scheduled): Documentation site + landing page

Later, **not now**, this team will own:
1. A public-facing product documentation site built with Docusaurus, following the same approach
   already proven in `developer-docs/` (which is the internal/engineering-facing site — the new
   one is user/product-facing and does not replace it).
2. A product landing page.

Both are explicitly out of scope until the product itself is far enough along to document
publicly. The repo is a monorepo today and neither site will ship as part of the final user-facing
product offering as currently structured — a packaging/repo-structure decision for whenever this
work actually starts, not now. Do not begin either until directed.

---

## 8. What stays out of scope (unchanged from the superseded plan, plus one addition)

Real-time per-call HITL, mandatory SpiceDB/resource ownership, ML risk scoring, a custom MCP proxy,
full multi-tenant runtime, advanced compliance exports — all still explicitly deferred. **Added by
this plan:** the public docs site and landing page (§7) are deferred the same way — recorded so
they aren't lost, not built until explicitly started.

---

## 9. Immediate Next Steps (in order)

1. ~~Write G6's missing `CLOSURE_SUMMARY.md` to its canonical location~~ — done as part of adopting
   this plan (`docs/PHASES/G6_WORKSTREAMS/CLOSURE_SUMMARY.md`).
2. Frontend team executes the §3 documentation-accuracy fix (small, unblocks nothing else, but
   should not be left inaccurate any longer than necessary).
3. Scaffold `docs/PHASES/G7_WORKSTREAMS/` (per `docs/README.md`'s checkpoint-folder convention)
   with two active workstream files — `01_BACKEND_G7.md` (§5) and `02_AI_GATEWAY_G7.md` (§6).
   Frontend is not active in G7 (its next scheduled work is G8) and is correctly absent from that
   checkpoint's reference doc, per §4's independence model.
4. AI/Gateway team's first concrete action inside G7: pick and document the real/realistic backend
   (§6), then hand its credential requirements to Backend team.
5. Backend team's first concrete action inside G7: design the token-exchange mechanism against
   that concrete backend (§5).
