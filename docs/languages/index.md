---
title: Language support
type: reference
description: The languages Gavel analyzes, the tools per language, how lint aspects flow to SARIF, and how coverage is collected.
---

# Language Support

Gavel analyzes code quality through Bazel aspects — build-system-level
plugins that run static analyzers alongside compilation and produce
[SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html)
output. Each supported language has lint aspects, an architecture
validation aspect, and uses Bazel's native coverage infrastructure.

## Supported languages

| Language | Lint tools | Versions | Archtest | Coverage | Gavel-exclusive |
|----------|-----------|----------|----------|----------|-----------------|
| [Go](go.md) | golangci-lint | 2.11.4 | Yes | LCOV | — |
| [Java](java.md) | PMD, CPD, SpotBugs, Error Prone | PMD 7.24.0, SpotBugs 4.9.8, Error Prone 2.49.0 | Yes | LCOV | Error Prone, CPD |
| [Kotlin](java.md) | (Java tools) | (Java versions) | Yes | LCOV | Error Prone, CPD |
| [Python](python.md) | Ruff, Bandit, pycompile | Ruff 0.15.12, Bandit 1.9.4 | Yes | LCOV | pycompile |
| [TypeScript](typescript.md) | ESLint | — | Yes | Vitest JSON + LCOV fallback | — |
| [Rust](rust.md) | Clippy | (toolchain) | Yes | LCOV | — |

Source of truth: `core/infrastructure/platform/bazel/catalog/` — `aspect.go`,
`language_rules.go`, `tool_repo.go`.

## How aspects work

Gavel registers aspects via `.gavel/gavel.bazelrc`, which is included in
the workspace's `.bazelrc`. When `gavel judge` runs, it executes:

```
bazel build \
  --aspects=@gavel_tools//lint/aspects:defs.bzl%<lang>_<tool>_submission_aspect \
  --output_groups=gavel_submissions \
  --keep_going \
  -- <targets>
```

(The `--` marker precedes the targets so that excluded packages can be passed
as Bazel negative patterns; see [configuration](../configuration.md) `exclude`.)

> The lint aspects, per-tool wrappers, and the archtest library live in the
> external **`gavel_tools`** Bazel module, not in this repo. Paths below
> prefixed `lint/` are relative to that module — see its own docs for the
> authoritative layout. Gavel's interface to it is SARIF files on disk.

Each aspect:

1. Checks if the target provides the expected provider (e.g., `GoLibrary`
   for Go, `JavaInfo` for Java, `PyInfo` for Python)
2. Collects source files from the target's `srcs` attribute
3. Runs a wrapper binary that invokes the linter and produces a SARIF file
4. Returns the SARIF via the `gavel_submissions` output group

After Bazel completes, gavel walks `bazel-bin/` collecting all `*.sarif`
files matching the aspect's suffix (e.g., `.golangci.sarif`, `.pmd.sarif`).

### Wrapper pattern

Every lint tool is invoked through a Go wrapper binary at
`lint/lang/<language>/<tool>/` in the `gavel_tools` module. The wrapper:

- Resolves the tool binary (pinned version from Bazel or PATH fallback)
- Sets up the environment (GOPATH, GOMODCACHE, HOME, etc.)
- Runs the tool with SARIF output
- Handles tool-specific quirks (config file detection, classpath, etc.)

Wrappers run with `execution_requirements = {"no-sandbox": "1"}` because
most linters need filesystem access beyond Bazel's sandbox (module caches,
toolchain binaries, config files).

### Caching

Bazel caches aspect outputs based on declared inputs. When source files
change, Bazel invalidates the action and re-runs the aspect. Each
language doc describes its caching behavior — especially edge cases where
the tool reads files not declared as inputs (see [Go](go.md) for the
most significant example).

## Architecture validation

Every language has an `archtest` aspect that validates import/dependency
rules defined in `.gavel/architecture.yml`. The archtest wrappers parse
source files and check that imports follow the declared layer rules (e.g.,
"domain imports nothing outside domain").

Architecture aspects use the same wrapper pattern as lint aspects. The
shared archtest library lives in `lint/archtest/` in the `gavel_tools` module.

## Coverage

All languages use Bazel's native `bazel coverage` with
`--combined_report=lcov`. Gavel parses the LCOV output to compute
per-language coverage percentages.

Exception: TypeScript/JavaScript uses a composite strategy — vitest
coverage JSON is preferred when available, with LCOV as fallback.
See [TypeScript](typescript.md) for details.

Coverage collection is configured via:
- `--instrumentation_filter=//...` (default: all targets)
- `--test_size_filters=small,medium` (default)

## Adding a new language

Checklist (step 1 is in this repo; steps 2–5 are in the `gavel_tools` module):

1. **Catalog** (gavel): add entries to `languageAspects`, `languageRuleKinds`,
   and `languageTools` in `core/infrastructure/platform/bazel/catalog/`
2. **Aspect** (gavel_tools): add the aspect implementation to `lint/aspects/defs.bzl`
3. **Wrapper** (gavel_tools): create `lint/lang/<language>/<tool>/` with the Go
   binary that invokes the tool and produces SARIF
4. **Tool repository** (gavel_tools): declare the tool binary repo in the
   module's `MODULE.bazel` (`http_archive` / `http_file`)
5. **Archtest wrapper** (gavel_tools): add `<language>` support to `lint/archtest/`
6. **Example repo**: add `examples/<language>-repo/` with a working Bazel
   workspace and `.gavel/gavel.yaml`
7. **Tests**: wrapper unit tests + integration test via `gavel judge` on
   the example repo
8. **Documentation**: add `docs/languages/<language>.md`
