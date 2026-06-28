---
title: Go
type: reference
description: Go analysis via golangci-lint — the dual-aspect strategy, caching edge cases, and coverage.
---

# Go

Gavel analyzes Go code with [golangci-lint](https://golangci-lint.run/), a
meta-linter that aggregates 50+ analyzers into a single tool. It is the
de facto standard for Go linting, used by Kubernetes, Prometheus, and
Terraform.

| | |
|---|---|
| **Lint tool** | golangci-lint 2.11.4 |
| **Aspect** | `go_golangci_lint_submission_aspect` |
| **Archtest** | `go_archtest_submission_aspect` |
| **Bazel rule kinds** | `go_library`, `go_binary`, `go_test` |
| **SARIF suffix** | `.golangci.sarif` |
| **Wrapper** | `lint/lang/go/golangci_lint/wrapper/main.go` |
| **Coverage** | LCOV via `bazel coverage` |

## State of the art: Go linting in Bazel

Integrating golangci-lint with Bazel is an unsolved problem in the
ecosystem. The fundamental tension:

- **golangci-lint** analyzes complete Go packages. Its type-checker needs
  all files in the package — including `_test.go` — to resolve types,
  interfaces, and cross-file references.
- **Bazel** works by targets, not by directories. A `go_library` target
  declares its source files explicitly. A `go_test` target in the same
  package is a separate target with its own `srcs`.
- **Bazel aspects** attached to a `go_library` only see the library's
  `srcs`. The test files are invisible to the aspect, even though
  golangci-lint reads them from the filesystem.

This means an aspect that runs golangci-lint on a `go_library` target
analyzes files (test files) that it does not declare as Bazel inputs.
When those files change, Bazel does not invalidate the action cache,
and stale SARIF output is served.

### What others tried

**1. Aspect Build `rules_lint`** — Had a `lint_golangci_aspect`. Hit the
same transitive-srcs problem. Declared it a "fatal bug" and removed
golangci-lint support entirely (Jan 2024). As of 2026, there is no
golangci-lint support in rules_lint.

- [Issue #129: golangci-lint doesn't pull in dependencies](https://github.com/aspect-build/rules_lint/issues/129)
- [PR #207: Remove golangci-lint (breaking change)](https://github.com/aspect-build/rules_lint/pull/207)

**2. `nogo` (rules_go built-in)** — Runs `go/analysis` framework
analyzers as part of the `GoCompilePkg` action, at compilation time.
Caching is correct because the compiler sees all package files. But
nogo only supports analyzers written against the `go/analysis` API — it
cannot run the full golangci-lint suite. Other limitations: no `--fix`
mode, lint failures block compilation, hard to parallelize.

- [nogo documentation](https://github.com/bazel-contrib/rules_go/blob/master/go/nogo.rst)

**3. `nogo-analyzer` / `nogo-analyzer-golangci-lint`** (sluongng) —
Proof-of-concept to port golangci-lint analyzers into the nogo framework.
Author explicitly warns: "should NOT be used except for research purposes."

- [nogo-analyzer](https://github.com/sluongng/nogo-analyzer)
- [nogo-analyzer-golangci-lint](https://github.com/sluongng/nogo-analyzer-golangci-lint)

**4. `bazel-nogo-lint`** (anuvu) — Another attempt to port golangci-lint
linters to rules_go. Incomplete, unmaintained since 2020.

- [bazel-nogo-lint](https://github.com/anuvu/bazel-nogo-lint)

**5. Native golangci-lint Bazel support** — Open issue since Oct 2020.
Two approaches proposed (single analysis tool, config generation), neither
implemented.

- [golangci-lint Issue #1473](https://github.com/golangci/golangci-lint/issues/1473)

**6. Comparing Bazel linting systems** (Aspect Build) — Evaluates nogo,
aspects, and other approaches. Concludes that aspects are the right model
but acknowledges the file-discovery problem: tools that "talk to git and
look around the whole workspace" conflict with Bazel's hermetic model.

- [Comparing Bazel linting systems](https://hackmd.io/@aspect/static-analysis)

### What all approaches have in common

They either give up on golangci-lint (nogo, rules_lint) or accept broken
caching (`no-sandbox` + undeclared inputs). No published solution keeps
golangci-lint, correct Bazel caching, AND test file coverage.

## Gavel's solution: dual-aspect strategy + sibling-source tracking

Gavel solves the cache invalidation problem with two mechanisms:

1. **Dual-aspect strategy** — generate two SARIFs per package (one from
   `go_library`, one from `go_test`) so each target's `srcs` trigger
   cache invalidation for their own file type.
2. **Sibling-source tracking** — for `go_test` targets, include the
   `.go` **source files** of the same-package library (reached through
   `deps`/`embed`) as lint action inputs, so editing a library file
   invalidates the test's lint cache.

| Target | `--skip-tests` | Inputs | Analyzes |
|--------|----------------|--------|----------|
| `go_library` | Yes | library `.go` files | non-test code only |
| `go_test` | No | test `.go` files + sibling library `.go` files + dep `.x` archives | entire package (including tests) |

### Sibling-source tracking

golangci-lint runs at package-directory level with `no-sandbox`, so it
reads **every** `.go` file in the package from the filesystem —
including the library files that belong to the sibling `go_library`
target, not the `go_test` target's own `srcs`. Bazel only invalidates
an action when a **declared** input changes, so those library files
must be declared as inputs of the `go_test` lint action.

The compiled `.x` archive is **not** sufficient for this: `.x` is Go
export data (the package's public interface), so body-local edits — a
renamed local variable, a changed function body, the implementation
details where most lint findings live — leave `.x` byte-identical. A
`go_test` action that tracks only `.x` therefore serves a **stale
SARIF** on implementation-only edits (a finding that no longer exists,
or a missing new finding) until the package's public API changes or
`bazel clean` runs.

To fix this, each Go target's aspect exposes its own `.go` source files
via the `GavelGoSrcInfo` provider. `_collect_sibling_srcs(ctx)` reads
that provider from the `go_test`'s `deps`/`embed`, keeps only deps in
the **same package** (`dep[GavelGoSrcInfo].package == ctx.label.package`),
and adds those source files to the lint action inputs. Editing any
library source file in the package now changes a declared input of the
test's lint action, invalidating its cache. The `.x` archives are still
tracked for cross-package interface changes.

### Cache invalidation by file type

When a **test file** changes, Bazel re-evaluates the `go_test` target
(its `srcs` changed), re-runs the lint aspect, and produces a fresh
SARIF. The `go_library` aspect stays cached because its inputs did not
change.

When a **library file** changes, the `go_library` aspect re-runs (its
`srcs` changed). The `go_test` aspect also re-runs because the sibling
library's source files — declared inputs via `GavelGoSrcInfo` — changed,
producing fresh SARIF that reflects the library edit (including
body-local, interface-stable edits that leave `.x` unchanged).

### Deduplication

Since both aspects analyze the non-test files, findings from library code
appear in both SARIFs with the same fingerprint. `ExtractFindings` in
`core/application/casefile/evidencedto/finding.go` deduplicates by
fingerprint before counting, so each finding is counted once.

### Trade-offs

- Non-test code is analyzed twice per package (once by each aspect).
  Because the `go_test` action declares the sibling library sources as
  inputs, a library edit re-runs both aspects — the cost of correct
  cache invalidation.
- Packages without `go_test` targets only get the `go_library` SARIF
  (with `--skip-tests`). Test files in those packages are not analyzed
  until a `go_test` target is added.

### Key files

| File | Role |
|------|------|
| `lint/aspects/defs.bzl` | Dual-aspect + `GavelGoSrcInfo` sibling-source tracking |
| `lint/lang/go/golangci_lint/wrapper/main.go` | `--skip-tests` flag → `--tests=false` to golangci-lint |
| `core/application/casefile/evidencedto/finding.go` | `ExtractFindings` dedup by fingerprint |

## Aspect mechanics

The aspect implementation in `defs.bzl`:

1. Checks `_is_go_lint_target(target, ctx)` — returns true for `GoLibrary`
   providers OR `go_test` rule kind
2. Collects `.go` source files from `ctx.rule.attr.srcs`
3. For `go_test` targets, collects same-package sibling library `.go`
   sources via `_collect_sibling_srcs(ctx)` (using the `GavelGoSrcInfo`
   provider) plus dep library `.x` archives via `_collect_dep_outputs(ctx)`
   for cache invalidation
4. Runs golangci-lint with inputs = `srcs` + sibling sources + dep
   outputs + `go.mod` + `go.sum` + lint config
5. For `go_library` targets, passes `--skip-tests` to the wrapper
6. Runs the wrapper with `no-sandbox` (see below)
7. Returns the SARIF via `gavel_submissions` output group, plus
   `GavelGoSrcInfo` carrying the target's own sources for sibling tracking

### Why `no-sandbox`

golangci-lint shells out to the Go compiler, which needs:
- `GOPATH` and `GOMODCACHE` for module resolution
- `GOROOT` for the Go standard library
- `HOME` for various Go toolchain caches

Bazel's sandbox restricts filesystem access to declared inputs. Running
golangci-lint in a sandbox causes module resolution failures. The
`no-sandbox` execution requirement allows the wrapper to set up the
full Go environment.

## Caching behavior

| Change | `go_library` aspect | `go_test` aspect | Mechanism |
|--------|--------------------|--------------------|-----------|
| Edit a `.go` library file | re-generated | re-generated | library srcs changed + dep `.x` changed |
| Edit a `_test.go` file | cached | re-generated | test srcs changed |
| Edit `go.mod` or `go.sum` | re-generated | re-generated | declared input changed |
| Edit `.golangci.yml` | re-generated | re-generated | `gavel_lint_config` filegroup changed |
| No changes | cached | cached | all inputs unchanged |

### GOCACHE and GOLANGCI_LINT_CACHE

The wrapper manages two caches:

- `GOCACHE` → `/tmp/gavel-go-build-cache` — persistent across runs,
  speeds up Go compilation within golangci-lint
- `GOLANGCI_LINT_CACHE` → fresh tmpdir per run — prevents stale linter
  state from affecting results

## Configuration

golangci-lint reads `.golangci.yml` from the workspace root. The wrapper
auto-detects this file at line 67 of `main.go`. Any golangci-lint
configuration documented at [golangci-lint.run](https://golangci-lint.run/usage/configuration/)
is supported.

Gavel quality gate rules that apply to Go findings:

- `code_quality.max_error` — maximum error-severity findings allowed
- `coverage.min` — minimum coverage percentage
- `architecture.max` — maximum architecture violations

## Known limitations

1. **Config file changes require `gavel_lint_config` filegroup.** The
   `gavel init` command generates a `gavel_lint_config` filegroup in the
   root BUILD.bazel that captures linter config files via glob. If this
   filegroup is missing, config changes are not detected.
2. **Packages without `go_test` targets** do not get test file analysis.
   golangci-lint runs with `--tests=false` on the `go_library` aspect.
3. **`no-sandbox` breaks hermeticity.** The linter sees the full host
   filesystem, which means results may vary across machines with different
   Go toolchain installations.
4. **golangci-lint v2 migration.** The wrapper targets golangci-lint v2
   output flags (`--output.sarif.path`). Projects using v1 config format
   may need to migrate.

## Wrapper reference

```
lint/lang/go/golangci_lint/wrapper/main.go
```

| Flag | Default | Description |
|------|---------|-------------|
| `--golangci-lint` | (PATH lookup) | Path to the golangci-lint binary |
| `--go` | (PATH lookup) | Path to the Go binary |
| `--package` | (required) | Go package directory to lint |
| `--out` | (required) | SARIF output file path |
| `--skip-tests` | `false` | Exclude test files from analysis |
