---
title: Java
type: reference
description: Java analysis via PMD, CPD, SpotBugs and Error Prone, run hermetically as Bazel aspects.
resource: https://github.com/gavelcode/gavel-tools/tree/main/lint/lang/java
tags: [java, pmd, spotbugs, error-prone, analyzers]
---

# Java

Gavel analyzes Java with four static analyzers — the broadest coverage of any
supported language — run as hermetic Bazel aspects from gavel-tools.

| | |
|---|---|
| **Lint tools** | PMD, CPD, SpotBugs, Error Prone |
| **Aspects** | `java_pmd_submission_aspect`, `java_cpd_submission_aspect`, `java_spotbugs_submission_aspect`, `java_error_prone_submission_aspect` |
| **Archtest** | `java_archtest_submission_aspect` |
| **Bazel targets** | `java_library`, `java_binary`, `java_test` |
| **SARIF suffixes** | `.pmd.sarif`, `.cpd.sarif`, `.spotbugs.sarif`, `.errorprone.sarif` |
| **Coverage** | LCOV via `bazel coverage` |

## What each tool catches

- **PMD** — source-level flaws: unused variables, empty catch blocks,
  unnecessary object creation. Configurable rulesets via XML.
- **CPD** (Copy-Paste Detector) — duplicated code across files, using PMD's
  engine.
- **SpotBugs** — bug patterns in **compiled bytecode** (null dereferences,
  resource leaks, concurrency issues); it analyzes the target's `.jar`, so its
  findings reflect the compiled output.
- **Error Prone** — Google's compile-time checker; AST-level mistakes
  (incompatible-type comparisons, missing `@Override`, …).

## How it runs

The aspects run each tool sandboxed, threading the Bazel JDK toolchain onto the
action rather than a host `java`. Source-based tools (PMD, CPD, Error Prone)
receive explicit file paths — so, unlike Go, they have no test-file cache
problem; SpotBugs consumes the target's jar. The per-tool mechanics live in
gavel-tools ([the hermetic analyzer driver](https://github.com/gavelcode/gavel-tools/blob/main/docs/tier-model.md)).

## Configuration

PMD reads a ruleset (its default, or a custom XML); SpotBugs and Error Prone use
their default rule sets. Which tools run for a project is its `gavel.yaml`
`tooling` selection (see [configuration](../configuration.md)).
