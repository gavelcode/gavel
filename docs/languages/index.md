---
title: Language support
type: reference
description: The languages Gavel analyzes and how their analyzers, run hermetically by gavel-tools, flow to SARIF and coverage.
resource: https://github.com/gavelcode/gavel-tools/blob/main/lint/catalog.yaml
tags: [languages, analyzers, sarif]
---

# Language support

Gavel analyzes code quality by running each language's static analyzers as
**hermetic Bazel aspects** that emit [SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html).
The aspects, wrappers and pinned tool versions all live in the external
[gavel-tools](https://github.com/gavelcode/gavel-tools) module — its
[catalog](https://github.com/gavelcode/gavel-tools/blob/main/lint/catalog.yaml)
is the source of truth for which tools exist per language. Gavel's interface to
all of it is **SARIF files on disk**; it never imports gavel-tools' Go.

## Supported languages

| Language | Lint tools | Architecture | Coverage |
|----------|-----------|--------------|----------|
| [Go](go.md) | golangci-lint | archtest | LCOV |
| [Java](java.md) | PMD, CPD, SpotBugs, Error Prone | archtest | LCOV |
| [Python](python.md) | Ruff, Bandit, pycompile | archtest | LCOV |
| [TypeScript](typescript.md) | ESLint | archtest | vitest JSON + LCOV fallback |
| [Rust](rust.md) | Clippy | archtest | LCOV |

A project selects which of these it runs per language in its `gavel.yaml`
`tooling` map (see [configuration](../configuration.md)); the pinned tool
versions live in the gavel-tools catalog.

## How analysis runs

`gavel judge` builds gavel-tools' submission aspects over the project's targets
and collects the `*.sarif` files they emit under `bazel-bin/`:

```
bazel build \
  --aspects=@gavel_tools//lint/aspects:defs.bzl%<lang>_<tool>_submission_aspect \
  --output_groups=gavel_submissions \
  --keep_going -- <targets>
```

Every analyzer runs **sandboxed and hermetic** — no host toolchain, every input
declared. How each is made hermetic (the golangci package-graph driver, the
ESLint pnpm-store repair) and the maintenance contract for keeping it working
live in gavel-tools:
[The hermetic analyzer driver](https://github.com/gavelcode/gavel-tools/blob/main/docs/tier-model.md).

## Architecture validation

Every language also has an `archtest` aspect that checks the import/dependency
rules in `.gavel/architecture.yml` (see [configuration](../configuration.md)).
The rules are language-agnostic; each language's archtest wrapper parses its
imports.

## Coverage

Coverage is Bazel's native `bazel coverage --combined_report=lcov`, parsed into
per-project percentages — except TypeScript, which prefers vitest's own coverage
(see [TypeScript](typescript.md)).

## Adding a language

A new language is almost entirely gavel-tools work — add its catalog entry,
aspect, wrapper and tool repo there (gavel-tools'
[CONTRIBUTING](https://github.com/gavelcode/gavel-tools/blob/main/CONTRIBUTING.md)
walks through it). Gavel consumes it automatically once published: a project
just lists the new tools in its `tooling` map.
