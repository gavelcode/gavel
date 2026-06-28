---
title: TypeScript
type: reference
description: TypeScript and JavaScript analysis via ESLint with a composite vitest and LCOV coverage strategy.
---

# TypeScript

Gavel analyzes TypeScript and JavaScript code with ESLint and provides
a composite coverage strategy that handles vitest-based projects.

| | |
|---|---|
| **Lint tool** | ESLint |
| **Aspect** | `typescript_eslint_submission_aspect` |
| **Archtest** | `typescript_archtest_submission_aspect` |
| **Bazel rule kinds** | `ts_project`, `ts_library`, `js_library`, `js_binary`, `js_test` |
| **SARIF suffix** | `.eslint.sarif` |
| **Wrapper** | `lint/lang/typescript/eslint/wrapper/main.go` |
| **Coverage** | Vitest JSON + LCOV fallback (composite) |

## Aspect mechanics

The ESLint aspect:

1. Collects `.ts` and `.tsx` source files from `srcs`
2. Sets the `BAZEL_BINDIR` environment variable to `ctx.bin_dir.path` —
   ESLint needs this to resolve generated files and Bazel output paths
3. Passes the ESLint binary as both a tool and an executable (unlike
   other languages where the tool is a standalone binary)
4. Runs the wrapper with file paths and produces SARIF

```starlark
env = {"BAZEL_BINDIR": ctx.bin_dir.path},
```

This `BAZEL_BINDIR` injection is unique to the TypeScript aspect. It
allows ESLint to find compiled outputs and type declarations generated
by `ts_project` rules.

## Coverage: composite strategy

TypeScript coverage uses a two-tier approach implemented in
`core/infrastructure/platform/bazel/collector/composite/coverage.go`:

1. **Primary**: look for vitest coverage JSON output. If found, convert
   it to LCOV format using `core/infrastructure/platform/bazel/runner/jscoverage.go`
2. **Fallback**: use Bazel's standard `--combined_report=lcov` output

This composite approach exists because:
- Many TypeScript projects use vitest as their test runner
- vitest produces its own coverage format (JSON) that is more accurate
  than Bazel's LCOV instrumentation for JavaScript
- `bazel coverage` with `rules_js` doesn't always produce reliable LCOV

The composite collector detects vitest output automatically. No user
configuration needed.

## Caching behavior

| Change | Cache invalidation |
|--------|--------------------|
| Edit a `.ts` or `.tsx` file | ESLint SARIF re-generated |
| Edit ESLint config | All SARIFs re-generated (config in `gavel_lint_config` filegroup) |

TypeScript aspects do not have the Go test-file problem. ESLint
receives explicit file paths as arguments.

## Configuration

ESLint reads its configuration from the standard locations
(`eslint.config.js`, `.eslintrc.*`, etc.) in the workspace root.
Config files are tracked via the `gavel_lint_config` filegroup.

## Known limitations

1. **Config file tracking requires `gavel_lint_config` filegroup.**
   `gavel init` generates the filegroup with globs for `eslint.config.*`
   and `.eslintrc.*`.
2. **`BAZEL_BINDIR` dependency.** The aspect injects this env var,
   but ESLint plugins that depend on other Bazel outputs may fail
   if those outputs are not in the declared inputs.
3. **Vitest coverage detection.** If vitest is not configured to
   output coverage JSON, the fallback LCOV may have lower accuracy
   for JavaScript/TypeScript.

## Wrapper reference

```
lint/lang/typescript/eslint/wrapper/main.go
```

| Flag | Default | Description |
|------|---------|-------------|
| `--eslint` | (required) | Path to the ESLint binary |
| `--out` | (required) | SARIF output file path |
| (positional) | | Source file paths to lint |
