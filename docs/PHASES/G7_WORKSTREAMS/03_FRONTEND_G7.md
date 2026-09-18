# G7 Ticket — Frontend Team (`admin-ui/`, `frontend/`)

**Status:** OPEN — ready to start now.
**Team:** Frontend. **Plan:** `docs/PHASES/AGENTGATE_V1_3_TEAM_PARALLEL_EXECUTION_PLAN.md` §7.
**What comes after this checkpoint:** `docs/PHASES/PROGRESS_AND_ROADMAP.md` — the checkpoint tracker (G1→G11), with your team's next tasks already scoped and their dependencies marked.
**This file is self-contained.** Unlike Backend/AI-Gateway, none of your tasks below are blocked
on anyone — you are the team others are waiting on for one thing (§3), not the reverse.

---

## 0. Read first

- `admin-ui/FLOW_AND_ARCHITECTURE.md` — already corrected as of 2026-09-18 to state accurately
  which panels are real vs. mock and why (the previous version's reasoning for the Audit and Tools
  panels was stale, referencing a pre-G5 backend state; that's now fixed). Read the corrected
  version, not from memory.
- `admin-ui/README.md` §"Verified" — also corrected 2026-09-18: the claimed Playwright coverage was
  a one-time manual session, not committed/reproducible tests. Task B below is what actually closes
  that gap.

## 1. Non-negotiables

- Every displayed policy/decision state must come from the backend — never imply success before
  the server confirms it.
- Mock data lives only under `admin-ui/src/mock/` — never mixed into `frontend/src/models` (the
  frozen contract layer other teams read from). This is already true today; keep it that way.
- Do not describe `admin-ui` as "the" production UI in any durable doc — that's still open
  (`docs/DECISIONS/OPEN_DECISIONS.md` O-009). Keep building it; just don't overstate its status.

## 2. Task A — Draft the two new read-API contracts for Backend to review

Backend's ticket (`01_BACKEND_G7.md` §5) is building two new endpoints. Rather than wait for them
to propose a shape and you review it, **draft the shape yourself now** — you have everything
needed already, and this gives you real work instead of idling:

1. **Audit query** — propose the JSON response shape for
   `GET /api/v1/workspaces/{workspace_id}/audit-events`, based on the real Go type already in the
   codebase (`agentgate/internal/audit/types.go`'s `StoredRecord`): `id`, `workspace_id`,
   `sequence_number`, `execution_id`, `timestamp`, `event_type`, `decision`, `reason`,
   `principal_agent_id`, `principal_roles`, `principal_on_behalf_of`, `tool_backend_id`,
   `tool_name`, `tool_risk`, `policy_version`, `policy_hash`, `redacted_arguments`. Decide what
   pagination shape you want (`cursor`+`next_cursor`, or `before_sequence`+`limit`) — propose one,
   don't leave it open-ended. Note explicitly whether you need `canonical_payload`/`prev_hash`/
   `row_hash` (internal chain-verification fields) surfaced in the UI at all, or whether the
   backend should exclude them by default.
2. **Tool inventory** — propose the shape for `GET /api/v1/workspaces/{workspace_id}/tools`, based
   on `agentgate/internal/toolregistry/toolregistry.go`'s `GovernanceRecord`: `ToolID` (itself
   `BackendID`+`ToolName`), `Known`, `Risk`, `RegisteredFingerprint`. **Important constraint to
   design around:** this will list only *statically configured* tools, not tools discovered from
   live traffic — there is no dynamic discovery mechanism yet. Do not design the Tools & Resources
   panel's copy/empty-states around an assumption of discovering new unclassified tools; that's not
   what this API will deliver in this checkpoint.

Write both as a short markdown doc, e.g. `admin-ui/PROPOSED_G8_API_CONTRACTS.md`, and hand it to
Backend team. This does not block you from anything else below.

## 3. Task B — Automated test coverage for the two real panels

**This is the task the project is depending on you for** — `admin-ui`'s own README previously
overstated its test coverage (corrected 2026-09-18); this closes that gap for real, starting with
the two panels that are genuinely real (not mock):

1. **Policies workflow** — the full list → create candidate → validate → dry-run → activate →
   rollback flow, against `MockGovernanceClient` (deterministic, no network needed) at minimum.
2. **Decision Tester** — exercising the fixture set it already ships with.

Use `frontend/`'s existing pattern as your reference — it already has full Vitest coverage (63
tests, 7 suites) for the same underlying contract layer `admin-ui` consumes; you're testing the
*rendering and interaction* on top of that already-tested contract, not re-testing the contract
itself. Pick Vitest + React Testing Library for component/interaction tests (add them to
`admin-ui/package.json`'s `devDependencies` and add a `"test"` script — there isn't one yet, only
`dev`/`build`/`preview`/`typecheck`). Playwright is acceptable too if you prefer true
browser-driven E2E, but whichever you choose, it must be **committed, config-and-all, and runnable
with one command** — the previous manual walkthrough's core flaw was that nobody else could re-run
it.

**DoD for Task B:**
1. A `test` script exists in `admin-ui/package.json` and running it exercises both panels.
2. Tests are deterministic (no reliance on a live backend being up) — use the mock client.
3. At least one test would fail if activation were displayed as successful before the backend
   confirmed it (guards the non-negotiable in §1).
4. `admin-ui/README.md` "Verified" section is updated again to describe the *actual* committed
   coverage, replacing the "not yet automated" caveat added 2026-09-18.

## 4. Task C — Continue normal admin-ui development

Outside of Tasks A/B, keep building per the UI's own existing, already-good architecture doc.
Nothing here is blocked. Do not start wiring the Audit or Tools panels to a real endpoint yet —
those endpoints don't exist until Backend's Task C (`01_BACKEND_G7.md` §5) ships and you've
reviewed the shape (Task A). When they do, that wiring is G8 work (next checkpoint), not this
ticket.

## 5. Collaboration & Blockers

| Your task | Depends on | From whom | What happens if you skip ahead anyway |
|---|---|---|---|
| Task A (draft contracts) | Nothing — start immediately | — | — |
| Task B (test coverage) | Nothing — start immediately | — | — |
| Task C (continued dev) | Nothing, **except**: don't wire Audit/Tools panels to a real endpoint | Backend's Task C must ship first, and you must review its shape (Task A) | Wiring against a guessed, unreviewed shape risks rework once the real endpoint's actual shape lands — wait for the review loop to close. |
| — (others' dependency on you) | Backend's Task C contract *freeze* is blocked on your Task A review | Backend team is waiting on you | Do Task A promptly — you are the one other teams are blocked on here, not the other way around. |

## 6. Before you open a PR

```bash
cd admin-ui
npm run typecheck
npm run build
npm test            # once Task B adds this script
```
```bash
cd frontend
npm run typecheck
npm run build
npm test
```

- Confirm `admin-ui/src/mock/` is still the only place new mock data lives — a change that starts
  mixing mock shapes into `frontend/src/models` is a boundary violation, not a small refactor;
  don't do it, flag it instead.
- Diff review: nothing here should touch `agentgate/` Go source or `gateway/`/`deploy/` config.

## 7. Deliverables

- `admin-ui/PROPOSED_G8_API_CONTRACTS.md`, reviewed with Backend team.
- Committed, runnable test coverage for the Policies workflow and Decision Tester.
- `admin-ui/README.md` updated to reflect the real (now automated) coverage.
- Handoff report + digest in `docs/PHASES/G7_WORKSTREAMS/results/` (gitignored).

## 8. Explicit non-goals

- Do not wire the Audit Logs or Tools & Resources panels to a real backend endpoint this ticket —
  the endpoints don't exist yet (that's G8, after Backend's Task C ships).
- Do not implement a tool-classification write path in the UI — no backend write path exists or is
  planned in this checkpoint (`01_BACKEND_G7.md` §5 non-goal).
- Do not build the structured Cedar rule-builder, real OIDC login, or the public docs/landing
  sites — all explicitly deferred per the 3-team plan §7/§8.
- Do not declare `admin-ui` "the" production UI framework in any doc — O-009 is still open.
