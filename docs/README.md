# AgentGate — Documentation Map

Start here whenever you're not sure where something lives or what's currently true. This page
doesn't duplicate content — it says what exists, where, and why, so nothing has to be
rediscovered by scrolling through chat history.

**Right now:** `docs/DEVELOPMENT/CURRENT_STATUS.md` — one page, always current. If you only read
one thing, read that. (As of 2026-10-14: the G1–G7 checkpoint-folder process below closed after
G7; G8 through G11 and everything since — v1.0.0-rc1 production readiness, the hardening pass, the
website/demo pack — happened under a different, skill-driven planning approach,
`docs/superpowers/plans/` + `.superpowers/` working ledgers, gitignored and not part of this
folder structure. This map's file index below reflects the G1–G7 era; treat anything it doesn't
mention — `docs/OPERATIONS/RUNBOOK.md`'s current content, `docs/DEMO_SCRIPT.md`,
`docs/demo-scenarios/`, `docs/handover/` — as real and current regardless.)

**How we work:** `/WORKFLOW.md` (repository root) — roles, the ChatGPT↔Claude↔you loop, the
per-checkpoint codebase-documentation requirement, and the folder conventions this map describes.
Self-contained; this is the file to hand a Lead Architect session that has no repo access.

## The five kinds of document in this repo

### 1. Core project context — rarely changes, read once

The permanent "what and why," written to be read cold by a new engineer or a new AI session with
no other context required. Change these only when something actually, factually changes about
the product/stack — not for routine progress.

### 2. Living status and decisions — always current, no history to scroll

Rewritten in place as things change. Never append a history log to these — that belongs in a
checkpoint's own docs (kind 4/5 below).

### 3. Strategy — exactly one active plan at a time

There have been several planning strategies as the project's understanding matured. Old
strategies are **never deleted** — deleting them would erase the reasoning trail — but each
carries an explicit "superseded by X" banner at the top the moment it stops being current, so
nobody executes a dead plan by accident. **The active plan is always the one with no "superseded"
banner on it.**

### 4. Checkpoints — one self-contained folder per checkpoint, same shape every time

Each checkpoint (G1, G2, ...) gets `docs/PHASES/G{N}_WORKSTREAMS/`, always structured the same
way so there's nothing new to learn at G2, G3...:

```
docs/PHASES/G{N}_WORKSTREAMS/
├── 00_G{N}_CHECKPOINT_REFERENCE.md   the shared definition-of-done for this checkpoint
├── 0X_<WORKSTREAM>_DETAILED.md       per-workstream task tickets (what Claude executes)
├── <CONTRACT_OR_EVIDENCE_DOCS>.md    durable output: frozen contracts, security evidence,
│                                      environment references — committed, permanent
├── CLOSURE_SUMMARY.md                written once the checkpoint merges — the durable,
│                                      human-readable record of what shipped AND a codebase
│                                      walkthrough with diagrams (see kind 5 and /WORKFLOW.md §4)
└── results/                          GITIGNORED — ephemeral reports + digests for the
                                       human↔ChatGPT review loop; never permanent, never
                                       referenced from code or from durable docs
```

Raw prompts (the literal text fed to Claude, usually written by ChatGPT) live in
`docs/prompts/G{N}/` — also gitignored; they're working material, not project history.

### 5. Checkpoint closure summaries — the durable "what happened" record

At the end of each checkpoint, once its branches merge, one `CLOSURE_SUMMARY.md` is written into
that checkpoint's folder. It exists specifically so a human reviewer never has to lose track of
what AI-written code actually does: a plain-language recap of what shipped, a **codebase
walkthrough with at least one diagram** explaining how the new pieces fit together and why they
were built that way, and what's carried forward. Full requirements in `/WORKFLOW.md` §4.

### 6. Archive — superseded or point-in-time, kept for history, never "current" again

`docs/PHASES/archive/` — one flat folder, no sub-structure, for anything that has permanently
stopped describing the present: dead strategies once their "superseded" banner is added, and
one-time snapshot reports (like a repository recon done before any code existed). Nothing here
is deleted; everything here is annotated with why it's here and what replaced it.

## Full file index

```
(repo root)/
└── WORKFLOW.md                                  [core, canonical] the workflow doc — see above

docs/
├── README.md                                    you are here — the map
├── PROJECT_DEFINITION.md                        [core] the single source of truth: product
│                                                 scope, architecture, what we explicitly aren't
│                                                 building, the decisions log (§12)
├── TECH_STACK.md                                [core] technology choices + rationale; §4's own
│                                                 "O-N" open-items table is historical — see its
│                                                 2026-09-13 status note for what's resolved vs.
│                                                 superseded by OPEN_DECISIONS.md
│
├── DECISIONS/
│   └── OPEN_DECISIONS.md                        [living] the canonical unresolved-question
│                                                 registry (O-001...O-009); resolved items move to
│                                                 its own "Resolved" section, never deleted
│
├── SECURITY/
│   └── PRODUCTION-INVARIANTS.md                 [living, binding] the security contract. No
│                                                 later work may contradict it without an
│                                                 explicit, recorded exception
│
├── DEVELOPMENT/
│   ├── CURRENT_STATUS.md                        [living] where the project is *right now* —
│                                                 rewritten in place, not appended to
│   ├── AI_DEVELOPMENT_MODEL.md                  [redirect stub] moved to /WORKFLOW.md at repo
│                                                 root; kept only so existing references to this
│                                                 path keep working
│   ├── SETUP.md                                 [living reference] local build/run/test guide
│                                                 for the agentgate/ Go module
│   ├── CI_BASELINE.md                           [living reference] what CI checks and why
│   ├── BRANCHING_AND_MERGING.md                 [living reference] git workflow for 3-team
│   │                                             independence model; branch naming, when to merge,
│   │                                             cross-team coordination points
│   └── OSS_READINESS.md                         [living backlog] what's needed before public/OSS
│                                                 launch (license decision, community-health
│                                                 files) — deliberately deferred, tracked here so
│                                                 it isn't lost
│
├── PHASES/
│   ├── PROGRESS_AND_ROADMAP.md                   [living tracker] the single answer to "what's
│   │                                             done, what's in flight, what's next, who's
│   │                                             blocked on whom" — every checkpoint G1→G11
│   │                                             broken down by team, with dependencies marked
│   │                                             IND/DEP. Start here for scheduling questions
│   ├── AGENTGATE_V1_10_DAY_PARALLEL_TEAM_EXECUTION_PLAN.md
│   │                                             [superseded, kept] 5-workstream gated model that
│   │                                             carried G1–G6 to closure; superseded once its
│   │                                             critical path (G1–G6) froze
│   ├── AGENTGATE_V1_15_DAY_PRODUCTION_PLAN.md   [superseded, kept] sequential 15-day plan;
│   │                                             superseded by the 10-day parallel plan after its
│   │                                             own Day 1/Day 2 completed
│   ├── DAY-01-TASK-01.md                        [complete, kept] the Day-1 task spec that
│   │                                             produced PRODUCTION-INVARIANTS.md — historical
│   │                                             record of what was assigned, still cited by
│   │                                             living docs
│   ├── DAY-02-TASK-02.md                        [complete, kept] the Day-2 task spec that
│   │                                             produced the decision core now frozen as the G1
│   │                                             contract
│   │
│   ├── G1_WORKSTREAMS/                          [CLOSED CHECKPOINT — G1: PASS/CLOSED/FROZEN, 2026-09-12]
│   ├── G2_WORKSTREAMS/                          [CLOSED CHECKPOINT — G2: PASS/CLOSED/FROZEN, 2026-09-13]
│   ├── G3_WORKSTREAMS/                          [CLOSED CHECKPOINT — G3: PASS/CLOSED/FROZEN, 2026-09-13]
│   ├── G4_WORKSTREAMS/                          [CLOSED CHECKPOINT — G4: PASS/CLOSED/FROZEN, 2026-09-14]
│   ├── G5_WORKSTREAMS/                          [CLOSED CHECKPOINT — G5: PASS/CLOSED/FROZEN, 2026-09-14]
│   ├── G6_WORKSTREAMS/                          [CLOSED CHECKPOINT — G6: PASS/CLOSED/FROZEN, 2026-09-16,
│   │   │                                         evidence-integrity addendum pending, see G7]
│   │                                             Real MCP End-to-End Enforcement — the last
│   │                                             checkpoint run under the superseded 5-workstream
│   │                                             model. See its `CLOSURE_SUMMARY.md`. Its
│   │                                             implementation is real and sound; an independent
│   │                                             review (2026-09-18) found its E2E *evidence* was
│   │                                             overstated (see G7 Task 0 below) — the fix is
│   │                                             corrective closeout, not a reopened verdict.
│   │
│   ├── G7_WORKSTREAMS/                          [CLOSED CHECKPOINT — first under the 3-team model,
│   │   │                                         opened 2026-09-18, closed 2026-10-08 per
│   │   │                                         CURRENT_STATUS.md] Downstream credential exchange
│   │   │                                         (O-001), G6 evidence closeout, and groundwork for
│   │   │                                         a realistic (non-toy) demonstration system.
│   │   │                                         NOTE: unlike every other closed checkpoint, this
│   │   │                                         folder has no CLOSURE_SUMMARY.md or
│   │   │                                         00_G7_CHECKPOINT_REFERENCE.md — a real gap,
│   │   │                                         flagged 2026-10-14, not yet resolved.
│   │   ├── 01_BACKEND_G7.md                     Backend team's ticket — self-contained
│   │   ├── 02_AI_GATEWAY_G7.md                  AI/Gateway team's ticket — self-contained
│   │   ├── 03_FRONTEND_G7.md                    Frontend team's ticket — self-contained
│   │   └── results/                             GITIGNORED — ephemeral reports + digests
│   │
│   ├── (G8–G11 ran without this checkpoint-folder process — no docs/PHASES/G{8,9,10,11}_WORKSTREAMS/
│   │   exist. Their outcomes are summarized in CURRENT_STATUS.md's Checkpoint Milestones only.)
│   │
│   └── archive/                                 [historical — see docs/README.md kind 6]
│       ├── MASTER_PLAN.md                       original Phase 0-7 plan; superseded after Phase 0
│       ├── PHASE-01-MINIMAL-E2E-ENFORCEMENT.md  2nd-generation Phase-1 plan; superseded
│       ├── PHASE-01-TASKS.md                    its task register; superseded
│       ├── TASK-01-01-GO-MODULE-AND-REPOSITORY-SCAFFOLD.md
│       │                                         completed task spec, kept for history
│       ├── REPOSITORY_BASELINE.md               one-time repo recon snapshot (2026-08-22,
│       │                                         before any code existed) — permanently
│       │                                         historical, never "current" again
│       ├── AGENTGATE_V1_3_TEAM_PARALLEL_EXECUTION_PLAN.md
│       │                                         3-team operating model that carried G7 to close;
│       │                                         archived 2026-10-14 after being found still
│       │                                         marked "active" here with no superseded banner,
│       │                                         long after G8-G11 had superseded it in practice
│       └── PROGRESS_AND_ROADMAP.md              checkpoint tracker; archived 2026-10-14, stopped
│                                                 being updated after G7
│
└── prompts/                                     GITIGNORED — raw prompt text fed to Claude,
    ├── phase0/                                  working material, not project history
    │   ├── 01_repository_bootstrap_prompt.md
    │   ├── 02_ci_baseline_prompt.md
    │   └── 03_task-01-01_prompt.md
    └── g1/
        ├── 01_GO_BACKEND_G1_PROMPT.md
        ├── 02_GATEWAY_MCP_G1_PROMPT.md
        ├── 03_FRONTEND_UI_G1_PROMPT.md
        ├── 04_QA_SECURITY_G1_PROMPT.md
        └── 05_DEVOPS_G1_PROMPT.md
```

## Quick answers

- **"What are we building and why?"** → `docs/PROJECT_DEFINITION.md`
- **"What's the state of things today?"** → `docs/DEVELOPMENT/CURRENT_STATUS.md`
- **"What plan are we actually following?"** → `docs/DEVELOPMENT/CURRENT_STATUS.md` names the
  active plan directly (as of 2026-10-14: work under `docs/superpowers/plans/`, gitignored). The
  G1–G7-era strategy docs below are all now historical — `AGENTGATE_V1_3_TEAM_PARALLEL_EXECUTION_PLAN.md`
  moved to `archive/` 2026-10-14 after being found stale with no banner.
- **"What happened at G1 / G2 / ... / G7?"** → that checkpoint's `CLOSURE_SUMMARY.md` (G7 is
  missing one — see the note by its entry above), or the full detail in its
  `docs/PHASES/G{N}_WORKSTREAMS/` folder. For G8-G11, see CURRENT_STATUS.md's Checkpoint
  Milestones — no per-checkpoint folder exists for those.
- **"What's done, what's next, and who's blocked on whom?"** →
  `docs/DEVELOPMENT/CURRENT_STATUS.md` for the current answer. `docs/PHASES/archive/PROGRESS_AND_ROADMAP.md`
  is the old tracker (G1-G7 only, archived 2026-10-14).
- **"What's the frozen authorization contract?"** → `docs/PHASES/G1_WORKSTREAMS/GO_BACKEND_G1_CONTRACT.md`
- **"Is there an open question blocking this?"** → `docs/DECISIONS/OPEN_DECISIONS.md`
- **"What security invariant governs this?"** → `docs/SECURITY/PRODUCTION-INVARIANTS.md`
- **"How do I set up and run this locally?"** → `docs/DEVELOPMENT/SETUP.md`
- **"What does CI check?"** → `docs/DEVELOPMENT/CI_BASELINE.md`
- **"How should I branch and merge for this project?"** → `docs/DEVELOPMENT/BRANCHING_AND_MERGING.md`
- **"What's left before this can be public/OSS?"** → `docs/DEVELOPMENT/OSS_READINESS.md`
- **"Is this document still current, or old?"** → if it's under `docs/PHASES/archive/`, or it has
  a "superseded by X" banner at the top, it's history, not instruction. Everything else is
  current as of its own last-edit date.
