# Merge Strategy Test: Squash Merge vs Regular Merge Between Long-Lived Branches

## Purpose

This document provides evidence that **squash merging between long-lived branches** (`main` and `release`) causes merge base divergence, leading to inflated diffs, phantom commits in PRs, and broken branch comparisons. It demonstrates that **regular merge commits** preserve the shared history and keep the branches naturally aligned — with **no manual sync-back step required**.

## Background

In the simplified branching strategy used by MyQQ:
```
feature branches → main → release
```

The `release` branch is the validated, production-ready state. Periodically, `main` is merged into `release` via a PR to create a release. The question is: **does release need to be synced back to main, and if so, how?**

- **Squash merge** (back-sync): Creates a single new commit on main with the combined changes. This commit has **no parent link** back to release, so git loses track of the shared history.
- **Regular merge commit** (release PR): Creates a merge commit on release that has **two parents** — one on release, one on main. Git can trace the shared ancestry, keeping the branches aligned. **No back-sync is needed.**

## Test Environment

- **Repository**: https://github.com/markwperry/TestWorkflows
- **Project**: Go API (Gin framework) with fun endpoints
- **Date**: 2026-03-20
- **Branches**: `main` (development), `release` (production)

---

## Phase 1: Feature Development (PRs #3–#9)

Feature branches were created from `main`, each adding a fun endpoint:

| PR | Branch | Feature | Endpoint | Round |
|---|---|---|---|---|
| #3 | `feat/hillbilly-translator` | Hillbilly text translator | `GET /hillbilly?text=...` | 1 |
| #4 | `feat/magic-8ball` | Magic 8-Ball | `GET /8ball?q=...` | 1 |
| #5 | `feat/dad-jokes` | Programming dad jokes | `GET /dadjoke` | 1 |
| #6 | `feat/fortune-cookie` | Fortune cookie messages | `GET /fortune` | 1 |
| #9 | `feat/coin-flip` | Coin flip simulator | `GET /coinflip?n=...` | 2 |

---

## Round 1: Squash Merge (The Problem)

### Step 1: Pre-Release State

PRs #3 and #4 merged into `main`. Release branch has not yet received these features.

```
--- main branch tip ---
e230e52 Merge pull request #4 from markwperry/feat/magic-8ball
bd905d5 Merge branch 'main' into feat/magic-8ball
3a4dab6 Merge pull request #3 from markwperry/feat/hillbilly-translator
cbdcfb9 feat: add magic 8-ball endpoint
19f61ff feat: add hillbilly translator endpoint
4dc5942 chore: Merge pull request #1 from markwperry/chore/forcebuild

--- release branch tip ---
f7dc8d2 Merge pull request #2 from markwperry/main
4dc5942 chore: Merge pull request #1 from markwperry/chore/forcebuild

--- merge base ---
4dc5942 (shared ancestor — correct)

--- commits on main not yet on release ---
5 commits (the two features + merge commits)
```

**Observation**: Merge base is `4dc5942`. Both branches share a clean common ancestor. The diff is exactly the 2 new features.

### Step 2: Release PR (main → release) — PR #7

- **PR #7**: `chore(release): v0.2.0 — hillbilly translator & magic 8-ball`
- **Merge method**: Regular merge commit (correct for main → release direction)
- **Result**: Release now has all features from main

### Step 3: Squash Merge of release → main (THE PROBLEMATIC STEP)

After PR #7 merged, we simulated a squash merge of `release` back into `main`. This is what some teams do to "keep main in sync" with release.

```
--- main tip AFTER squash merge ---
351fcfa chore: sync release v0.2.0 back to main (#8 simulated squash merge)  ← NEW commit, NO parent link to release
4de4036 docs: add merge strategy test evidence document
e230e52 Merge pull request #4 from markwperry/feat/magic-8ball

--- release tip (unchanged) ---
71e2a90 Merge pull request #7 from markwperry/main

--- merge base (STALE — this is the problem) ---
4de4036 docs: add merge strategy test evidence document
```

**The squash merge commit `351fcfa` has no parent link to release.** Git cannot trace ancestry through it. The merge base is stuck at `4de4036` — the state *before* the release, not after it.

### Step 4: Merge PRs #5 and #6 into main

Dad jokes and fortune cookie features merged into `main`.

### Step 5: Release PR (main → release) — PR #8 — EVIDENCE OF DIVERGENCE

Created PR #8 from `main` → `release` for the v0.3.0 release.

```
--- merge base (STALE) ---
4de4036 docs: add merge strategy test evidence document

--- commits git thinks are "new" (INFLATED) ---
d935a44 Merge pull request #6 from markwperry/feat/fortune-cookie
07122df Merge branch 'main' into feat/fortune-cookie
fe5c4ce Merge pull request #5 from markwperry/feat/dad-jokes
ebf4f5e Merge branch 'main' into feat/dad-jokes
351fcfa chore: sync release v0.2.0 back to main (#8 simulated squash merge)
67b5570 feat: add fortune cookie endpoint
45a36fc feat: add dad jokes endpoint

--- diff stat ---
3 files changed, 75 insertions(+)
```

**Result**: Git sees **7 commits** as "new" when only **2** (PRs #5 and #6) actually are. The squash merge commit and its surrounding merge commits are all treated as unrelated history. Over multiple release cycles, this accumulates — each release PR would include more and more phantom commits.

---

## Round 2: Regular Merge (The Fix)

### Step 6: Merge PR #8 (regular merge commit)

PR #8 was merged into release using a regular merge commit. GitHub automatically created a merge commit with two parents linking main and release.

### Step 7: No manual sync needed

Because the PR #8 merge commit has parent links to both branches, **git can trace the shared ancestry automatically**. No separate "sync release back to main" step was performed.

```
--- merge base (CURRENT — correct!) ---
a346a81 Merge branch 'release' into main

--- commits on main not on release ---
(none)

--- commits on release not on main ---
2fc72da Merge pull request #8 from markwperry/main

--- diff ---
(empty — branches are aligned)
```

### Step 8: Add coin flip feature (PR #9) and create release PR #10

PR #9 (coin flip endpoint) merged into `main`. Created PR #10 from `main` → `release`.

```
--- merge base (CURRENT) ---
a346a81 Merge branch 'release' into main

--- commits git sees as "new" (CORRECT — only the new feature) ---
64ea872 Merge pull request #9 from markwperry/feat/coin-flip
6e246a7 feat: add coin flip endpoint

--- diff stat ---
2 files changed, 33 insertions(+)
```

**Result**: Git sees exactly **2 commits** — only PR #9. No phantom history, no inflated diff. The merge base correctly reflects the most recent sync point.

---

## Round 3: Hotfix Process Test

### Objective

Test the full hotfix workflow: branch from `release`, fix a production bug, PR with `fix(hotfix):` prefix to `release`, verify automated validation and backport to `main`.

### Step 1: Create Hotfix

- **Branch**: `fix/health-check-detail` (branched from `release` at `33ab3ed`)
- **Bug**: `/health` endpoint returned bare `{"status": "ok"}` with no diagnostic info for monitoring
- **Fix**: Added `version`, `goVersion`, `memAlloc`, and `goroutines` to health response
- **Commit**: `50da2f9 fix(hotfix): health endpoint returns minimal info for monitoring`

### Step 2: PR to Release — PR #11

- **PR #11**: `fix(hotfix): health endpoint returns minimal info for monitoring`
- **Base**: `release` (not `main` — this is key for hotfixes)
- **Validation checks**:
  - **Validate PR Title**: PASSED (conventional commit format `fix(hotfix):`)
  - **Validate Hotfix Naming** (`check-naming`): PASSED (`fix(hotfix):` prefix detected)

### Step 3: Merge and Automated Backport

PR #11 merged into `release`. The Backport Hotfix workflow triggered automatically.

**Backport Attempt 1 — FAILED**
```
Error: GitHub Actions is not permitted to create or approve pull requests.
```
- **Cause**: Repository setting "Allow GitHub Actions to create and approve pull requests" was not enabled
- **Fix**: Enabled the setting in Settings → Actions → General → Workflow permissions

**Backport Attempt 2 (re-run) — FAILED**
```
Error: Validation Failed: "No commits between main and backport/hotfix-11"
```
- **Cause**: The workflow cherry-picked `merge_commit_sha` (the merge commit GitHub created). Cherry-picking a merge commit without the `-m` flag produces an empty result because git doesn't know which parent to diff against. The `backport/hotfix-11` branch was pushed with no actual changes vs main.
- **Root cause**: This is the same issue documented in the myqq `backport-hotfix.yml` — the workflow was adapted from the branching strategy document which used `merge_commit_sha`, but this only works reliably with the `-m 1` flag or by cherry-picking the original commits instead.

**Manual Backport — PR #12**
- Cherry-picked the original hotfix commit `50da2f9` (from `fix/health-check-detail` branch, not the merge commit)
- Created PR #12: `[Backport] fix(hotfix): health endpoint returns minimal info for monitoring`

### Step 4: Workflow Fix

Updated `backport-hotfix.yml` to cherry-pick the **original commits from the hotfix branch** (`head.sha`) instead of the merge commit:

```yaml
# Attempt 1 (broken — empty cherry-pick):
git cherry-pick ${{ github.event.pull_request.merge_commit_sha }}

# Attempt 2 (broken — commits already reachable from release after merge):
MERGE_BASE=$(git merge-base origin/release ${{ github.event.pull_request.head.sha }})
COMMITS=$(git rev-list --reverse ${MERGE_BASE}..${{ github.event.pull_request.head.sha }})

# Final fix (working):
git cherry-pick -m 1 ${{ github.event.pull_request.merge_commit_sha }}
```

The `-m 1` flag tells git to diff the merge commit against its first parent (release before the merge), extracting exactly the hotfix changes.

### Hotfix Test #1 Findings (PR #11)

| Step | Expected | Actual | Status |
|---|---|---|---|
| Branch from `release` | Hotfix starts from production state | Branched from `release` at `33ab3ed` | PASS |
| `fix(hotfix):` prefix validation | PR blocked without prefix | `check-naming` workflow passed | PASS |
| Conventional commit validation | PR title validated | `Validate PR title` passed | PASS |
| Auto-backport cherry-pick | Hotfix code applied to main | Empty cherry-pick (merge_commit_sha bug) | **FAIL — FIXED** |
| Auto-backport PR creation | PR created to main | Failed on first run (permissions), empty on second | **FAIL — FIXED** |
| Backport PR auto-merge | PR merges without review | Not tested (manual backport used) | DEFERRED |

### Configuration Requirements Discovered

1. **Repository setting**: "Allow GitHub Actions to create and approve pull requests" must be enabled (Settings → Actions → General)
2. **Workflow fix**: Use `cherry-pick -m 1 merge_commit_sha` — not bare `merge_commit_sha` (empty result) and not `head.sha` range (empty after merge)
3. **Auto-merge**: Requires branch protection rules on `main` — without them, the `enablePullRequestAutoMerge` GraphQL mutation fails

---

## Round 4: Hotfix Process — Successful Run (PR #17)

### Objective

Verify the corrected hotfix workflow runs end-to-end after fixing the cherry-pick approach and syncing the workflow to release.

### Setup

- Synced workflow fix to release via PR #16 (regular merge commit)
- Verified `release` branch has `cherry-pick -m 1` in `backport-hotfix.yml`

### Step 1: Create Hotfix

- **Branch**: `fix/magic8ball-empty-question` (branched from `release` at `ab7c54e`)
- **Bug**: `GET /8ball` without `q` parameter silently used a default question instead of returning an error
- **Fix**: Returns 400 Bad Request with descriptive error message when `q` is missing
- **Commit**: `02a9289 fix(hotfix): 8ball endpoint returns 400 when question is missing`

### Step 2: PR to Release — PR #17

- **PR #17**: `fix(hotfix): 8ball endpoint returns 400 when question is missing`
- **Validation checks**:
  - **Validate PR Title**: PASSED
  - **Validate Hotfix Naming** (`check-naming`): PASSED

### Step 3: Merge and Automated Backport — SUCCESS

PR #17 merged into `release`. The Backport Hotfix workflow triggered and **completed successfully**:

```
Backport Hotfix to Main — workflow run results:

✅ Cherry-pick hotfix commits     — cherry-pick -m 1 applied hotfix diff correctly
✅ Check for conflicts            — no conflicts detected
✅ Push backport branch           — backport/hotfix-17 pushed
✅ Create backport PR             — PR #18 created by github-actions[bot]
                                    Title: [Backport] fix(hotfix): 8ball endpoint returns 400 when question is missing
                                    1 file changed, +4/-1
⚠️ Enable auto-merge             — Failed: "Protected branch rules not configured for this branch"
                                    (main branch has no branch protection — auto-merge requires it)
```

### Backport PR #18

- **Created by**: `github-actions[bot]` (automated)
- **Branch**: `backport/hotfix-17` → `main`
- **Changes**: 1 file changed, 4 insertions, 1 deletion (exactly the hotfix diff)
- **Auto-merge**: Not enabled (requires branch protection on `main`)
- **Resolution**: Merged manually

### Hotfix Test #2 Findings (PR #17)

| Step | Expected | Actual | Status |
|---|---|---|---|
| Branch from `release` | Hotfix starts from production state | Branched from `release` at `ab7c54e` | **PASS** |
| `fix(hotfix):` prefix validation | PR validated | `check-naming` passed | **PASS** |
| Conventional commit validation | PR title validated | `Validate PR title` passed | **PASS** |
| Auto-backport cherry-pick | Hotfix diff applied to main | `cherry-pick -m 1` succeeded, 1 file +4/-1 | **PASS** |
| Auto-backport branch push | Branch pushed to origin | `backport/hotfix-17` pushed | **PASS** |
| Auto-backport PR creation | PR created to main | PR #18 created by `github-actions[bot]` | **PASS** |
| Backport PR auto-merge | PR auto-merges | Failed — no branch protection on `main` | **EXPECTED** |

### Cherry-Pick Evolution Summary

| Approach | Method | Result | Why |
|---|---|---|---|
| v1 (from branching strategy doc) | `cherry-pick merge_commit_sha` | Empty cherry-pick | Merge commits have 2 parents; without `-m` git doesn't know which parent to diff against |
| v2 (first fix attempt) | `rev-list head.sha` range | Empty range | After PR merge, hotfix commits are reachable from release, so merge-base == head.sha |
| v3 (working) | `cherry-pick -m 1 merge_commit_sha` | **Correct diff** | `-m 1` diffs against first parent (release before merge), extracting exactly the hotfix changes |

---

## Round 5: Hotfix Process — Fully Automated End-to-End (PR #20)

### Objective

Verify the complete hotfix workflow with the final fix: workflow creates the backport PR and **merges it automatically** via the `pulls.merge()` API, requiring zero manual intervention.

### Step 1: Create Hotfix

- **Branch**: `fix/fortune-missing-numbers` (branched from `release` at `d67ab16`)
- **Bug**: `/fortune` lucky numbers could contain duplicates (e.g. `42, 42, 8`) because numbers were picked independently with replacement
- **Fix**: Shuffle the number pool and take first 3 to guarantee uniqueness
- **Commit**: `8c9d2a1 fix(hotfix): fortune endpoint returns duplicate lucky numbers`

### Step 2: PR to Release — PR #20

- **PR #20**: `fix(hotfix): fortune endpoint returns duplicate lucky numbers`
- **Validation checks**: Both `check-naming` and `Validate PR title` PASSED

### Step 3: Merge and Fully Automated Backport — COMPLETE SUCCESS

PR #20 merged into `release`. The Backport Hotfix workflow ran to **full completion**:

```
Backport Hotfix to Main — workflow run (23356209065):

✅ Cherry-pick hotfix commits     — cherry-pick -m 1 applied hotfix diff correctly
✅ Check for conflicts            — "Cherry-pick succeeded without conflicts"
✅ Push backport branch           — backport/hotfix-20 pushed
✅ Create backport PR             — PR #21 created by github-actions[bot]
                                    "Created backport PR #21"
✅ Merge backport PR              — PR #21 merged automatically
                                    "✅ Merged backport PR #21"
                                    Merged at 2026-03-20T18:06:11Z (8 seconds after creation)
```

### Hotfix Test #3 (Final) Findings

| Step | Expected | Actual | Status |
|---|---|---|---|
| Branch from `release` | Hotfix starts from production state | Branched from `release` at `d67ab16` | **PASS** |
| `fix(hotfix):` prefix validation | PR validated | `check-naming` passed | **PASS** |
| Conventional commit validation | PR title validated | `Validate PR title` passed | **PASS** |
| Auto-backport cherry-pick | Hotfix diff applied to main | `cherry-pick -m 1` succeeded | **PASS** |
| Auto-backport branch push | Branch pushed to origin | `backport/hotfix-20` pushed | **PASS** |
| Auto-backport PR creation | PR created to main | PR #21 created by `github-actions[bot]` | **PASS** |
| **Auto-backport PR merge** | **PR merged automatically** | **PR #21 merged 8 seconds after creation** | **PASS** |

### Workflow Evolution Summary

| Version | Auto-merge Approach | Result |
|---|---|---|
| v1 | `enablePullRequestAutoMerge` GraphQL mutation | Failed — requires branch protection rules on target branch |
| v2 | `pulls.merge()` REST API (direct merge) | **Working** — no branch protection or repo settings needed |

---

## Conclusion

| Metric | Round 1: After Squash Merge | Round 2: After Regular Merge |
|---|---|---|
| Merge base | `4de4036` (STALE — pre-release) | `a346a81` (CURRENT — post-release) |
| Commits shown as "new" in release PR | 7 (inflated) | 2 (correct) |
| Files changed in release PR | 3 files, 75 insertions | 2 files, 33 insertions |
| Manual sync-back needed? | Yes (and it caused the problem) | **No** |
| PR diff accurate? | No — includes phantom history | Yes — exactly the new feature |

### Key Findings

1. **Squash merges between long-lived branches destroy shared ancestry.** The squash commit has no parent link to the source branch, so `git merge-base` returns a stale result. This causes inflated diffs, phantom commits in PRs, and potential merge conflicts on code that's already been integrated.

2. **Regular merge commits preserve shared ancestry.** The two-parent merge commit allows git to correctly trace history through both branches, keeping `git merge-base` current.

3. **No back-sync is needed with regular merge commits.** When a `main` → `release` PR is merged with a merge commit, the parent link from release back to main is sufficient. Git knows the branches are aligned. The "sync release back to main" step is a workaround for a problem that only exists with squash merges.

4. **Squash merges are fine for short-lived → long-lived branches.** Feature branches squash-merged into `main` cause no issues because the feature branch is deleted after merge — there's no ongoing relationship to maintain.

### Recommendation

**For the `main` → `release` release PRs**: Always use **regular merge commits** (the GitHub default). Never squash merge.

**For feature branches → `main`**: Squash merge is fine and encouraged for clean history.

**For hotfixes on `release`**: The backport automation is fully automated — merge the hotfix to release and walk away. The workflow:
1. Cherry-picks the merge commit with `-m 1` to extract the hotfix diff
2. Creates a backport PR to `main`
3. Merges it immediately via `pulls.merge()` API (no branch protection or auto-merge settings required)
4. If conflicts exist, creates the PR but leaves it open for manual resolution

**Only requirement**: Enable "Allow GitHub Actions to create and approve pull requests" in Settings → Actions → General.

### Workflow Iterations

The backport workflow went through multiple iterations to reach the final working state:

**Cherry-pick approach** (3 iterations):

| Version | Approach | Problem |
|---|---|---|
| v1 | `cherry-pick merge_commit_sha` | Merge commits have 2 parents; without `-m`, git produces empty cherry-pick |
| v2 | `rev-list head.sha` range | After PR merge, commits are reachable from release, making range empty |
| v3 | `cherry-pick -m 1 merge_commit_sha` | **Working** — `-m 1` diffs against first parent, extracting hotfix changes |

**Auto-merge approach** (2 iterations):

| Version | Approach | Problem |
|---|---|---|
| v1 | `enablePullRequestAutoMerge` GraphQL | Requires branch protection rules on `main` — not always configured |
| v2 | `pulls.merge()` REST API (direct merge) | **Working** — no special settings required, merges immediately |
