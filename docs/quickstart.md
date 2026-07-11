---
title: Quickstart
type: tutorial
description: From zero to a first quality report in five minutes — install, init, judge.
tags: [quickstart, getting-started]
---

# Quickstart

From zero to first quality report in 5 minutes.

## Requirements

| Requirement | Supported |
|-------------|-----------|
| Bazel       | **8.0 or newer**, bzlmod (`MODULE.bazel`) |
| OS          | Linux, macOS (amd64 / arm64) — Windows is not supported |

Bazel **8.0** is the floor: gavel's Go analyzer references `rules_go` by the
canonical repository name Bazel 8 mints (`rules_go+`), and Bazel 7 uses a
different scheme (`rules_go~`) that cannot resolve it, so `gavel judge` fails at
repository resolution before any build runs. Non-Go repos may still work on
Bazel 7, but 7.x is unsupported. The flag itself lives in
[`catalog.yaml`](../core/infrastructure/platform/bazel/catalog/catalog.yaml).

**Restricted / hermetic networks.** `gavel init` adds the gavel module registry
(`gavelcode.github.io`) to your `.bazelrc`, and `gavel_tools` fetches its
analyzer binaries from `github.com` and `repo1.maven.org`. A repo with a
locked-down `--downloader_config` (an allowlist plus `block *` mirror) must
allowlist those hosts, or vendor `gavel_tools` into its own mirror — gavel
cannot bypass a hermetic download policy, by design. The authoritative host
list is `gavel_tools`' `lint/lang/*/*/repositories.bzl`.

## 1. Install

```sh
# Homebrew — macOS & Linux
brew install gavelcode/tap/gavel

# or install script — Linux, CI, Docker
curl -fsSL https://raw.githubusercontent.com/gavelcode/gavel/main/install.sh | sh
```

Or download a binary from the [releases page](https://github.com/gavelcode/gavel/releases). Confirm with `gavel --version`.

Building from source instead: `bazel build //apps/cli/cmd/gavel` (binary at `bazel-bin/apps/cli/cmd/gavel/gavel_/gavel`).

## 2. Initialize

Interactive mode — Gavel asks for project name, Bazel pattern, and languages:

```bash
gavel init
```

From an existing config:

```bash
gavel init --from .gavel/gavel.yaml
```

This generates:
- `.gavel/gavel.yaml` — project definitions and quality gate rules
- `.gavel/architecture.yml` — DDD layer rules (auto-generated per language)
- `.gavel/gavel.bazelrc` — Bazel aspect registrations
- `.gavel/gavel.MODULE.bazel` — tool dependency declarations

It also appends `try-import` and `include` lines to your `.bazelrc` and
`MODULE.bazel`, and adds `common --registry=https://gavelcode.github.io/registry`
to your `.bazelrc` so Bazel can resolve `gavel_tools`. To move that pin later,
see [Upgrading gavel_tools](upgrading.md).

## 3. Verify

```bash
gavel validate
```

Checks that Bazel integration is healthy. Fix any issues before proceeding.

## 4. First judge run

```bash
gavel judge
```

This runs the full pipeline:
1. Lint aspects execute via Bazel (golangci-lint, PMD, ESLint, etc.)
2. Coverage is collected via `bazel coverage`
3. Architecture constraints are checked via archtest aspects
4. Quality gate is evaluated, verdict is rendered

**First run behavior:**
- The verdict may be FAIL if your code does not meet all quality gate
  thresholds — this is expected for existing codebases.
- A **baseline** is created regardless of the verdict. It is saved in
  `.gavel/baseline/<project>/` and contains fingerprints of all current
  findings, architecture violation IDs, and coverage percentage.
- From the second run onward, the quality gate evaluates only **new**
  findings and violations compared to this baseline.

## 5. Interpret the output

```
  payments
    src/api/handler.go
        42  error    null check missing              PMD:NullCheck  NEW
        18  warning  exposed field                   SpotBugs:EI_EXPOSE

  42 findings (↓5 since last run)
  coverage: 73.5% (↑5.0%)
  3 new · 8 fixed · 31 existing
```

- **NEW** findings are introduced since the baseline — these are what
  the quality gate evaluates.
- **fixed** findings were in the baseline but no longer appear.
- **existing** findings are in the baseline and still present — they do
  not affect the quality gate verdict.

## 6. Share the baseline

```bash
git add .gavel/baseline/
git commit -m "chore: add gavel baseline"
```

Committing the baseline to git ensures the entire team shares the same
reference point. CI and other developers will compare against this
baseline.

## 7. Iterate

Fix some findings, then run `gavel judge` again. The delta shows your
progress:

```
  3 new · 8 fixed · 31 existing
```

Each passing run updates the baseline with the current state. Each
failing run ratchets the baseline — resolved findings are removed but
new findings are not added, so they remain visible until fixed.

## 8. Configure the quality gate

Edit `.gavel/gavel.yaml` to adjust thresholds for your project:

```yaml
quality_gate:
  findings:
    max_error: 0           # zero new errors allowed
  coverage:
    min: 60                # minimum 60% coverage
    min_delta: 0.0         # coverage must not decrease
  architecture_violations:
    max: 0                 # zero new violations allowed
```

All finding and violation thresholds evaluate against **new** items only
(compared to the baseline). Coverage thresholds are absolute. Use
`min_delta` to require coverage improvement and `min_resolved` to require
a minimum number of findings resolved per run.

See [Configuration Reference](configuration.md) for the full schema.

## Useful flags

| Flag | Purpose |
|------|---------|
| `--project <name>` | Analyze a single project |
| `--json` | Structured output for CI pipelines |
| `--absolute` | Evaluate all findings, ignore baseline (release gates) |
| `--quick` | Skip coverage and architecture checks |
| `--output-sarif report.sarif` | Export SARIF for IDE/GitHub integration |
