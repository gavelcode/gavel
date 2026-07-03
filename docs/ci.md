---
title: CI integration
type: how-to
description: Running gavel judge in CI pipelines using the single static binary.
tags: [ci, automation]
---

# CI Integration

Gavel is designed to run in CI pipelines. The CLI is a single static Go
binary with no external dependencies beyond Bazel.

## Installing in CI

Most CI runners don't ship `gavel`. Add an install step before running it:

```yaml
- name: Install Gavel
  run: curl -fsSL https://raw.githubusercontent.com/gavelcode/gavel/main/install.sh | sh
```

On macOS runners, `brew install gavelcode/tap/gavel` also works.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | All projects passed the quality gate |
| 1 | One or more projects failed the quality gate |

Use the exit code directly in CI to gate merges.

## Basic usage

```bash
gavel judge --json
```

`--json` produces structured output suitable for parsing. Without it,
Gavel renders a human-readable terminal report with colors and grouping.

## Baseline strategies

### Local baseline (no server)

Commit `.gavel/baseline/` to git. CI checks out the repo, runs
`gavel judge`, and the baseline is already available from the checkout.

```yaml
# GitHub Actions example
- name: Run Gavel
  run: gavel judge --json
```

Limitation: CI does not commit back, so the baseline only updates when a
developer pushes from a local run. This works for small teams where
developers run `gavel judge` locally before pushing.

### Server baseline (centralized)

Use `--server` and `--token` to fetch the baseline from a Gavel server
and submit results back. The server maintains the baseline centrally.

```yaml
- name: Run Gavel
  run: |
    gavel judge --json \
      --server ${{ secrets.GAVEL_SERVER_URL }} \
      --token ${{ secrets.GAVEL_TOKEN }}
```

This is the recommended approach for teams and CI, as the baseline is
always up to date regardless of local developer runs.

## PR gating

Pass PR metadata to associate the analysis with a pull request:

```bash
gavel judge --json \
  --pr-number $PR_NUMBER \
  --pr-branch $BRANCH \
  --pr-author $AUTHOR \
  --pr-title "$TITLE"
```

When using server mode, this creates or updates the pleading (PR record)
on the server, enabling the web dashboard to show PR-level quality data.

## GitHub PR checks

`gavel report` delivers a judged verdict to GitHub as a check run: a
merge-blocking pass/fail with new findings annotated inline on the
pull-request diff. It reads the verdict `gavel judge` caches under
`.gavel/results/` and never re-runs analysis, so it runs as a **separate
step after `judge`**. Its flags and defaults are defined in
[`clispec/v1/clispec.yaml`](../clispec/v1/clispec.yaml).

In GitHub Actions the built-in `GITHUB_TOKEN` already carries the
`checks:write` permission, so no GitHub App or extra secret is needed:

```yaml
jobs:
  gavel:
    runs-on: ubuntu-latest
    permissions:
      checks: write            # gavel report creates the check run
    steps:
      - uses: actions/checkout@v4
      - name: Install Gavel
        run: curl -fsSL https://raw.githubusercontent.com/gavelcode/gavel/main/install.sh | sh
      - name: Judge
        run: gavel judge
      - name: Report to the PR
        if: always()           # decorate the PR even when the gate fails
        run: gavel report
```

`gavel report` reads `GITHUB_TOKEN`, `GITHUB_REPOSITORY`, and `GITHUB_SHA`
from the Actions environment by default. `if: always()` runs it even when
`judge` failed the gate — which is exactly when you want the red check.
`judge` writes the verdict cache before it exits, so the failing verdict
is available to report. Report's own exit code reflects **delivery**, not
the verdict: the pass/fail rides in the check run's conclusion, so
`gavel judge` stays the gate.

Two caveats:

- **Fork pull requests.** GitHub grants the default `GITHUB_TOKEN` only
  read access on PRs from forks, so the check run cannot be created there.
  Same-repo branch PRs work without extra setup; forks need a GitHub App
  or `pull_request_target` (which carries its own security trade-offs).
- **Head commit.** On `pull_request` events `GITHUB_SHA` is the merge
  commit, not the PR head. Pass `--commit ${{ github.event.pull_request.head.sha }}`
  when you need annotations anchored to the head commit.

## Release gates

Use `--absolute` to evaluate against ALL findings, not just new ones.
This ignores the baseline entirely and is useful for release gates or
nightly audits:

```bash
gavel judge --json --absolute
```

## SARIF export

Export findings as SARIF 2.1.0 for integration with GitHub Code Scanning
or IDE tooling:

```bash
gavel judge --output-sarif report.sarif
```

Upload to GitHub Code Scanning:

```yaml
- name: Run Gavel
  run: gavel judge --output-sarif report.sarif
  continue-on-error: true

- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: report.sarif
```

## Single project

Scope the analysis to one project when the monorepo is large:

```bash
gavel judge --project payments --json
```

## Timeout

Set a maximum time for the entire judge run (default 30 minutes):

```bash
gavel judge --timeout 10m
```
