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

**For hotfixes on `release`**: The existing backport automation (cherry-pick + auto-merge PR) correctly handles syncing hotfixes back to `main` without creating the squash merge problem.
