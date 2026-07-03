---
title: Go
type: reference
description: Go analysis via golangci-lint, run hermetically as a Bazel aspect.
resource: https://github.com/gavelcode/gavel-tools/tree/main/lint/lang/go
tags: [go, golangci-lint, analyzers]
---

# Go

Gavel analyzes Go with [golangci-lint](https://golangci-lint.run/), the de-facto
meta-linter that aggregates 50+ analyzers, run as a **hermetic** Bazel aspect
from gavel-tools.

| | |
|---|---|
| **Lint tool** | golangci-lint |
| **Aspect** | `go_golangci_lint_submission_aspect` |
| **Archtest** | `go_archtest_submission_aspect` |
| **Bazel targets** | `go_library`, `go_binary`, `go_test` |
| **SARIF suffix** | `.golangci.sarif` |
| **Coverage** | LCOV via `bazel coverage` |

## How it runs

Running golangci-lint under Bazel is notoriously hard: it type-checks whole
packages (including `_test.go`) while Bazel works per-target, so a naive aspect
either reads undeclared files — breaking the cache — or misses test files.
gavel-tools solves this hermetically with a static package-graph driver, no
`no-sandbox` and no host `go`. The full story — why the ecosystem hasn't solved
it, how the driver works, and the contract for bumping `rules_go` /
golangci-lint — is in gavel-tools:

- [The hermetic analyzer driver](https://github.com/gavelcode/gavel-tools/blob/main/docs/tier-model.md)

## Configuration

golangci-lint reads `.golangci.yml` from the workspace root; any
[golangci-lint configuration](https://golangci-lint.run/usage/configuration/)
applies. gavel tracks the config file so edits invalidate the aspect cache.

The quality-gate rules that act on Go findings are the usual `findings`,
`coverage` and `architecture_violations` (see [configuration](../configuration.md)).
