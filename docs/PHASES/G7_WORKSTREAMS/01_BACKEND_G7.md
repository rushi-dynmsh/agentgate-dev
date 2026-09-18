# G7 Ticket — Backend Team (`agentgate/`)

**Status:** OPEN — ready to start now.
**Team:** Backend. **Plan:** `docs/PHASES/AGENTGATE_V1_3_TEAM_PARALLEL_EXECUTION_PLAN.md` §5.
**What comes after this checkpoint:** `docs/PHASES/PROGRESS_AND_ROADMAP.md` — the checkpoint tracker (G1→G11), with your team's next tasks already scoped and their dependencies marked.
**This file is self-contained** — you do not need to read the other two teams' tickets to start
Task A. Read the "Collaboration & Blockers" section (§4) before starting Task B.

---

## 0. Read first (5 minutes, don't skip)

- `docs/DEVELOPMENT/CURRENT_STATUS.md` — what exists today, package by package.
- `docs/SECURITY/PRODUCTION-INVARIANTS.md` — binding; nothing in this ticket may contradict it.
- `docs/DECISIONS/OPEN_DECISIONS.md` O-001 (downstream identity) and O-004 (MCP revisions) — open,
  relevant to Task B; do not silently resolve either.

## 1. Non-negotiables (apply to every task below, no exceptions)

- Missing/invalid identity → DENY. Unknown/unclassified tool → DENY. Malformed input → DENY.
  Policy evaluation error → DENY.
- No ALLOW may be returned unless its audit record is durably persisted first (the existing
  `audit.AuditedDecisionService` invariant — do not build a second path around it).
- **New as of this checkpoint:** a test that silently substitutes an easier check when its real
  precondition isn't met, and still reports success, is a bug — not acceptable in new or modified
  test code. See Task A.

## 2. Task A — G6 Evidence Closeout (do this first, before Task B or C)

**Why:** an independent verification pass (2026-09-17/18, cross-checked directly against this
repo's code) found that G6's *implementation* is real and correctly wired, but its *end-to-end
enforcement evidence* is currently overstated. This is corrective closeout on an
already-approved checkpoint, per `/WORKFLOW.md` §2 — **do not reopen G6's scope or its
`CLOSURE_SUMMARY.md`.** Fix the evidence, record an addendum, move on.

**Exact bugs, verified by reading the code directly (file: `agentgate/qa/g6enforcement/enforcement_test.go`):**

1. **11 of the 12 test functions silently skip their live-path assertions.** Every scenario except
   `TestScenario07` and `TestScenario12` is structured as:
   ```go
   if isLiveGatewayAvailable() {
       // ...real HTTP/gRPC assertions against a live gateway...
   }
   // ...then unconditionally falls through to an in-process harness call...
   h := setupInProcessHarness(t)
   ```
   `isLiveGatewayAvailable()` returns `false` whenever `getBackendURL() + "/healthz"` isn't
   reachable — which is **always true** in a plain `go test ./...` and in CI (`.github/workflows/ci.yml`
   runs `go test -race ./...` with no docker-compose step at all). The in-process fallback is a
   legitimate fast unit check of the adapter/server logic, but it is not the same claim as
   "E2E-enforced against the live topology," and the test file currently makes no distinction
   between the two — anyone running the suite gets 12/12 green with zero live-path coverage and no
   indication of that.

2. **`TestScenario07_AgentGateUnavailable` (line ~432) is a tautology.** Its body:
   ```go
   var nilServer *authz.Server
   if nilServer != nil {
       t.Fatal("expected nil server")
   }
   ```
   `nilServer` is declared and never assigned anything but `nil` — this condition is always false.
   The rest of the test does an unrelated raw-socket connection-refused check against
   `http://127.0.0.1:9099/unreachable`, which proves nothing about the gateway's actual fail-closed
   behavior when AgentGate itself goes down.

3. **`TestScenario12_OversizedBody` (line ~599) cannot fail on the one outcome it should catch.**
   ```go
   if allowed && h.backendCalls > 0 {
       t.Log("Oversized payload rejected or sanitized")   // <-- t.Log, not t.Fatal/t.Error
   }
   if !allowed && code != int32(codes.PermissionDenied) {
       t.Fatalf(...)   // only fires if already denied with the wrong code
   }
   ```
   If an oversized payload is incorrectly **allowed**, this test only logs a message and passes.
   It provides zero regression protection against the one failure mode it's named for.

**What to do:**

- Rewrite `TestScenario07_AgentGateUnavailable` to actually exercise `authz.Server`/the adapter
  with a genuinely unavailable dependency (e.g. a `DecisionService` that returns a connection
  error) and assert the call is denied — not a tautological check on an unused nil variable.
- Rewrite `TestScenario12_OversizedBody` so an unexpected `allowed == true` on an oversized payload
  is a `t.Fatalf`, not a `t.Log`.
- Restructure the file so there are two distinctly named test surfaces: a fast **unit** suite
  (`TestUnit_*` or an existing naming convention you introduce and document) that always runs
  in-process, and a **live E2E** suite that `t.Skip()`s loudly with a clear message
  ("SKIPPED: no live gateway reachable at $URL — set env var X to run this suite") when
  `isLiveGatewayAvailable()` is false, rather than silently substituting the unit path. The 12
  DoD scenarios' *live* claim must come only from the live suite actually running, not from the
  unit fallback masquerading as it.
- Coordinate with AI/Gateway team (see §4) to get a live topology to actually run the new live
  suite against at least once, and capture that run's output as this checkpoint's real evidence,
  replacing the one-time manual `G6_GATEWAY_CONTRACT_OBSERVED.json` capture.

**DoD for Task A:**
1. `TestScenario07` and `TestScenario12` (or their renamed equivalents) can each genuinely fail —
   verify this yourself by temporarily breaking the behavior they test and confirming the test
   catches it, then revert.
2. The live-path and unit-path test surfaces are structurally separate and distinctly named; the
   live surface skips loudly (not silently) when no live gateway is present.
3. The live surface has been run at least once against a real topology (with AI/Gateway's help,
   §4) and that run's output is committed as this checkpoint's evidence.
4. A corrective-closeout addendum is added to `docs/PHASES/G6_WORKSTREAMS/CLOSURE_SUMMARY.md` (a
   new dated section, not an edit to the existing PASS verdict) stating what was found and fixed.
5. `go build ./...`, `go vet ./...`, `gofmt -l .` clean; `go test ./...` still green for every
   other package — this fix must not break any currently-passing test.

## 3. Task B — Downstream Scoped Identity & Token Exchange (resolves O-001)

**Do not start the implementation half of this until AI/Gateway's Task B (their ticket §2) has
told you which real backend it targets and what credential model that backend needs — see §4.**
You may and should start the *design* half immediately.

Design and implement the production mechanism by which AgentGate obtains a downstream credential
for the real MCP backend that:

- Represents the correct caller/on-behalf-of identity, audience, and lifetime.
- Is never the raw inbound bearer token (no passthrough) — a client's token must never be the same
  bytes that reach the backend.
- Fails closed on credential-issuance failure — the call is denied, not sent uncredentialed.
- Keeps the caller/on-behalf-of identity auditable: extend `internal/audit`'s existing
  `DecisionRecord`/`StoredRecord` types (see `agentgate/internal/audit/types.go`) with a new field
  for the downstream credential reference/audience if one doesn't already fit — do not invent a
  second, parallel audit path.

**Suggested boundary:** a new package, e.g. `agentgate/internal/credential`, invoked only after an
`ALLOW` decision from `audit.AuditedDecisionService.Evaluate()`, never before. Keep it decoupled
from both Cedar (`internal/decision`/`internal/policy`) and the Envoy adapter (`internal/authz`) —
those packages should call into this one, not the reverse.

**DoD:**
1. Correct downstream audience; scoped/short-lived credential where the target backend supports it.
2. Automated test proving the token the backend receives differs from the token the client sent.
3. Credential-issuance failure denies the call (test this explicitly — simulate the issuance
   failing and assert no backend call happens).
4. Caller/on-behalf-of identity is present in the resulting audit record.
5. Replay/confusion risk addressed appropriately for the chosen mechanism (e.g. credential scoped
   to the specific `execution_id` where the backend's model supports it).

## 4. Collaboration & Blockers — read before starting Task B or C

| Your task | Depends on | From whom | What happens if you skip ahead anyway |
|---|---|---|---|
| Task A (fix tests) | Nothing — start immediately | — | — |
| Task A's "run the live suite once" step | AI/Gateway's Task A (`02_AI_GATEWAY_G7.md` §2) standing up a reachable `deploy/g6` topology | AI/Gateway team | You can finish the code-side fix without this, but Task A's DoD item 3 (a real run) stays open until they deliver it — say so honestly in your report rather than claiming it's done. |
| Task B implementation (not design) | AI/Gateway's Task B backend/credential-model choice (`02_AI_GATEWAY_G7.md` §3) | AI/Gateway team | Building a credential mechanism against a hypothetical backend risks having to redo it once the real backend's actual requirements (API key vs OAuth vs scoped PAT) are known. Do the design/interface work now; hold the concrete implementation for their input. |
| Task C's response-shape freeze | Frontend's proposed contract draft (`03_FRONTEND_G7.md` §2) | Frontend team | You can implement Task C's endpoints against your own best guess, but do not mark the contract "frozen" until Frontend has reviewed it — repeating the exact "built ahead of an agreed contract" mistake that produced the admin-ui consistency gap this checkpoint is partly cleaning up. |

**If you get blocked for real** (not just "waiting is mildly inconvenient"): report it per
`/WORKFLOW.md`'s DONE/CONTRACT/BLOCKED/RISK/NEXT format, and keep working on whichever of Task A/B/C
isn't blocked. Do not sit idle waiting for another team.

## 5. Task C — Governance Read-Surface Expansion (unblocks Frontend's Audit/Tools panels)

Build two new read-only, authenticated REST endpoints in `internal/govapi`, following the exact
existing pattern (`AdminAuthMiddleware` from `agentgate/internal/govapi/middleware.go`, same
`ErrorResponse{Error: ErrorDetail{Code, Message}}` error shape used by the 8 existing policy
routes in `agentgate/internal/govapi/handler.go`).

1. **Audit query API** — `GET /api/v1/workspaces/{workspace_id}/audit-events`
   - Query params: `limit` (default/max reasonable, e.g. 100), `cursor` or `before_sequence` for
     pagination (`internal/audit` already has `GetRecordBySequence` — use it for cursor resolution).
   - Wraps the **existing** `Store.ListRecords(ctx, workspaceID, limit)` method (both
     `MemoryStore` and `PostgresStore` already implement this in
     `agentgate/internal/audit/memory.go` / `postgres.go`) — this is genuinely a thin read wrapper,
     not a new subsystem.
   - Response: JSON array of `StoredRecord` (see `agentgate/internal/audit/types.go` for the exact
     fields already there: `id`, `workspace_id`, `sequence_number`, `execution_id`, `timestamp`,
     `event_type`, `decision`, `reason`, `principal_agent_id`, `principal_roles`,
     `principal_on_behalf_of`, `tool_backend_id`, `tool_name`, `tool_risk`, `policy_version`,
     `policy_hash`, `redacted_arguments`, `canonical_payload`, `prev_hash`, `row_hash`) — reuse
     these field names and JSON tags as-is; do not invent a parallel shape.
   - **No new write path. No new redaction rules** — `redacted_arguments` is already redacted at
     write time by `internal/audit/redact.go`; this endpoint must not expose `canonical_payload` in
     a way that defeats that (review whether `canonical_payload` should be excluded from the HTTP
     response entirely — it's an internal chain-verification artifact, not operator-facing data;
     default to excluding it unless there's a real reason to expose it).

2. **Tool registry read API** — `GET /api/v1/workspaces/{workspace_id}/tools`
   - `internal/toolregistry.Registry` (see `agentgate/internal/toolregistry/toolregistry.go`)
     **currently only supports `Lookup` by exact `ToolID` — there is no enumeration method.** You
     must add one (e.g. `func (r *Registry) List() []GovernanceRecord`) before this endpoint can
     exist. This is a small, additive change to the existing type — do not restructure `Registry`.
   - The registry is currently a **static, startup-loaded fixture list**
     (`agentgate/cmd/agentgate/main.go` builds it from a hardcoded `[]RegistryEntry{...}` literal).
     This endpoint lists *configured* tools only — it does **not** discover unclassified tools seen
     in live traffic (that data doesn't exist anywhere yet). State this limitation explicitly in
     the endpoint's response or documentation; do not let the UI imply more than this delivers.
   - Response: JSON array reflecting `GovernanceRecord`'s fields (`ToolID` — itself
     `BackendID`/`ToolName`, `Known`, `Risk`, `RegisteredFingerprint`) — see
     `agentgate/internal/toolregistry/toolregistry.go`.
   - **Explicit non-goal for this ticket:** no tool-classification *write* path (an operator
     changing a tool's risk from the UI). That's a separate decision with its own authorization
     review — if Frontend wants one, it's a new ticket, not bundled here.

**DoD for Task C:**
1. Both endpoints use `AdminAuthMiddleware` exactly as the existing routes do — no new auth
   mechanism invented.
2. Both are read-only; `go vet`/code review confirms no mutation path was added.
3. Response shapes are documented (a short markdown contract doc, same style as
   `docs/PHASES/G1_WORKSTREAMS/GO_BACKEND_G1_CONTRACT.md`) and shared with Frontend for review
   **before** being called "frozen" (§4).
4. Existing `govapi` tests (`handler_test.go`) still pass; new tests cover both endpoints
   including the auth-rejection and empty-result cases.

## 6. Before you open a PR (working-state checklist — every task, every PR)

```bash
cd agentgate
gofmt -l .          # must print nothing
go vet ./...        # must exit 0
go build ./...       # must exit 0
go test ./...        # every package must report ok
```

- Diff review: does every changed file trace back to Task A, B, or C above? If you found something
  else wrong while in there, record it in `docs/DECISIONS/OPEN_DECISIONS.md` or as a note in your
  handoff report — do not fix it silently in the same PR (`CLAUDE.md` scope discipline).
- Commit message describes what changed and why; do not touch `admin-ui/`, `frontend/`, or
  `gateway/` config beyond what Task C's contract review genuinely requires.

## 7. Deliverables

- Fixed `qa/g6enforcement` test file + corrective-closeout addendum in G6's `CLOSURE_SUMMARY.md`.
- `internal/credential` (or equivalent) package + tests, wired into the ALLOW path only.
- Two new `govapi` endpoints + tests + a short contract doc, reviewed by Frontend.
- Handoff report + digest per `write-handoff-package` convention, in
  `docs/PHASES/G7_WORKSTREAMS/results/` (gitignored).

## 8. Explicit non-goals

- Do not resolve O-004 (MCP revision support) as part of this ticket.
- Do not add a tool-classification write path (§5, Task C non-goal).
- Do not modify `admin-ui/` or `gateway/` configuration yourself — hand contract questions to the
  relevant team.
- Do not reopen or re-verdict G6's `CLOSURE_SUMMARY.md` — addendum only.
