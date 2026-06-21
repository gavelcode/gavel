---
title: CI integration
type: how-to
description: Running gavel judge in CI pipelines using the single static binary.
---

# CI Integration

Gavel is designed to run in CI pipelines. The CLI is a single static Go
binary with no external dependencies beyond Bazel.

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
