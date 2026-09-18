# Branching and Merging Strategy

**For:** 3-team independent development with clear checkpoints and cross-team coordination.
**Stability guarantee:** `development` is always in a working state (passes CI, no broken tests).
**Updated:** 2026-09-18

---

## Branch Structure

```
master
  ↑
  └─ (release tags: v1.0.0, v1.0.1, ...)
  
development (stable integration, always green)
  ↑
  ├─ feature/backend-g7-task-a (Team A)
  ├─ feature/backend-g7-task-b (Team A)
  ├─ feature/gateway-g7-task-a (Team B)
  ├─ feature/gateway-g7-task-b (Team B)
  ├─ feature/frontend-g7-task-a (Team C)
  ├─ feature/frontend-g7-task-b (Team C)
  └─ fix/documentation-correction-yyyy-mm-dd (Anyone, as needed)
```

**Naming convention:** `<type>/<team>-g<checkpoint>-task-<letter>`

| Type | When | Examples |
|---|---|---|
| `feature/` | New code, endpoints, functionality | `feature/backend-g7-task-b` |
| `fix/` | Bug fixes, doc corrections, security patches | `fix/documentation-update-2026-09-18` |
| `experimental/` | Throwaway spikes, not for merging | `experimental/gateway-oauth-probe` |

---

## Per-Team Workflow

### 1. Start a task

**Create a feature branch from `development`:**
```bash
git fetch origin
git checkout -b feature/backend-g7-task-a origin/development
```

**Work independently.** Your branch is yours — force-push as needed while working.

### 2. While working

**Keep your branch up to date** (optional, but recommended if `development` gets ahead):
```bash
git fetch origin
git rebase origin/development
# or, if you prefer merges:
git merge origin/development
```

**Commit early and often.** Small, logical commits are easier to review and debug.

**Run checks locally before pushing:**
```bash
# Backend
cd agentgate
gofmt -l .
go vet ./...
go test ./...

# AI/Gateway
cd gateway/harness
go test ./...
docker compose -f deploy/g6/docker-compose.yml config -q

# Frontend
cd frontend/app
npm run typecheck
npm run build
npm test
```

### 3. Ready for review

**Push your branch:**
```bash
git push origin feature/backend-g7-task-a
```

**Open a PR** (via GitHub) against `development`:
- Title: one sentence, describing the change
- Description: per `/WORKFLOW.md` — what shipped, diagram (if applicable), where to look, carried-forward items
- Assign: your team lead or a peer from your team for internal review first
- Link: the checkpoint ticket this completes (e.g., `docs/PHASES/G7_WORKSTREAMS/01_BACKEND_G7.md` §2)

**Wait for CI to pass.** All checks must be green before merging:
- `.github/workflows/ci.yml` (Go tests, linting, builds)
- Code review approval (at least 1, from your team)

### 4. Merge to development

**One of:**
1. **Squash and merge** (simplest, if your commits are exploratory)
   - Use if: many small "work in progress" commits
   - Results in: one clean commit on `development`
   ```bash
   # GitHub UI: "Squash and merge"
   ```

2. **Rebase and merge** (preserves your commit history)
   - Use if: each commit stands on its own
   - Results in: your commits appear linearly on `development`
   ```bash
   # GitHub UI: "Rebase and merge"
   ```

3. **Create a merge commit** (shows the branch existed)
   - Use if: you want the branch history visible
   - Results in: a merge commit + your commits visible
   ```bash
   # GitHub UI: "Create a merge commit"
   ```

**Default:** squash and merge (keeps `development` history clean for the 3-team model).

**After merge:**
- Delete the feature branch (GitHub UI option after merge)
- Pull `development` locally to stay synced
  ```bash
  git fetch origin
  git checkout development
  git pull origin development
  ```

---

## Cross-Team Dependencies

Three documented cross-team points in G7 (see `PROGRESS_AND_ROADMAP.md` §6):

| Blocking | Blocked on | Coordination |
|---|---|---|
| Backend Task B (credential implementation) | AI/Gateway Task B (`CREDENTIAL_REQUIREMENTS.md`) | AI/Gateway opens a PR with just the requirements doc → Backend reviews before coding the mechanism |
| Backend Task C (read API freeze) | Frontend Task A (contract review) | Frontend opens a PR with proposed contracts → Backend reviews and updates implementation if needed |
| Backend Task A (live test run) | AI/Gateway Task A (portable deploy script) | AI/Gateway merges their fix first → Backend runs their live suite once both are merged |

**Coordination flow:**
1. Team A opens a PR with their blocking output (requirements, contract proposal, etc.)
2. Team B reviews that PR in their own context (does it match their understanding?)
3. Once approved, merge it
4. Team B continues their own work knowing the contract is stable

**No separate coordination PRs.** The deliverable PR *is* the coordination point.

---

## Checkpoint Closure

When a task is complete and ready to close:

### 1. Write closure artifacts

Per `/WORKFLOW.md` §4:
- `CLOSURE_SUMMARY.md` in `docs/PHASES/G{N}_WORKSTREAMS/`
- Handoff report + digest in `docs/PHASES/G{N}_WORKSTREAMS/results/` (gitignored)

### 2. Merge handoff report

Open a final PR titled **`docs: close <Team> G{N} <Task Letter>`** containing:
- The `CLOSURE_SUMMARY.md` (new file or update to existing)
- Any final doc corrections
- Updated `PROGRESS_AND_ROADMAP.md` (if this team's task is done)

**No code changes in closure PRs** — they are documentation only.

### 3. Tag the closure commit

Once the closure PR merges:
```bash
git fetch origin
git checkout development
git pull origin development
git tag -a g7-backend-task-a-2026-09-20 -m "Backend G7 Task A closed"
git push origin g7-backend-task-a-2026-09-20
```

**Tag format:** `g<checkpoint>-<team>-task-<letter>-YYYY-MM-DD`

This makes it easy to find when a task closed and what commit it was.

---

## Releasing (master)

When a checkpoint is fully closed by all teams (see `PROGRESS_AND_ROADMAP.md` §2):

```bash
git fetch origin
git checkout master
git pull origin master
git merge origin/development
git tag -a v1.<checkpoint> -m "G{N} checkpoint release"
git push origin master
git push origin v1.<checkpoint>
```

**Only the Lead Architect does this.** Timing: once all teams' G{N} work is closed and checkpoint is PASS/CLOSED/FROZEN.

---

## What NOT to do

| ❌ Do not | Why |
|---|---|
| Force-push to `development` or `master` | Loses history for other teams; breaks their local clones |
| Commit directly to `development` | Bypasses review; skips CI validation |
| Merge code before CI passes | Hidden failures surface in `development` |
| Open a PR without a corresponding ticket link | Other teams can't trace why the change exists |
| Leave feature branches stale (>2 weeks without activity) | Indicates blocked/abandoned work; unclear status |
| Squash commits that should stay separate | Destroys useful history for debugging later |
| Cherry-pick between teams' branches | Creates merge conflicts and duplicate fixes; always go through `development` |

---

## Local troubleshooting

**Your branch is behind `development`:**
```bash
git fetch origin
git rebase origin/development
# or merge if you prefer:
git merge origin/development
```

**You need to undo the last commit (not yet pushed):**
```bash
git reset --soft HEAD~1
# Edit files as needed
git commit -m "..."
```

**You need to undo the last commit (already pushed):**
```bash
# Do NOT force-push. Instead, create a new commit that reverts:
git revert HEAD
git push origin feature/your-branch
```

**You accidentally committed to `development` locally:**
```bash
# Create a feature branch from where you are
git branch feature/oops-fix
# Reset development to origin
git checkout development
git reset --hard origin/development
# Switch to your branch with the accidental commit
git checkout feature/oops-fix
# Now open a normal PR from feature/oops-fix
```

---

## Questions?

- **Merge vs. rebase?** Ask in your team Slack — there's no single right answer, just consistency within your team's workflow.
- **Should I push work-in-progress?** Only if it's been 24h+ of work you want to back up. Otherwise, keep it local.
- **Can I work on two tasks at once?** Yes — create two branches and switch between them. Keep them independent.
- **Do I need to ask permission to create a branch?** No. Create as many as you need.
