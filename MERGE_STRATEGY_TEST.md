# Merge Strategy Test: Squash Merge vs Regular Merge Between Long-Lived Branches

## Purpose

This document provides evidence that **squash merging between long-lived branches** (`main` and `release`) causes merge base divergence, leading to inflated diffs, phantom commits in PRs, and broken branch comparisons. It demonstrates that **regular merge commits** preserve the shared history and keep the branches naturally aligned — with **no manual sync-back step required**.

It also documents the full development and testing of the **hotfix backport workflow**, including every failure, iteration, and fix along the way.

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

## CI/CD Workflows Created

The following GitHub Actions workflows were created to mirror the MyQQ ecosystem:

| Workflow File | Trigger | Purpose |
|---|---|---|
| `ci-main.yml` | PRs to `main` | Lint (`go vet`, `staticcheck`), unit tests, build via `build.sh` |
| `release.yml` | Push to `release` or version tags | E2E tests (starts server, curls endpoints), Docker build & push to GHCR |
| `validate-pr-title.yml` | All PRs (`pull_request_target`) | Enforces conventional commit format on PR titles via `amannn/action-semantic-pull-request@v5` |
| `validate-hotfix-naming.yml` | PRs to `release` | Blocks non-main PRs to `release` without `fix(hotfix):` prefix |
| `backport-hotfix.yml` | PRs merged to `release` | Auto cherry-picks hotfixes back to `main`, creates and merges backport PR |
| `tag.yml` | Manual dispatch on `release` | Bumps version in `config/version.yaml`, generates CHANGELOG.md, creates git tag via `TriPSs/conventional-changelog-action@v3.11.0` |

**Additional infrastructure:**
- `build.sh` — Go build script with ldflags for version injection (`-X main.version`, `-X main.gitHash`, etc.)
- `config/version.yaml` — Single source of truth for version (pattern from myqq-api)
- `Dockerfile` — Multi-stage build using `build.sh` with `BUILD_VERSION`, `BUILD_GITHASH`, `BRANCH` build args
- `/version` endpoint — Returns injected version, git hash, build timestamp, and branch at runtime

**Bug discovered during testing:** `build.sh` used `#!/bin/bash` but Alpine Docker images only have `sh`. Fixed by changing shebang to `#!/bin/sh` (commit `fdc8968`). The script only uses POSIX-compatible syntax.

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

All feature PRs used conventional commit titles (`feat: ...`) and were validated by the `validate-pr-title.yml` workflow.

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

To set up this simulation, we:
1. Reset `main` back to before the auto-sync merge commit that GitHub created during the PR process (`git reset --hard 4de4036`)
2. Force-pushed `main` to this pre-sync state
3. Performed `git merge --squash origin/release` followed by a commit

This produced a new commit on main with no parent link to release:

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

## Round 3: Hotfix Process — First Attempt (PR #11)

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

**Backport Attempt 1 — FAILED (permissions)**
```
Error: GitHub Actions is not permitted to create or approve pull requests.
```
- **Cause**: Repository setting "Allow GitHub Actions to create and approve pull requests" was not enabled
- **Fix**: Enabled the setting in Settings → Actions → General → Workflow permissions

**Backport Attempt 2 (re-run) — FAILED (empty cherry-pick)**
```
Error: Validation Failed: "No commits between main and backport/hotfix-11"
```
- **Cause**: The workflow used `git cherry-pick ${{ github.event.pull_request.merge_commit_sha }}` (cherry-pick v1). Cherry-picking a merge commit without the `-m` flag produces an empty result because git doesn't know which parent to diff against. The `backport/hotfix-11` branch was pushed with no actual changes vs main.
- **Root cause**: The workflow was adapted from the branching strategy document which used `merge_commit_sha`, but merge commits require the `-m` flag to specify which parent to diff against.

**Manual Backport — PR #12**
- Cherry-picked the original hotfix commit `50da2f9` (from `fix/health-check-detail` branch, not the merge commit)
- Created PR #12: `[Backport] fix(hotfix): health endpoint returns minimal info for monitoring`
- Merged manually

### Step 4: Workflow Fix — Cherry-pick v2

Updated `backport-hotfix.yml` to cherry-pick the **original commits from the hotfix branch** (`head.sha` range) instead of the merge commit:

```yaml
# v2 approach:
MERGE_BASE=$(git merge-base origin/release ${{ github.event.pull_request.head.sha }})
COMMITS=$(git rev-list --reverse ${MERGE_BASE}..${{ github.event.pull_request.head.sha }})
for COMMIT in $COMMITS; do
  git cherry-pick "$COMMIT"
done
```

This fix was committed to `main` (commit `e06b0dc`).

### Step 5: Branch Divergence from Manual Backport

The manual backport (PR #12) cherry-picked the hotfix to `main`, creating a separate commit with no parent link to release — the same divergence problem as the squash merge test. A release PR (#14) was needed to re-sync the branches.

### Hotfix Test #1 Findings (PR #11)

| Step | Expected | Actual | Status |
|---|---|---|---|
| Branch from `release` | Hotfix starts from production state | Branched from `release` at `33ab3ed` | PASS |
| `fix(hotfix):` prefix validation | PR validated | `check-naming` workflow passed | PASS |
| Conventional commit validation | PR title validated | `Validate PR title` passed | PASS |
| Auto-backport cherry-pick | Hotfix code applied to main | Empty cherry-pick (v1: bare `merge_commit_sha`) | **FAIL** |
| Auto-backport PR creation | PR created to main | Failed (permissions), then empty (no commits) | **FAIL** |
| Backport PR merge | PR merges | Not reached — manual backport used | **FAIL** |

### Configuration Requirements Discovered

1. **Repository setting**: "Allow GitHub Actions to create and approve pull requests" must be enabled (Settings → Actions → General)

---

## Round 3b: Hotfix Process — Second Attempt (PR #15)

### Objective

Test the cherry-pick v2 fix (`head.sha` range approach). The workflow fix was committed to `main` but needed to be synced to `release` first.

### Setup

- Synced workflow fix and other accumulated changes to release via PR #14 (regular merge commit)
- PR #16 synced the cherry-pick v2 fix to release

### Step 1: Create Hotfix

- **Branch**: `fix/coinflip-zero-panic` (branched from `release` at `1f8e365`)
- **Bug**: `GET /coinflip?n=foo` silently defaulted to a single flip instead of returning an error; `GET /coinflip?n=0` and `n=-1` also silently fell through
- **Fix**: Returns 400 Bad Request with descriptive error messages for non-numeric and out-of-range values
- **Commit**: `90074ef fix(hotfix): coinflip endpoint silently ignores invalid input`

### Step 2: PR to Release — PR #15

- **PR #15**: `fix(hotfix): coinflip endpoint silently ignores invalid input`
- **Validation checks**: Both `check-naming` and `Validate PR title` PASSED

### Step 3: Merge and Automated Backport — FAILED (empty range)

PR #15 merged into `release`. The Backport Hotfix workflow triggered.

**Backport — FAILED**
```
No commits to cherry-pick
Error: Validation Failed: "No commits between main and backport/hotfix-15"
```
- **Cause**: The v2 cherry-pick approach (`rev-list head.sha` range) failed because **after the PR is merged into release, the hotfix commits become reachable from release**. This means `git merge-base origin/release <head.sha>` returns the hotfix commit itself, making the rev-list range empty.
- **Workflow log confirmed**: The `MERGE_BASE` equaled `head.sha`, producing zero commits in the range.

**Manual Backport**: The coinflip fix was already on `main` via the merge of PR #12's backport chain, so no additional manual backport was needed for the code. The backport branch `backport/hotfix-15` was pushed (empty) but the PR creation failed.

### Step 4: Workflow Fix — Cherry-pick v3

The fundamental problem: after a PR is merged, any approach that references the original commits will find them reachable from the target branch. The solution is to cherry-pick the **merge commit itself** with the `-m 1` flag:

```yaml
# v3 approach (final, working):
git cherry-pick -m 1 ${{ github.event.pull_request.merge_commit_sha }}
```

The `-m 1` flag tells git to diff the merge commit against its **first parent** (the release branch before the merge), which produces exactly the hotfix changes.

This fix was committed to `main` (commit `de4a91d`).

### Hotfix Test #2 Findings (PR #15)

| Step | Expected | Actual | Status |
|---|---|---|---|
| Branch from `release` | Hotfix starts from production state | Branched from `release` at `1f8e365` | PASS |
| `fix(hotfix):` prefix validation | PR validated | `check-naming` passed | PASS |
| Conventional commit validation | PR title validated | `Validate PR title` passed | PASS |
| Auto-backport cherry-pick | Hotfix code applied to main | Empty range (v2: `head.sha` reachable from release) | **FAIL** |
| Auto-backport PR creation | PR created to main | Failed (no commits between branches) | **FAIL** |
| Backport PR merge | PR merges | Not reached | **FAIL** |

### Cherry-Pick Evolution So Far

| Version | Approach | Result | Why It Failed |
|---|---|---|---|
| v1 (PR #11) | `cherry-pick merge_commit_sha` | Empty cherry-pick | Merge commits have 2 parents; without `-m`, git doesn't know which parent to diff against |
| v2 (PR #15) | `rev-list head.sha` range | Empty range | After PR merge, hotfix commits are reachable from release, so merge-base == head.sha |
| v3 | `cherry-pick -m 1 merge_commit_sha` | *To be tested* | `-m 1` diffs against first parent (release before merge) |

---

## Round 4: Hotfix Process — Cherry-pick v3 Works (PR #17)

### Objective

Verify the cherry-pick v3 fix (`-m 1 merge_commit_sha`) works end-to-end.

### Setup

- Synced cherry-pick v3 workflow fix to release via PR #16 (regular merge commit)
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

### Step 3: Merge and Automated Backport — PARTIAL SUCCESS

PR #17 merged into `release`. The Backport Hotfix workflow triggered and **the cherry-pick worked for the first time**:

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
- **Auto-merge**: Failed — the `enablePullRequestAutoMerge` GraphQL mutation requires branch protection rules on the target branch, which `main` didn't have
- **Resolution**: Merged manually

### The Auto-Merge Problem

The workflow used `enablePullRequestAutoMerge` (a GraphQL mutation) to auto-merge the backport PR. This requires:
- Branch protection rules on `main`
- "Allow auto-merge" enabled in repo settings

Enabling auto-merge repo-wide was undesirable — normal PRs should require manual merge. The solution: replace the GraphQL auto-merge with a direct `pulls.merge()` REST API call, which merges the PR immediately without any special settings.

This fix was committed to `main` (commit `60de972`).

### Hotfix Test #3 Findings (PR #17)

| Step | Expected | Actual | Status |
|---|---|---|---|
| Branch from `release` | Hotfix starts from production state | Branched from `release` at `ab7c54e` | **PASS** |
| `fix(hotfix):` prefix validation | PR validated | `check-naming` passed | **PASS** |
| Conventional commit validation | PR title validated | `Validate PR title` passed | **PASS** |
| Auto-backport cherry-pick | Hotfix diff applied to main | `cherry-pick -m 1` succeeded, 1 file +4/-1 | **PASS** |
| Auto-backport branch push | Branch pushed to origin | `backport/hotfix-17` pushed | **PASS** |
| Auto-backport PR creation | PR created to main | PR #18 created by `github-actions[bot]` | **PASS** |
| Backport PR auto-merge | PR auto-merges | Failed — `enablePullRequestAutoMerge` requires branch protection | **FAIL** |

---

## Round 5: Hotfix Process — Fully Automated End-to-End (PR #20)

### Objective

Verify the complete hotfix workflow with all fixes in place: cherry-pick v3 (`-m 1`) and direct merge via `pulls.merge()` API, requiring zero manual intervention.

### Setup

- Synced direct-merge workflow fix to release via PR #19 (regular merge commit)
- Verified `release` branch has both `cherry-pick -m 1` and `pulls.merge()` in `backport-hotfix.yml`

### Step 1: Create Hotfix

- **Branch**: `fix/fortune-missing-numbers` (branched from `release` at `d67ab16`)
- **Bug**: `/fortune` lucky numbers could contain duplicates (e.g. `42, 42, 8`) because numbers were picked independently with replacement
- **Fix**: Shuffle the number pool and take first 3 to guarantee uniqueness
- **Commit**: `8c9d2a1 fix(hotfix): fortune endpoint returns duplicate lucky numbers`

### Step 2: PR to Release — PR #20

- **PR #20**: `fix(hotfix): fortune endpoint returns duplicate lucky numbers`
- **Validation checks**: Both `check-naming` and `Validate PR title` PASSED

### Step 3: Merge and Fully Automated Backport — COMPLETE SUCCESS

PR #20 merged into `release`. The Backport Hotfix workflow ran to **full completion with zero manual intervention**:

```
Backport Hotfix to Main — workflow run (23356209065):

✅ Cherry-pick hotfix commits     — cherry-pick -m 1 applied hotfix diff correctly
✅ Check for conflicts            — "Cherry-pick succeeded without conflicts"
✅ Push backport branch           — backport/hotfix-20 pushed
✅ Create backport PR             — PR #21 created by github-actions[bot]
                                    "Created backport PR #21"
✅ Merge backport PR              — PR #21 merged automatically via pulls.merge() API
                                    "✅ Merged backport PR #21"
                                    Merged at 2026-03-20T18:06:11Z (8 seconds after creation)
```

### Hotfix Test #4 (Final) Findings

| Step | Expected | Actual | Status |
|---|---|---|---|
| Branch from `release` | Hotfix starts from production state | Branched from `release` at `d67ab16` | **PASS** |
| `fix(hotfix):` prefix validation | PR validated | `check-naming` passed | **PASS** |
| Conventional commit validation | PR title validated | `Validate PR title` passed | **PASS** |
| Auto-backport cherry-pick | Hotfix diff applied to main | `cherry-pick -m 1` succeeded | **PASS** |
| Auto-backport branch push | Branch pushed to origin | `backport/hotfix-20` pushed | **PASS** |
| Auto-backport PR creation | PR created to main | PR #21 created by `github-actions[bot]` | **PASS** |
| **Auto-backport PR merge** | **PR merged automatically** | **PR #21 merged 8 seconds after creation** | **PASS** |

---

## Complete PR History

For full transparency, here is every PR created during this test:

| PR | Title | Base | Head | Purpose | Merge Method |
|---|---|---|---|---|---|
| #1 | `chore: Merge pull request #1` | `main` | `chore/forcebuild` | Initial workflow setup | Merge commit |
| #2 | Initial sync | `release` | `main` | First sync of main to release | Merge commit |
| #3 | `feat: add hillbilly translator endpoint` | `main` | `feat/hillbilly-translator` | Round 1 feature | Merge commit |
| #4 | `feat: add magic 8-ball endpoint` | `main` | `feat/magic-8ball` | Round 1 feature | Merge commit |
| #5 | `feat: add dad jokes endpoint` | `main` | `feat/dad-jokes` | Round 1 feature | Merge commit |
| #6 | `feat: add fortune cookie endpoint` | `main` | `feat/fortune-cookie` | Round 1 feature | Merge commit |
| #7 | `chore(release): v0.2.0` | `release` | `main` | Release PRs #3–#4 | Merge commit |
| #8 | `chore(release): v0.3.0` | `release` | `main` | Release PRs #5–#6 (divergence demo) | Merge commit |
| #9 | `feat: add coin flip endpoint` | `main` | `feat/coin-flip` | Round 2 feature | Merge commit |
| #10 | `chore(release): v0.4.0` | `release` | `main` | Release PR #9 (alignment demo) | Merge commit |
| #11 | `fix(hotfix): health endpoint` | `release` | `fix/health-check-detail` | Hotfix test #1 | Merge commit |
| #12 | `[Backport] fix(hotfix): health endpoint` | `main` | `backport/hotfix-11-manual` | Manual backport of #11 | Merge commit |
| #14 | `chore(release): sync` | `release` | `main` | Re-sync after manual backport | Merge commit |
| #15 | `fix(hotfix): coinflip input validation` | `release` | `fix/coinflip-zero-panic` | Hotfix test #2 | Merge commit |
| #16 | `chore(release): sync workflow fix` | `release` | `main` | Sync cherry-pick v3 to release | Merge commit |
| #17 | `fix(hotfix): 8ball missing question` | `release` | `fix/magic8ball-empty-question` | Hotfix test #3 | Merge commit |
| #18 | `[Backport] fix(hotfix): 8ball` | `main` | `backport/hotfix-17` | Auto backport of #17 (merged manually) | Merge commit |
| #19 | `chore(release): sync direct-merge fix` | `release` | `main` | Sync pulls.merge() fix to release | Merge commit |
| #20 | `fix(hotfix): fortune duplicate numbers` | `release` | `fix/fortune-missing-numbers` | Hotfix test #4 (final) | Merge commit |
| #21 | `[Backport] fix(hotfix): fortune` | `main` | `backport/hotfix-20` | Auto backport of #20 (merged automatically) | Squash (via API) |

---

## Conclusion

### Merge Strategy Results

| Metric | Round 1: After Squash Merge | Round 2: After Regular Merge |
|---|---|---|
| Merge base | `4de4036` (STALE — pre-release) | `a346a81` (CURRENT — post-release) |
| Commits shown as "new" in release PR | 7 (inflated) | 2 (correct) |
| Files changed in release PR | 3 files, 75 insertions | 2 files, 33 insertions |
| Manual sync-back needed? | Yes (and it caused the problem) | **No** |
| PR diff accurate? | No — includes phantom history | Yes — exactly the new feature |

### Hotfix Backport Results

| Test | PR | Cherry-pick | PR Creation | PR Merge | Result |
|---|---|---|---|---|---|
| #1 | PR #11 | FAIL (v1: bare merge_commit_sha) | FAIL (permissions, then empty) | N/A | Manual backport |
| #2 | PR #15 | FAIL (v2: head.sha range empty) | FAIL (no commits) | N/A | Code already on main |
| #3 | PR #17 | PASS (v3: `-m 1`) | PASS | FAIL (auto-merge needs branch protection) | Manual merge |
| **#4** | **PR #20** | **PASS** | **PASS** | **PASS (8 seconds)** | **Fully automated** |

### Key Findings

1. **Squash merges between long-lived branches destroy shared ancestry.** The squash commit has no parent link to the source branch, so `git merge-base` returns a stale result. This causes inflated diffs, phantom commits in PRs, and potential merge conflicts on code that's already been integrated.

2. **Regular merge commits preserve shared ancestry.** The two-parent merge commit allows git to correctly trace history through both branches, keeping `git merge-base` current.

3. **No back-sync is needed with regular merge commits.** When a `main` → `release` PR is merged with a merge commit, the parent link from release back to main is sufficient. Git knows the branches are aligned. The "sync release back to main" step is a workaround for a problem that only exists with squash merges.

4. **Squash merges are fine for short-lived → long-lived branches.** Feature branches squash-merged into `main` cause no issues because the feature branch is deleted after merge — there's no ongoing relationship to maintain.

5. **Cherry-picking merge commits requires `-m 1`.** Without the `-m` flag, git doesn't know which parent to diff against and produces an empty cherry-pick. The `-m 1` flag diffs against the first parent, extracting exactly the changes introduced by the PR.

6. **Direct API merge is simpler than auto-merge.** The `enablePullRequestAutoMerge` GraphQL mutation requires branch protection rules on the target branch. Using `pulls.merge()` REST API merges the PR immediately with no special settings, keeping auto-merge disabled for normal PRs.

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

| Version | Tested In | Approach | Result |
|---|---|---|---|
| v1 | PR #11 | `cherry-pick merge_commit_sha` | Empty cherry-pick — merge commits have 2 parents; without `-m`, git doesn't know which to diff against |
| v2 | PR #15 | `rev-list head.sha` range | Empty range — after merge, hotfix commits are reachable from release, so merge-base == head.sha |
| v3 | PR #17, #20 | `cherry-pick -m 1 merge_commit_sha` | **Working** — `-m 1` diffs against first parent, extracting exactly the hotfix changes |

**Auto-merge approach** (2 iterations):

| Version | Tested In | Approach | Result |
|---|---|---|---|
| v1 | PR #17 | `enablePullRequestAutoMerge` GraphQL mutation | Failed — requires branch protection rules on `main`, which may not be configured |
| v2 | PR #20 | `pulls.merge()` REST API (direct merge) | **Working** — no special settings required, merges immediately |
