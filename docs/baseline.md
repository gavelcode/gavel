---
title: Baseline
type: explanation
description: How Gavel separates new findings from existing debt — what the baseline is, when it is created, and the pass and ratchet update rules.
---

# Baseline

The baseline is Gavel's mechanism for distinguishing new findings from
existing ones. It enables "hold the line" — new code must be clean, but
existing debt is tracked separately and reduced over time.

## What it is

A baseline is a snapshot stored in `.gavel/baseline/<project>/` containing
three files:

| File | Content |
|------|---------|
| `findings` | One fingerprint per line — identifies each known finding |
| `architecture` | One violation ID per line (`rule:sourcePkg:targetPkg`) |
| `coverage` | A single number — coverage percentage at snapshot time |

These files are plain text and can be inspected directly.

## When it is created

The baseline is created automatically on the **first `gavel judge` run**,
regardless of the verdict. Even if the quality gate fails (e.g., coverage
below threshold), the baseline is seeded with the current state so that
delta tracking works from the second run onward.

## When it is updated

The update behavior depends on the verdict:

| Verdict | Behavior | Effect |
|---------|----------|--------|
| **pass** | Full update | Baseline is replaced with the current state (findings, violations, coverage) |
| **fail** | Ratchet | Only resolved items are removed. New items are NOT added. Coverage is preserved from the previous baseline |

The ratchet ensures the baseline can only **shrink** on failure — it
never grows. If you introduce 10 new findings and the verdict fails,
those 10 findings stay outside the baseline and will be reported as
"new" on every subsequent run until you fix them or the verdict passes.

When `--absolute` is used, the baseline is not updated at all.

## Per-branch baselines

Baselines are keyed by branch. `main` and `feature-x` have independent
baselines. In practice:

- The **quality gate** (tracking filter) compares findings against the
  project's default branch baseline — typically `main`. This means a PR
  branch sees findings as "new" if they are not in `main`'s baseline.
- The **baseline update** saves to the current branch's baseline. A PR
  branch updates its own baseline, not `main`'s.

This mirrors the "new code period" pattern used by SonarQube: the main
branch is the reference, feature branches are evaluated against it.

## Sharing the baseline

Commit `.gavel/baseline/` to git:

```bash
git add .gavel/baseline/
git commit -m "chore: update gavel baseline"
```

This ensures the entire team and CI share the same reference point.

## Resetting the baseline

To start fresh, delete the baseline directory and re-run:

```bash
rm -rf .gavel/baseline/<project>/
gavel judge
```

The first run after reset creates a new baseline with the current state.

## Baseline with server mode

When using `--server URL`, the CLI fetches the baseline from the server
before analysis and submits results back afterward. The server maintains
a centralized baseline that persists across CI runs without requiring
git commits of the baseline files.

Local baseline files (`.gavel/baseline/`) are still used as a fallback
when the server is unreachable.
