---
title: Python
type: reference
description: Python analysis via Ruff, Bandit, and pycompile.
---

# Python

Gavel analyzes Python code with three tools: Ruff (primary linter),
Bandit (security), and pycompile (syntax validation). pycompile is
Gavel-exclusive.

| | |
|---|---|
| **Lint tools** | Ruff 0.15.12, Bandit 1.9.4, pycompile |
| **Aspects** | `python_ruff_submission_aspect`, `python_bandit_submission_aspect`, `python_pycompile_submission_aspect` |
| **Archtest** | `python_archtest_submission_aspect` |
| **Bazel rule kinds** | `py_library`, `py_binary`, `py_test` |
| **SARIF suffixes** | `.ruff.sarif`, `.bandit.sarif`, `.pycompile.sarif` |
| **Wrappers** | `lint/lang/python/{ruff,bandit,pycompile}/wrapper/main.go` |
| **Coverage** | LCOV via `bazel coverage` |

## Tool overview

**Ruff** — Extremely fast Python linter written in Rust. Replaces
flake8, isort, pycodestyle, pyflakes, and dozens of other tools in a
single binary. Gavel's primary Python lint tool.

**Bandit** — Security-focused analyzer from PyCQA. Finds common
security issues: hardcoded passwords, use of `eval()`, insecure
hash algorithms, SQL injection patterns.

**pycompile** — Syntax validation using Python's built-in `py_compile`
module. Catches syntax errors that prevent import. **Gavel-exclusive**:
not available in `rules_lint` or standard Bazel setups.

## Aspect mechanics

### Ruff and pycompile

Both receive individual `.py` file paths:

1. Check for `PyInfo` provider
2. Collect `.py` source files from `srcs`
3. Pass file paths to the wrapper
4. Wrapper runs the tool and produces SARIF

Inputs are source files only. No Python environment or virtualenv needed.

### Bandit

Bandit requires access to its own `site-packages` for security rule
plugins:

1. Checks for `PyInfo` provider
2. Collects `.py` source files
3. Resolves `site-packages` from the `_bandit_packages` attribute
   (`@bandit//:site_packages`)
4. Passes `--site-packages <dir>` plus file paths to the wrapper

The site-packages files are declared as inputs, so changes to Bandit's
dependencies correctly invalidate the cache.

## Caching behavior

| Change | Cache invalidation |
|--------|--------------------|
| Edit a `.py` file | All three SARIFs re-generated |
| Update Bandit version | Bandit SARIF re-generated |
| Update Ruff version | Ruff SARIF re-generated |

Python aspects do not have the Go test-file problem. All three tools
receive explicit file paths as arguments, matching the declared inputs.

## Configuration

Ruff reads `ruff.toml` or `pyproject.toml` from the workspace root.
Bandit reads `.bandit` or `pyproject.toml`. Config files are not
declared as Bazel inputs (same limitation as other languages).

## Gavel-exclusive tools

**pycompile** is only available through Gavel. It provides a baseline
syntax check that catches import-breaking errors before other linters
run. Marked in `gavelExclusiveAspects` in `catalog/aspect.go`.

## Known limitations

1. **Config file tracking requires `gavel_lint_config` filegroup.**
   `gavel init` generates the filegroup with globs for `ruff.toml`,
   `.bandit`, and `pyproject.toml`.
2. **No virtualenv support in Ruff aspect.** Ruff analyzes files in
   isolation without access to installed packages. Type-checking rules
   that require import resolution may not work.
3. **Bandit site-packages is pinned.** The Bandit version and its
   plugins are pinned in `lint/lang/python/bandit/repositories.bzl`.

## Wrapper reference

| Wrapper | Flags |
|---------|-------|
| `lint/lang/python/ruff/wrapper/main.go` | `--ruff <binary> --out <sarif> <files...>` |
| `lint/lang/python/bandit/wrapper/main.go` | `--site-packages <dir> --out <sarif> <files...>` |
| `lint/lang/python/pycompile/wrapper/main.go` | `--out <sarif> <files...>` |
