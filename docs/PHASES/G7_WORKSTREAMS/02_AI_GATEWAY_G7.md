# G7 Ticket — AI / Gateway Team (`gateway/`, `deploy/`, MCP, demonstration system)

**Status:** OPEN — ready to start now.
**Team:** AI / Gateway. **Plan:** `docs/PHASES/AGENTGATE_V1_3_TEAM_PARALLEL_EXECUTION_PLAN.md` §6.
**What comes after this checkpoint:** `docs/PHASES/PROGRESS_AND_ROADMAP.md` — the checkpoint tracker (G1→G11), with your team's next tasks already scoped and their dependencies marked.
**This file is self-contained** — you do not need to read the other two teams' tickets to start
Task A or begin Task B's backend-selection work. Read §4 before you need Backend's credential
mechanism.

---

## 0. Read first

- `docs/PHASES/G6_WORKSTREAMS/CLOSURE_SUMMARY.md` — the enforcement path you're building on top of.
- `deploy/g6/README.md` and `deploy/g6/docker-compose.yml` — the existing topology (5 services
  defined; `probe-authz` is `profiles: ["probe-only"]` and not started by default — the real
  4-service default topology is `postgres`, `agentgate`, `probe-mcp`, `agentgateway`).
- `docs/DECISIONS/OPEN_DECISIONS.md` O-001 — you own half of resolving this (the backend/credential
  requirements half; Backend team owns the mechanism itself).

## 1. Non-negotiables

- The gateway must never bypass AgentGate for governed traffic.
- No raw inbound bearer-token passthrough to any backend, toy or real.
- Evidence for anything you claim "proven" must be reproducible by someone else, on infrastructure
  other than your own machine — see Task A; this is now a standing rule
  (`AGENTGATE_V1_3_TEAM_PARALLEL_EXECUTION_PLAN.md` §9).

## 2. Task A — G6 Evidence Closeout, environment half (pairs with Backend's Task A)

**Why:** see `01_BACKEND_G7.md` §2 for the full finding. Summary: `qa/g6enforcement`'s live-path
assertions never actually run in CI or in a default local test run, because nothing stands up the
Docker topology, and the one existing live-topology proof
(`docs/PHASES/G6_WORKSTREAMS/G6_GATEWAY_CONTRACT_OBSERVED.json`) was a one-time manual capture on a
single developer's machine using a script with a hardcoded path.

**Exact bug, verified by reading the file (`deploy/g6/run-e2e-matrix.ps1`, line 10):**

```powershell
$env:GOTMPDIR = "d:\PROJECTS\AgentGate_Hackathon\agentgate-repo\.tmp"
```

This is an absolute path to one specific developer's checkout on one specific drive letter. The
rest of the script correctly uses `$PSScriptRoot`-relative paths (`Join-Path $PSScriptRoot
"docker-compose.yml"` etc.) — only this one line is wrong.

**What to do:**

1. Fix that line to be portable, e.g.:
   ```powershell
   $repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
   $env:GOTMPDIR = Join-Path $repoRoot ".tmp"
   ```
   or an equivalent that derives the path from `$PSScriptRoot` rather than hardcoding it. Test this
   actually works by running the script from a location other than the original author's checkout
   (a fresh clone, or at minimum a different directory) if at all possible.
2. Stand up `deploy/g6/docker-compose.yml` yourself and confirm it reaches a healthy state from a
   clean `docker compose up -d` (per its own health checks). Note anything that breaks doing this
   as a concrete, reproduced finding — do not guess at what might be wrong.
3. Once both fixes land, run Backend's corrected live-E2E test target (`01_BACKEND_G7.md` §2)
   against this real topology at least once. Save the output as this checkpoint's replacement
   evidence for `G6_GATEWAY_CONTRACT_OBSERVED.json`.
4. If CI can feasibly stand up this topology (a GitHub Actions job with `docker compose`), propose
   it; if not feasible in this pass, say exactly why (resource limits, secrets, etc.) rather than
   silently leaving it undone.

**DoD for Task A:**
1. `run-e2e-matrix.ps1` runs without editing any file first, from at least one environment that
   isn't the original author's machine.
2. The Docker topology comes up healthy from a clean `docker compose up -d`.
3. Backend's corrected live-E2E suite has actually been run against it, with output captured.
4. `go build ./...` and `go test ./...` still pass inside `gateway/harness` (its own module,
   `gateway/harness/go.mod`) — this fix must not break the existing harness tests
   (`gateway/harness/g1_scenarios_test.go`).

## 3. Task B — Realistic Demonstration System

**First half — do this now, does not depend on anyone:**

The current proof (`deploy/g6/`, `probe-mcp`) is a toy backend that only counts calls — correct for
isolating the enforcement boundary, but it does not demonstrate AgentGate's actual value to anyone
evaluating the product. Pick **one** concrete option and document why, before building anything:

- A well-known open-source MCP server (filesystem, GitHub, or similar) run against real or
  realistic data, or
- A small custom MCP server exposing 2–3 tools with genuine business meaning (e.g. a mock
  ticketing/CRM/database tool), with deliberately mixed risk classifications
  (`read`/`write`/`destructive`) so the governance story — deny-by-default, policy activation,
  rollback — is visibly meaningful rather than abstract.

Build this as an **additive** deployment, `deploy/demo/` — do not modify or replace `deploy/g6/`,
which stays as the enforcement-proof topology. Wire it through the same
`agentgateway → AgentGate → backend` path already proven in G6.

**Second half — this is what Backend is waiting on (§4):**

Document exactly what credential model your chosen backend needs to be called safely and
meaningfully — an API key? A scoped OAuth token? A short-lived PAT? What audience/scope would a
real deployment need? Write this as a short markdown note
(`deploy/demo/CREDENTIAL_REQUIREMENTS.md`) and hand it to Backend team. **Use a stub/test
credential in your own demo build in the meantime — do not wait for Backend's real mechanism to
build the rest of the demo.**

**DoD for Task B:**
1. A real (or realistic, non-toy) MCP backend is reachable through the proven enforcement path, in
   an additive `deploy/demo/` topology that leaves `deploy/g6/` untouched.
2. At least one tool is classified `destructive` or `write`, and a real policy change (candidate →
   validate → activate → observe) visibly changes whether it's allowed — not a hardcoded fixture.
3. A short README explains the scenario in plain language.
4. `deploy/demo/CREDENTIAL_REQUIREMENTS.md` exists and has been handed to Backend team.
5. Once Backend's real credential mechanism exists (a later task, not blocking this ticket's
   closure), this demo is updated to use it for real — track that as a follow-up note in your
   handoff report, not a new ticket.

## 4. Collaboration & Blockers

| Your task | Depends on | From whom | What happens if you skip ahead anyway |
|---|---|---|---|
| Task A | Nothing — start immediately | — | — |
| Task A step 3 (run live suite) | Backend's Task A test-code fix (`01_BACKEND_G7.md` §2) landing first | Backend team | You can still stand up and fix the environment without their fix; just run the live suite once their fix merges rather than against the still-broken tests. |
| Task B, first half (pick backend, build demo) | Nothing — start immediately | — | — |
| Task B, second half (hand off credential requirements) | Nothing on your end — this is an output, not an input | — | Backend is waiting on **you** here, not the reverse — see the row below. |
| — (Backend's dependency on you) | Backend's Task B *implementation* is blocked until you deliver `CREDENTIAL_REQUIREMENTS.md` | You owe Backend team this document | If you delay, Backend can only do design work, not build the real mechanism — pick your backend and write this doc early, don't leave it for last. |

**Reciprocal note:** you are the blocking party for one real dependency (Backend's credential
implementation), not the blocked party. Prioritize Task B's backend-selection step accordingly.

## 5. Before you open a PR

```bash
cd gateway/harness
go build ./...
go test ./...
```
```powershell
# From repo root, validate the compose file and any new deploy/demo config:
docker compose -f deploy/g6/docker-compose.yml config -q
docker compose -f deploy/demo/docker-compose.yml config -q   # once it exists
```

- Confirm `deploy/g6/` is unmodified in intent (only the one path-portability fix) — a diff that
  touches its security posture, JWT config, or `ext_authz` wiring beyond that one line needs a
  clear reason in your PR description.
- Do not touch `agentgate/internal/*` Go source yourself — that's Backend's domain; if the
  demonstration system needs a backend-side change, write it up and hand it to them rather than
  making it yourself.

## 6. Deliverables

- Fixed `run-e2e-matrix.ps1` + a real, reproduced live-topology run replacing the stale evidence
  capture.
- `deploy/demo/` — additive topology, README, `CREDENTIAL_REQUIREMENTS.md`.
- Handoff report + digest in `docs/PHASES/G7_WORKSTREAMS/results/` (gitignored).

## 7. Explicit non-goals

- Do not attempt G9's "integration point discovery" documentation task yet — that depends on
  Task B's demo system existing first and is intentionally a separate, later checkpoint
  (`AGENTGATE_V1_3_TEAM_PARALLEL_EXECUTION_PLAN.md` §6).
- Do not modify `agentgate/internal/*` Go source.
- Do not modify `frontend/app/` or `frontend/`.
- Do not implement the downstream credential *mechanism* — that's Backend's Task B; you provide
  the requirements, they build the mechanism.
