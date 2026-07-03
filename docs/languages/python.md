---
title: Python
type: reference
description: Python analysis via Ruff, Bandit and pycompile, run hermetically as Bazel aspects.
resource: https://github.com/gavelcode/gavel-tools/tree/main/lint/lang/python
tags: [python, ruff, bandit, analyzers]
---

# Python

Gavel analyzes Python with three tools, run as hermetic Bazel aspects from
gavel-tools.

| | |
|---|---|
| **Lint tools** | Ruff, Bandit, pycompile |
| **Aspects** | `python_ruff_submission_aspect`, `python_bandit_submission_aspect`, `python_pycompile_submission_aspect` |
| **Archtest** | `python_archtest_submission_aspect` |
| **Bazel targets** | `py_library`, `py_binary`, `py_test` |
| **SARIF suffixes** | `.ruff.sarif`, `.bandit.sarif`, `.pycompile.sarif` |
| **Coverage** | LCOV via `bazel coverage` |

## What each tool catches

- **Ruff** — a very fast Rust-written linter that replaces flake8, isort,
  pycodestyle, pyflakes and more; gavel's primary Python linter.
- **Bandit** — PyCQA's security analyzer: hardcoded passwords, `eval()`,
  insecure hashes, SQL-injection patterns.
- **pycompile** — syntax validation via Python's `py_compile`, a baseline check
  that catches import-breaking errors before the other tools run.

## How it runs

Each tool receives explicit `.py` file paths, so — unlike Go — there is no
test-file cache problem; Bandit additionally gets its pinned `site-packages` as
declared inputs for its rule plugins. The mechanics live in gavel-tools
([the hermetic analyzer driver](https://github.com/gavelcode/gavel-tools/blob/main/docs/tier-model.md)).

## Configuration

Ruff reads `ruff.toml` or `pyproject.toml`; Bandit reads `.bandit` or
`pyproject.toml`. Which tools run for a project is its `gavel.yaml` `tooling`
selection (see [configuration](../configuration.md)).
