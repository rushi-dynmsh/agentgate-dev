# G7 Ticket — Frontend Team (`frontend/app/`, `frontend/`)

**Status:** OPEN — Tasks A and B below are largely already satisfied by existing work; verify and
close the remaining gap (§2), then continue with Task D.
**Team:** Frontend. **Plan:** `docs/PHASES/AGENTGATE_V1_3_TEAM_PARALLEL_EXECUTION_PLAN.md` §7.
**What comes after this checkpoint:** `docs/PHASES/PROGRESS_AND_ROADMAP.md` — the checkpoint
tracker (G1→G11), with your team's next tasks already scoped and their dependencies marked.
**This file is self-contained.** None of your tasks below are blocked on anyone — you are the
team others are waiting on for one thing (§4), not the reverse.

---

## 0. Read first — and important context on this ticket's history

- `frontend/FRONTEND_IMPLEMENTATION.md` — the app's structural guide.
- `frontend/app/README.md` §"Verified" — states current, real test coverage.
- **This ticket originally targeted a different app, `admin-ui/`, built by a different teammate
  earlier the same day.** Two people independently built a full admin UI on 2026-09-18; `admin-ui/`
  was removed and `frontend/app/` kept as the production UI (**O-009, resolved** —
  `docs/DECISIONS/OPEN_DECISIONS.md`). Retargeted 2026-09-18. If you did the work described in §§2-3
  below before this retarget, it likely already satisfies most of this ticket — see the status
  notes in each task.

## 1. Non-negotiables

- Every displayed policy/decision state must come from the backend — never imply success before
  the server confirms it. (Already verified in `frontend/app`'s own test suite — see §3.)
- Mock data/fixtures stay separate from the frozen `frontend/src` contract layer other teams read
  from — `frontend/app` correctly imports it via the `@contract` alias rather than reimplementing
  it (verified: `frontend/app/vite.config.ts`'s `resolve.alias`).
- `frontend/app/` is now the settled production UI (O-009 resolved) — safe to describe it as such
  in durable docs going forward.

## 2. Task A — API contracts for Backend to review — **already done, verify only**

`frontend/app/PROPOSED_G8_API_CONTRACTS.md` already proposes both endpoint shapes (audit query
with `before_sequence`+`limit` pagination; tool inventory), matching Backend's ticket
(`01_BACKEND_G7.md` §5) closely. **Action needed:** confirm Backend has actually reviewed this
specific file (not a duplicate one someone else wrote) and hand it to them directly if not — this
is the one place another team is waiting on you (§4).

## 3. Task B — Automated test coverage for the real panels — **already substantially done**

`frontend/app` already has a `test` script (`vitest run`) and 4 real test files: policy lifecycle
(list → validate → create candidate → dry-run → activate → rollback) against
`MockGovernanceClient`, an explicit test that activation stays pending until the client promise
resolves (guards §1's non-negotiable directly), and decision-fixture rendering tests.

**One real difference from the original spec, not a gap:** the original ticket asked for
"Decision Tester panel" coverage. `frontend/app` has no separate Decision Tester screen — it tests
the underlying decision fixtures directly (`decision-fixtures.test.ts`). This is an acceptable
design choice, not something to retrofit a screen for just to match the old ticket's wording.

**Action needed:** confirm `frontend/app/README.md`'s "Verified" section (already present and
accurate as of this writing) stays in sync as more tests are added — no rewrite needed today.

## 4. Task C — Continue normal `frontend/app/` development

Outside of Tasks A/B, keep building. Nothing here is blocked. Do not start wiring the Audit or
Tools pages to a real endpoint yet — those endpoints don't exist until Backend's Task C
(`01_BACKEND_G7.md` §5) ships and reviews your Task A contract doc. When they do, that wiring is
G8 work (next checkpoint), not this ticket.

## 5. Task D — Write the real-vs-mock accounting `frontend/app/` currently lacks

The now-removed `admin-ui/` had a rigorous, screen-by-screen `FLOW_AND_ARCHITECTURE.md` (🟢 real /
🟡 mock-grounded / 🔴 reframed, with reasoning for each). `frontend/app/`'s `README.md` and
`FRONTEND_IMPLEMENTATION.md` don't have this level of accounting yet. Write one, using the old
file as a model for rigor, not content (`git show 98a038b:admin-ui/FLOW_AND_ARCHITECTURE.md` —
the screens differ, e.g. no Decision Tester, has Login/Settings differently). This matters because
without it, nothing stops this app from drifting into the same "claims more than it delivers" gap
the old one had before correction.

## 6. Collaboration & Blockers

| Your task | Depends on | From whom | What happens if you skip ahead anyway |
|---|---|---|---|
| Task A verification | Nothing — start immediately | — | — |
| Task B verification | Nothing — start immediately | — | — |
| Task C (continued dev) | Nothing, **except**: don't wire Audit/Tools to a real endpoint | Backend's Task C must ship first, and must have reviewed your Task A doc | Wiring against an unreviewed shape risks rework once the real endpoint's shape lands. |
| Task D | Nothing — start immediately | — | — |
| — (others' dependency on you) | Backend's Task C contract *freeze* is blocked on confirming your Task A doc was actually reviewed | Backend team is waiting on you | Confirm the hand-off explicitly — don't assume they found `frontend/app/PROPOSED_G8_API_CONTRACTS.md` on their own. |

## 7. Before you open a PR

```bash
cd frontend/app
npm run typecheck
npm run build
npm test
```
```bash
cd frontend
npm run typecheck
npm run build
npm test
```

- Confirm no fixture/mock data starts leaking into `frontend/src/models` (the frozen contract
  layer) — a boundary violation, not a small refactor; flag it instead of fixing it silently.
- Diff review: nothing here should touch `agentgate/` Go source or `gateway/`/`deploy/` config.

## 8. Deliverables

- Confirmation that Backend has reviewed `frontend/app/PROPOSED_G8_API_CONTRACTS.md` specifically.
- A new real-vs-mock accounting doc for `frontend/app/` (Task D).
- Handoff report + digest in `docs/PHASES/G7_WORKSTREAMS/results/` (gitignored).

## 9. Explicit non-goals

- Do not wire the Audit or Tools pages to a real backend endpoint this ticket — the endpoints
  don't exist yet (that's G8, after Backend's Task C ships).
- Do not implement a tool-classification write path — no backend write path exists or is planned
  in this checkpoint (`01_BACKEND_G7.md` §5 non-goal).
- Do not build the structured Cedar rule-builder, real OIDC login, or the public docs/landing
  sites — all explicitly deferred per the 3-team plan §7/§8.
- Do not retrofit a "Decision Tester" screen just to match this ticket's original wording — §3
  explains why the current approach (testing fixtures directly) is an acceptable substitute.
