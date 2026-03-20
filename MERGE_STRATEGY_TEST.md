# Merge Strategy Test: Squash Merge vs Regular Merge Between Long-Lived Branches

## Purpose

This document provides evidence that **squash merging between long-lived branches** (`main` and `release`) causes merge base divergence, leading to inflated diffs, phantom conflicts, and broken PR comparisons. It demonstrates that **regular merge commits** preserve the shared history and keep the branches aligned.

## Background

In the simplified branching strategy used by MyQQ:
```
feature branches → main → release
```

The `release` branch is the validated, production-ready state. Periodically, `main` is merged into `release` via a PR to create a release. The question is: **how should release be synced back to main?**

- **Squash merge**: Creates a single new commit on main with the combined changes. This commit has **no parent link** back to release, so git loses track of the shared history.
- **Regular merge commit**: Creates a merge commit that has **two parents** — one on main, one on release. Git can trace the shared ancestry, keeping the branches aligned.

## Test Environment

- **Repository**: https://github.com/markwperry/TestWorkflows
- **Project**: Go API (Gin framework) with fun endpoints
- **Date**: 2026-03-20
- **Branches**: `main` (development), `release` (production)

---

## Phase 1: Feature Development (PRs #3–#6)

Four feature branches were created from `main`, each adding a fun endpoint:

| PR | Branch | Feature | Endpoint |
|---|---|---|---|
| #3 | `feat/hillbilly-translator` | Hillbilly text translator | `GET /hillbilly?text=...` |
| #4 | `feat/magic-8ball` | Magic 8-Ball | `GET /8ball?q=...` |
| #5 | `feat/dad-jokes` | Programming dad jokes | `GET /dadjoke` |
| #6 | `feat/fortune-cookie` | Fortune cookie messages | `GET /fortune` |

**PRs #3 and #4** were merged into `main` first (for Round 1).
**PRs #5 and #6** are held back for Round 2.

---

## Round 1: Squash Merge (The Problem)

### Step 1: Pre-Release State

```
--- main branch tip ---
e230e52 Merge pull request #4 from markwperry/feat/magic-8ball
bd905d5 Merge branch 'main' into feat/magic-8ball
3a4dab6 Merge pull request #3 from markwperry/feat/hillbilly-translator
cbdcfb9 feat: add magic 8-ball endpoint
19f61ff feat: add hillbilly translator endpoint
4dc5942 chore: Merge pull request #1 from markwperry/chore/forcebuild
afe35a4 feat: add conventional commits, versioning, and release tagging
15ff4b4 chore: fix backport hotfix
4712a8c chore: force build with commit
f8fa0b3 Initial commit: Go API with GitHub Actions workflow suite

--- release branch tip ---
f7dc8d2 Merge pull request #2 from markwperry/main
4dc5942 chore: Merge pull request #1 from markwperry/chore/forcebuild
afe35a4 feat: add conventional commits, versioning, and release tagging
15ff4b4 chore: fix backport hotfix
4712a8c chore: force build with commit

--- merge base ---
4dc5942 (shared ancestor of main and release)

--- commits on main not yet on release ---
e230e52 Merge pull request #4 from markwperry/feat/magic-8ball
bd905d5 Merge branch 'main' into feat/magic-8ball
3a4dab6 Merge pull request #3 from markwperry/feat/hillbilly-translator
cbdcfb9 feat: add magic 8-ball endpoint
19f61ff feat: add hillbilly translator endpoint
```

**Observation**: Merge base is `4dc5942`. Both branches share a clean common ancestor. The diff between main and release is exactly the 2 new features — as expected.

### Step 2: Release PR (main → release)

- **PR #7**: `chore(release): v0.2.0 — hillbilly translator & magic 8-ball`
- **Merge method**: Regular merge commit (this is correct — getting code to release)
- **Result**: _(to be captured after merge)_

### Step 3: Squash Merge release → main

After merging PR #7, we simulate what happens when someone squash-merges `release` back into `main`.

- **Method**: Squash merge
- **Result**: _(to be captured)_
- **New merge base**: _(to be captured)_

### Step 4: Merge PRs #5 and #6 into main

Add dad jokes and fortune cookie features to main.

### Step 5: Create Release PR (main → release) — EVIDENCE OF DIVERGENCE

- **PR**: _(to be created)_
- **Expected problem**: The diff will show ALL code from Round 1 again (hillbilly + 8-ball), not just the new features, because git no longer recognizes that main and release share that code via a common ancestor.
- **Merge base**: _(to be captured — will be the OLD merge base, not the post-release one)_
- **Diff stats**: _(to be captured)_

---

## Round 2: Regular Merge (The Fix)

### Step 6: Merge the release PR anyway

Merge the Round 2 release PR (with the inflated diff) to get release up to date.

### Step 7: Regular merge release → main

This time, use a **regular merge commit** to sync release back into main.

- **Method**: Regular merge commit (two parents)
- **Result**: _(to be captured)_
- **New merge base**: _(to be captured)_

### Step 8: Create Release PR (main → release) — EVIDENCE OF ALIGNMENT

- **PR**: _(to be created)_
- **Expected result**: The diff will be empty or minimal — branches are aligned because the regular merge commit preserved the shared ancestry.
- **Merge base**: _(to be captured — should be the latest release tip)_
- **Diff stats**: _(to be captured)_

---

## Conclusion

_(To be filled in after completing both rounds)_

| Metric | After Squash Merge | After Regular Merge |
|---|---|---|
| Merge base | _(old/stale)_ | _(current/correct)_ |
| PR diff includes old code? | Yes | No |
| Files changed in PR | _(inflated)_ | _(correct)_ |
| Lines changed in PR | _(inflated)_ | _(correct)_ |

**Recommendation**: When syncing between long-lived branches (`main` ↔ `release`), always use **regular merge commits**, never squash merges. Squash merges are fine for feature branches into main (short-lived → long-lived), but destructive for long-lived ↔ long-lived branch synchronization.
