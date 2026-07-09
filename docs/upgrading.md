---
title: Upgrading gavel_tools
type: how-to
description: How to move a workspace between gavel_tools versions during the 0.x series, and why the version is pinned to the gavel CLI you run.
tags: [upgrade, versioning, gavel_tools, bazel]
---

# Upgrading gavel_tools

`gavel_tools` is the Bazel module that provides the analyzer aspects. Your
workspace pins it with one line in `MODULE.bazel`, written by `gavel init`:

```starlark
bazel_dep(name = "gavel_tools", version = "0.3.6")
```

## Why it is pinned to your gavel release

During 0.x the aspect API is not stable, so a gavel release is validated
against exactly **one** `gavel_tools` version — not a range. That pin is the
source of truth: `installer.gavelToolsVersion` is what `gavel init` writes, and
`installer/version_test.go` fails the build if it ever drifts from the
`gavel_tools` version gavel itself depends on. Treat the two as one unit and do
not mix an arbitrary `gavel_tools` with a gavel binary — the combination is
untested. (This is also why BCR publication waits for 1.0: a stable aspect API
is the pre-GA signal.)

## Moving to a new version

1. **Upgrade the gavel binary** ([releases](https://github.com/gavelcode/gavel/releases)).
2. **Match the pin.** Set the `version` in your `MODULE.bazel` `gavel_tools`
   `bazel_dep` to the one your new gavel writes. To read it without guessing,
   run `gavel init` in a throwaway directory and copy the line it emits.
3. Build once so Bazel updates `MODULE.bazel.lock` and resolves the module from
   the gavel registry (already on your `.bazelrc` as
   `--registry=https://gavelcode.github.io/registry`).

`gavel init` is for first-time setup: it *adds* the pin when absent but does not
rewrite an existing one, so re-running it in place will not migrate the version
for you — edit the `version` string yourself.

## Baselines survive the upgrade

Committed baselines in `.gavel/baseline/` are keyed by fingerprint, not by tool
version, so they carry across a `gavel_tools` bump. A version that changes an
analyzer's output will surface as new-vs-baseline deltas on the next `judge` —
which is the point: review them, then let the passing verdict ratchet the
baseline forward. See [Baseline](baseline.md).
