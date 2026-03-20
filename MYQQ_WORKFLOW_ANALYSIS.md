# MyQQ Workflow Analysis: Comparison with TestWorkflows Proven Patterns

## Purpose

This document compares the GitHub Actions workflows in the `myqq` repository against the patterns tested and proven in the `TestWorkflows` project on 2026-03-20. It identifies what works, what's broken, what changes after PR #2122 helped or hurt, and what it would take to make the myqq workflows function correctly.

## Context

- **PR #2122** (2026-02-23): Added dependabot lockfile automation, fixed hotfix syntax, renamed lint PR workflow, added concurrency control to test-pr
- **Subsequent PRs**: #2125, #2127, #2133, #2200, #2210, #2219 modified workflows further through 2026-03-18
- **TestWorkflows project**: Proven working patterns for backport, release, and hotfix workflows through 4 iterations of testing

---

## Workflow-by-Workflow Analysis

### 1. `validate-pr-title.yaml` — PR Title Validation

**MyQQ:**
```yaml
on:
  pull_request_target:
    types: [opened, edited, synchronize]
# Uses amannn/action-semantic-pull-request@v5
# Runs on: self-hosted
```

**TestWorkflows (proven):**
```yaml
on:
  pull_request_target:
    types: [opened, edited, synchronize]
# Uses amannn/action-semantic-pull-request@v5
# Runs on: ubuntu-latest
```

**Verdict: WORKS** — Identical logic. Only difference is `self-hosted` vs `ubuntu-latest` runner. No issues.

---

### 2. `validate-hotfix-naming.yml` — Hotfix PR Title Check

**MyQQ:**
```yaml
on:
  pull_request:
    branches: [release]
    types: [opened, synchronize, edited]
# Bash check: allows main PRs, requires fix(hotfix): for others
```

**TestWorkflows (proven):**
```yaml
# Identical logic
```

**Verdict: WORKS** — Identical to the proven version. No issues.

---

### 3. `test-pr.yaml` — PR Testing (PRs to main)

**MyQQ:**
```yaml
on:
  workflow_dispatch:
  pull_request:
    branches: [main]
    types: [opened, synchronize, reopened, ready_for_review]
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
# Runs: format-check, npm ci, unit tests
# Skips drafts
```

**Verdict: WORKS** — Standard PR testing. The concurrency control added in PR #2122 is a good improvement — cancels redundant runs when a PR is updated. No issues.

---

### 4. `test-release.yaml` — E2E Tests on Release PRs

**MyQQ:**
```yaml
on:
  workflow_dispatch:
  pull_request:
    branches: [release]
    types: [opened, synchronize, reopened, ready_for_review]
# 3 jobs: unit tests, build artifact, 3x parallel Cypress E2E
# Runs E2E on self-hosted runners
```

**Verdict: WORKS** — This is the key workflow that runs E2E tests on the integrated code before it reaches release. Correctly triggers on PRs to release. The 3-container parallel Cypress matrix is a reasonable approach for test splitting. No issues with the workflow logic itself.

**Note:** Added in PR #2125 ("Bind Release/Hotfix Workflows to Release Branch"). This was a **helpful** change — it ensures E2E tests run on release PRs specifically, which is the whole point of the branching strategy.

---

### 5. `test-main.yaml` — E2E Tests on Main (Manual)

**MyQQ:**
```yaml
on:
  workflow_dispatch:
# Same 3-job pattern as test-release.yaml but for main branch
# Only runs manually
```

**Verdict: WORKS, BUT QUESTIONABLE VALUE** — This is a manual-only E2E test for main. Added in PR #2125. It's not "complexity" — it's just a copy of test-release with `main` branch guards. The branching strategy says E2E should run on release, not main. This workflow exists as an escape hatch for ad-hoc validation.

**Assessment:** Not harmful, but not part of the core flow. Could be removed without affecting the branching strategy.

---

### 6. `prepare-release.yaml` — Release Preparation

**MyQQ:**
```yaml
on:
  workflow_dispatch:
    inputs:
      release_type: [patch, minor, major]
# Only runs on main
# Uses standard-version@9.0.0 to bump version + generate CHANGELOG
# Creates branch chore/release-vX.Y.Z
# Creates PR to main (not to release)
# Skips tag (tag happens after E2E on release)
```

**Verdict: WORKS, BUT CONFUSING FLOW** — This workflow bumps the version and creates a PR **back to main** (not to release). The flow is:

1. Run `prepare-release` on main → creates `chore/release-vX.Y.Z` branch with version bump
2. Merge that PR to main (version bump lands on main)
3. Then manually create a PR from main → release
4. E2E tests run on the release PR
5. Merge to release → tag-release auto-tags

This is functional but the extra step (PR to main for version bump, then separate PR to release) adds a manual step. In TestWorkflows, we used `TriPSs/conventional-changelog-action` directly on the release branch which is simpler.

**Assessment:** Works correctly. Added in PR #2127. Not broken, not unnecessarily complex — just a different (more cautious) approach to versioning.

---

### 7. `tag-release.yaml` — Auto-Tag on Release

**MyQQ:**
```yaml
on:
  push:
    branches: [release]
# Compares package.json version to latest git tag
# Creates tag if version is new
# Runs on: self-hosted
```

**Verdict: WORKS** — Simple and correct. When code lands on release with a new version in `package.json`, it creates the tag. Refactored in PR #2127 to be cleaner. No issues.

---

### 8. `backport-hotfix.yml` — Auto-Backport Hotfixes to Main

**MyQQ (current):**
```yaml
git cherry-pick ${{ github.event.pull_request.merge_commit_sha }}
# Then creates PR with enablePullRequestAutoMerge GraphQL mutation
```

**TestWorkflows (proven):**
```yaml
git cherry-pick -m 1 ${{ github.event.pull_request.merge_commit_sha }}
# Then creates PR and merges immediately via pulls.merge() API
```

**Verdict: BROKEN — Two known bugs:**

**Bug 1: Missing `-m 1` flag on cherry-pick**
The myqq workflow uses `git cherry-pick ${{ github.event.pull_request.merge_commit_sha }}` without the `-m 1` flag. We proved in TestWorkflows that this produces an **empty cherry-pick** because merge commits have two parents and git doesn't know which to diff against. This was the exact failure in our hotfix test #1 (PR #11).

**Bug 2: `enablePullRequestAutoMerge` requires branch protection**
The myqq workflow uses the GraphQL `enablePullRequestAutoMerge` mutation. We proved in TestWorkflows (hotfix test #3, PR #17) that this **fails with "Protected branch rules not configured"** if the target branch doesn't have branch protection set up. Even if branch protection exists, it requires the "Allow auto-merge" repo setting to be enabled.

**Fix required:**
```yaml
# Replace:
git cherry-pick ${{ github.event.pull_request.merge_commit_sha }}
# With:
git cherry-pick -m 1 ${{ github.event.pull_request.merge_commit_sha }}

# Replace enablePullRequestAutoMerge with:
await github.rest.pulls.merge({
  owner: context.repo.owner,
  repo: context.repo.repo,
  pull_number: pr.number,
  merge_method: 'squash'
});
```

**History:** The backport workflow was originally in the branching strategy document (PR #2122 era). PR #2200 ("GitHub actions for hotfix not working") tried to fix it, and PR #2210 ("Fixed backport hotfix") refined it further — but neither addressed the `-m 1` flag issue, which is the root cause.

---

### 9. `prepare-hotfix.yaml` — Hotfix Preparation

**MyQQ:**
```yaml
on:
  workflow_dispatch:
# Only runs on hotfix/* branches
# Commits any staged changes with fix(hotfix): prefix
# Bumps patch version with standard-version
# Creates PR to release branch
# Uses secrets.PR_OPENER token
```

**Verdict: WORKS** — Added in PR #2200. This is a convenience workflow that automates the tedious parts of hotfix preparation (version bump, changelog, PR creation). It's not strictly necessary — you could do these steps manually — but it reduces human error.

**Assessment:** This is **not added complexity**. It's automation of a manual process. Without it, developers would need to:
1. Manually run `npx standard-version --release-as patch`
2. Manually commit the version bump
3. Manually create the PR with the correct `fix(hotfix):` title

The workflow does all of this in one click. Added in PR #2200, refined in PR #2219.

---

### 10. `dependabot-lockfile.yml` — Dependabot Lock File Updates

**MyQQ:**
```yaml
on:
  pull_request:
    branches: [main]
    types: [opened, synchronize]
# Only runs for dependabot[bot]
# Updates package-lock.json and commits
```

**Verdict: WORKS** — Added in PR #2122. Standard pattern for keeping lockfiles in sync with Dependabot PRs. No issues.

---

### 11–14. Build Workflows (`build-dev.yaml`, `build-qa.yaml`, `build-preprod.yaml`, `build-prod.yaml`)

**Verdict: WORKS** — These are environment-specific build and deploy workflows. They follow a consistent pattern:
1. Generate sitemap for the target environment
2. Build Docker image with environment-specific tag
3. Push to Docker Hub
4. Create PR in `qqcw/flux_apps` to update Kubernetes deployment YAML

These are not related to the branching strategy and were not modified in the PR #2122+ changes. No issues.

---

### 15. `test_with_cypress_recording.yml` — Legacy Cypress Test

**Verdict: LIKELY OBSOLETE** — This appears to be an older version of the E2E test workflow. `test-release.yaml` and `test-main.yaml` (added in PR #2125) supersede it. It's manually triggered only. Not harmful but could be removed for clarity.

---

### 16. `test-qsys.yaml` — Qsys QA Testing

**Verdict: WORKS** — Manual trigger, runs Cypress against Qsys QA. Unrelated to the branching strategy. No issues.

---

## Summary: What Works, What's Broken

### Working Correctly (12 workflows)

| Workflow | Status | Notes |
|---|---|---|
| `validate-pr-title.yaml` | **WORKS** | Identical to proven pattern |
| `validate-hotfix-naming.yml` | **WORKS** | Identical to proven pattern |
| `test-pr.yaml` | **WORKS** | PR #2122 concurrency is a good addition |
| `test-release.yaml` | **WORKS** | PR #2125 correctly binds E2E to release |
| `test-main.yaml` | **WORKS** | Manual E2E escape hatch, could be removed |
| `prepare-release.yaml` | **WORKS** | PR #2127, different but valid versioning approach |
| `tag-release.yaml` | **WORKS** | PR #2127, clean auto-tagging |
| `prepare-hotfix.yaml` | **WORKS** | PR #2200, useful automation (not complexity) |
| `dependabot-lockfile.yml` | **WORKS** | PR #2122, standard pattern |
| `build-dev.yaml` | **WORKS** | Unchanged, not related to branching |
| `build-qa.yaml` | **WORKS** | Unchanged |
| `build-preprod.yaml` | **WORKS** | Unchanged |
| `build-prod.yaml` | **WORKS** | Unchanged |

### Broken (1 workflow, 2 bugs)

| Workflow | Bug | Impact | Fix |
|---|---|---|---|
| `backport-hotfix.yml` | Missing `-m 1` on cherry-pick | Empty cherry-pick, backport fails silently | Add `-m 1` flag |
| `backport-hotfix.yml` | `enablePullRequestAutoMerge` requires branch protection | Auto-merge fails if branch protection not configured | Replace with `pulls.merge()` API call |

### Potentially Obsolete (2 workflows)

| Workflow | Notes |
|---|---|
| `test_with_cypress_recording.yml` | Superseded by `test-release.yaml` and `test-main.yaml` |
| `test-qsys.yaml` | Manual QA testing, unrelated to branching strategy |

---

## Was PR #2122 Helpful or Harmful?

**PR #2122 was entirely helpful.** It made 4 changes:

| Change | Assessment |
|---|---|
| Added `dependabot-lockfile.yml` | **Helpful** — Standard automation, reduces manual lockfile commits |
| Fixed `backport-hotfix.yml` (removed `lower()`) | **Helpful** — `lower()` is not a valid GitHub Actions expression function |
| Added concurrency to `test-pr.yaml` | **Helpful** — Prevents redundant CI runs, saves runner time |
| Renamed `lint-pr.yaml` → `validate-pr-title.yaml` | **Helpful** — Better naming, no functional change |

None of these changes added complexity. They were all improvements.

---

## Did Changes After PR #2122 Help or Hurt?

### PR #2125: "Bind Release/Hotfix Workflows to Release Branch"
**HELPED** — Added `test-release.yaml` and `test-main.yaml`, properly binding E2E tests to the release branch. This is exactly what the branching strategy calls for.

### PR #2127: "Re-enabled Cypress Cloud capture for all E2E, enabled Video for main"
**HELPED** — Added `prepare-release.yaml` and refactored `tag-release.yaml`. The release preparation workflow automates version bumping and changelog generation. The tag workflow was cleaned up.

### PR #2133: "Update tag release yaml name"
**NEUTRAL** — Renamed `tag_release.yaml` → `tag-release.yaml`. Housekeeping only.

### PR #2200: "GitHub actions for hotfix not working"
**PARTIALLY HELPED** — Created `prepare-hotfix.yaml` (useful automation) and refactored `backport-hotfix.yml`. The prepare-hotfix workflow is good. However, the backport refactoring **did not fix the core `-m 1` bug** — the cherry-pick still fails for merge commits.

### PR #2210: "Fixed backport hotfix"
**DID NOT FIX THE CORE ISSUE** — Further refinements to `backport-hotfix.yml` improved conflict handling and error messages, but the cherry-pick still uses `merge_commit_sha` without `-m 1`.

### PR #2219: "Update prepare-hotfix yaml to commit change first"
**HELPED** — Fixed the ordering in `prepare-hotfix.yaml` to commit changes before running standard-version. Minor but correct fix.

---

## What Would It Take to Make MyQQ Work Like TestWorkflows?

### Required Changes (must fix)

**1. Fix `backport-hotfix.yml` cherry-pick (2 lines changed):**
```yaml
# Line 45 — change:
git cherry-pick ${{ github.event.pull_request.merge_commit_sha }}
# To:
git cherry-pick -m 1 ${{ github.event.pull_request.merge_commit_sha }}
```

**2. Fix `backport-hotfix.yml` auto-merge (replace GraphQL with REST):**

Replace the "Enable auto-merge" step (lines 89-125) with:
```yaml
      - name: Merge backport PR
        if: steps.check_conflicts.outputs.conflicts == 'false'
        uses: actions/github-script@v7
        with:
          script: |
            const pr = ${{ steps.create_pr.outputs.result }};
            await github.rest.pulls.merge({
              owner: context.repo.owner,
              repo: context.repo.repo,
              pull_number: pr.number,
              merge_method: 'squash'
            });
            console.log(`✅ Merged backport PR #${pr.number}`);
```

Or, combine the create + merge into a single step (as TestWorkflows does).

**3. Ensure "Allow GitHub Actions to create and approve pull requests" is enabled** in repo settings (Settings → Actions → General).

### Optional Improvements (nice to have)

- Remove `test_with_cypress_recording.yml` if it's truly superseded
- Remove `test-main.yaml` if manual E2E on main is never used
- Consider combining the "Create backport PR" and "Merge backport PR" into a single step for clarity

### No Changes Needed

Everything else works correctly. The build workflows, test workflows, validation workflows, release preparation, tagging, and hotfix preparation are all functional and well-structured.

---

## Complexity Assessment

**Claim: "Too much added complexity"**

The myqq repository has 16 workflow files. Breaking them down by category:

| Category | Count | Workflows | Complex? |
|---|---|---|---|
| Build/deploy (per environment) | 4 | build-dev, build-qa, build-preprod, build-prod | No — standard, existed before #2122 |
| Testing | 4 | test-pr, test-release, test-main, test-qsys | No — test-release is essential |
| Validation | 2 | validate-pr-title, validate-hotfix-naming | No — simple checks |
| Release management | 2 | prepare-release, tag-release | No — automates manual steps |
| Hotfix management | 2 | prepare-hotfix, backport-hotfix | No — automates manual steps |
| Dependency management | 1 | dependabot-lockfile | No — standard pattern |
| Legacy | 1 | test_with_cypress_recording | Could remove |

**Of the 16 workflows:**
- **4 existed before the branching strategy** (build-*)
- **4 are essential to the branching strategy** (validate-hotfix-naming, test-release, tag-release, backport-hotfix)
- **3 are useful automation** (prepare-release, prepare-hotfix, dependabot-lockfile)
- **3 are standard CI** (validate-pr-title, test-pr, test-qsys)
- **2 are potentially removable** (test-main, test_with_cypress_recording)

The workflows added after PR #2122 are: `test-release.yaml`, `test-main.yaml`, `prepare-release.yaml`, `tag-release.yaml` (refactored), `prepare-hotfix.yaml`. Of these, only `test-main.yaml` is arguably unnecessary. The rest are either essential to the branching strategy or helpful automation.

**Conclusion:** The "added complexity" claim is not supported by the evidence. The workflows are well-organized, each serves a clear purpose, and the only actual problem is a 2-line bug in the backport workflow's cherry-pick command.
