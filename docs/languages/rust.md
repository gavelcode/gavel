---
title: Rust
type: reference
description: Rust analysis via Clippy, delegated to the hermetic rules_rust aspect and converted to SARIF.
resource: https://github.com/gavelcode/gavel-tools/tree/main/lint/lang/rust
tags: [rust, clippy, analyzers]
---

# Rust

Gavel analyzes Rust with [Clippy](https://doc.rust-lang.org/clippy/), the
official linter. Unlike the other languages, gavel does not run Clippy itself: it
**delegates to `rules_rust`'s own hermetic
[`rust_clippy_aspect`](https://bazelbuild.github.io/rules_rust/rust_clippy.html)**
and only converts its output to SARIF.

| | |
|---|---|
| **Lint tool** | Clippy (from the Rust toolchain via `rules_rust`) |
| **Upstream aspect** | `@rules_rust//rust:defs.bzl%rust_clippy_aspect` |
| **Gavel aspect** | `rust_clippy_submission_aspect` (converter only) |
| **Archtest** | `rust_archtest_submission_aspect` |
| **Bazel targets** | `rust_library`, `rust_binary`, `rust_test` |
| **SARIF suffix** | `.clippy.sarif` |
| **Coverage** | LCOV via `bazel coverage` |

## How it runs

`rules_rust`'s aspect runs `rustc` with all inputs declared and writes Clippy's
JSON diagnostics; gavel's `rust_clippy_submission_aspect` runs after it
(`requires = [rust_clippy_aspect]`), reads those diagnostics and converts them to
SARIF. Because `rules_rust` already declares every input — sources, deps,
toolchain, `clippy.toml` — this is the cleanest caching story of any language,
with no gavel-owned linter subprocess. The converter lives in gavel-tools
([`lint/lang/rust/clippy/converter`](https://github.com/gavelcode/gavel-tools/tree/main/lint/lang/rust/clippy)).

## Configuration

Clippy reads `clippy.toml` (or `.clippy.toml`) from the workspace root, declared
by `rules_rust` as an aspect input so edits invalidate the cache. `gavel init`
also generates the `rules_rust` settings that make Clippy emit structured
diagnostics:

```
build:gavel-rust-clippy --@rules_rust//rust/settings:capture_clippy_output=True
build:gavel-rust-clippy --@rules_rust//rust/settings:clippy_output_diagnostics=True
```

Without these, the converter receives text output and produces zero findings.
Clippy's version is tied to the workspace's Rust toolchain — it is not separately
pinned.

## rules_rs consumers

Consumers that build Rust with [`rules_rs`](https://github.com/dzbarsky/rules_rs)
instead of `rules_rust` get a different aspect, published by the
`gavel_tools_rust_rs` module in place of the default `gavel_tools_rust`. Gavel never imposes a ruleset: when your root
`MODULE.bazel` declares `bazel_dep(name = "rules_rs", ...)`, `gavel init` adds
the variant module to `.gavel/gavel.MODULE.bazel` and points
`gavel-rust-clippy` at its aspect, and `gavel judge` resolves the same variant.
No gavel configuration is needed. Which dep selects which variant, and its build
flags, are declared by the `variants` entries of the gavel-tools catalog (see
`docs/catalog.md` in gavel-tools); detection scans the literal `bazel_dep` calls
in `MODULE.bazel` and the files it `include()`s.
