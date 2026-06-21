---
title: Java and Kotlin
type: reference
description: Java and Kotlin analysis via PMD, CPD, SpotBugs, and Error Prone.
---

# Java

Gavel analyzes Java (and Kotlin) code with four static analysis tools —
the broadest coverage of any supported language. Two are standard
open-source tools (PMD, SpotBugs), and two are Gavel-exclusive (CPD,
Error Prone).

| | |
|---|---|
| **Lint tools** | PMD 7.24.0, CPD (via PMD), SpotBugs 4.9.8, Error Prone 2.49.0 |
| **Aspects** | `java_pmd_submission_aspect`, `java_cpd_submission_aspect`, `java_spotbugs_submission_aspect`, `java_error_prone_submission_aspect` |
| **Archtest** | `java_archtest_submission_aspect` |
| **Bazel rule kinds** | `java_library`, `java_binary`, `java_test` |
| **SARIF suffixes** | `.pmd.sarif`, `.cpd.sarif`, `.spotbugs.sarif`, `.errorprone.sarif` |
| **Wrappers** | `tools/java/{pmd,cpd,spotbugs,error_prone}/wrapper/main.go` |
| **Coverage** | LCOV via `bazel coverage` |

## Tool overview

**PMD** — Source code analyzer that finds common programming flaws
(unused variables, empty catch blocks, unnecessary object creation).
Configurable rulesets via XML.

**CPD** (Copy-Paste Detector) — Finds duplicated code blocks across
files. Uses PMD's engine internally. **Gavel-exclusive**: not available
in `rules_lint` or standard Bazel setups.

**SpotBugs** — Bytecode analyzer that finds bugs in compiled Java
(null pointer dereferences, resource leaks, concurrency issues).
Operates on `.jar` files, not source.

**Error Prone** — Compile-time bug checker from Google. Catches common
mistakes at the AST level (comparison of incompatible types, missing
`@Override`, etc.). **Gavel-exclusive**.

## Aspect mechanics

### PMD and CPD

Both use the same underlying PMD binary. The aspect:

1. Checks for `JavaInfo` provider
2. Collects `.java` source files from `srcs`
3. Passes individual file paths to the wrapper
4. Wrapper runs PMD/CPD and produces SARIF

Inputs are source files only — no classpath needed.

### SpotBugs

SpotBugs analyzes **compiled bytecode**, not source:

1. Checks for `JavaInfo` provider
2. Extracts `runtime_output_jars` from `JavaInfo` (`.jar` files)
3. Passes jars to the wrapper
4. Wrapper runs SpotBugs on the jars and produces SARIF

This means SpotBugs findings reflect the compiled output, not just
source. Changes to dependencies that affect compilation can produce
new SpotBugs findings.

### Error Prone

Error Prone needs both source files and the compile-time classpath:

1. Checks for `JavaInfo` provider
2. Collects `.java` source files
3. Extracts `transitive_compile_time_jars` from `JavaInfo`
4. Passes both to the wrapper with `--classpath`
5. Also requires `error_prone.jar` and `dataflow_errorprone.jar` as tools

## Kotlin support

Kotlin projects using `kt_jvm_library`, `kt_jvm_binary`, or `kt_jvm_test`
trigger the same Java aspects. This works because Kotlin/JVM targets
expose `JavaInfo` and compile to `.class` files / `.jar` archives.

The Kotlin catalog entry in `languageAspects` maps to the same aspect
names as Java.

## Caching behavior

| Change | Cache invalidation |
|--------|--------------------|
| Edit a `.java` source file | PMD, CPD, Error Prone SARIFs re-generated |
| Edit a `.java` source file | SpotBugs SARIF re-generated (jar changes) |
| Add/remove a dependency | SpotBugs and Error Prone SARIFs re-generated (classpath changes) |

Java aspects do not have the Go test-file problem because PMD, CPD,
and Error Prone receive explicit file paths as arguments (not a package
directory). They only analyze the files passed to them, and those files
match the declared inputs.

SpotBugs operates on jars, which are Bazel outputs — correctly tracked.

## Configuration

PMD reads rulesets from its default configuration or a custom XML file.
SpotBugs and Error Prone use their default rule sets. Configuration
files are not currently declared as Bazel inputs (same limitation as
Go's `.golangci.yml`).

## Gavel-exclusive tools

**Error Prone** and **CPD** are only available through Gavel. Standard
Bazel setups and `rules_lint` do not provide aspects for these tools.
They are marked in `gavelExclusiveAspects` in `catalog/aspect.go`.

## Known limitations

1. **No custom PMD ruleset path in aspect.** The wrapper uses PMD's
   default rulesets. Custom rulesets require wrapper modification.
2. **SpotBugs requires compilation.** If the Java target fails to
   compile, SpotBugs cannot produce findings.
3. **Config file tracking requires `gavel_lint_config` filegroup.** Same
   mechanism as Go — `gavel init` generates the filegroup.

## Wrapper reference

| Wrapper | Flags |
|---------|-------|
| `tools/java/pmd/wrapper/main.go` | `--pmd <binary> --out <sarif> <files...>` |
| `tools/java/cpd/wrapper/main.go` | `--pmd <binary> --out <sarif> <files...>` |
| `tools/java/spotbugs/wrapper/main.go` | `--spotbugs <binary> --out <sarif> <jars...>` |
| `tools/java/error_prone/wrapper/main.go` | `--error-prone-jar <jar> --dataflow-jar <jar> --out <sarif> [--classpath <cp>] <files...>` |
