---
title: Baseline strategy, new vs existing findings
type: explanation
description: How new findings are separated from existing debt. Partially superseded — Bazel-aware diff scoping was dropped after benchmarking.
---

# Baseline Strategy: New vs Existing Findings

> **Partially superseded.** Strategy 2 (Bazel-aware git diff / rdeps scoping)
> was eliminated after benchmarking — Bazel's action cache provides sufficient
> incrementality without rdeps queries. Strategy 2 is now "committed baseline
> files + default baseline mode." See
> [incrementality-decision.md](incrementality-decision.md) for the current
> design and rationale.

## Problem

`gavel.yaml` exposes `max_new_findings` but there is no concept of "new."
Every `judge` run starts with empty in-memory repositories. All findings
are treated as new. A project with 500 existing findings can never pass a
quality gate — Gavel only works for greenfield projects at zero findings.

This is SonarQube's core value proposition ("new code period") and Gavel's
biggest adoption blocker.

## Current state

**Strategy 1: DONE.** `max_new_findings` renamed to `max_findings`.

**Local snapshots: DONE.** `gavel judge` saves fingerprint-based snapshots
to `.gavel/snapshots/<project>.json` after each run. On subsequent runs,
it computes a delta (new/fixed/existing counts, findings delta, coverage
delta) by comparing current fingerprints against the previous snapshot.
Terminal output shows progress arrows and the delta summary. JSON output
includes full delta data. Individual findings are tagged with `isNew` when
their fingerprint wasn't in the previous snapshot. This is a local-only,
per-machine baseline — not shared across CI runners.

**Strategy 2: DONE (redesigned).** rdeps scoping eliminated after
benchmarking. Replaced with committed baseline files in `.gavel/baseline/`
and default-on baseline comparison mode. See
[incrementality-decision.md](incrementality-decision.md).

**Strategy 3 (server fingerprint comparison): NOT STARTED.**

## Three strategies (incremental, not exclusive)

### Strategy 1: Honest totals (no baseline) — DONE

Renamed `max_new_findings` to `max_findings` in the YAML schema. Quality
gate evaluates total counts. No comparison, no lie.

This is the fallback when no baseline source is available.

Local snapshots provide run-over-run progress tracking on top of this
(fingerprint-based new/fixed/existing classification), but the quality gate
still evaluates totals.

### Strategy 2: Bazel-aware git diff (local baseline) — SUPERSEDED

> **Superseded by [incrementality-decision.md](incrementality-decision.md).**
> rdeps scoping was eliminated (71s query overhead > 30-60s Bazel cache
> verification). Strategy 2 is now committed baseline files in
> `.gavel/baseline/` with default-on baseline comparison mode.

**When:** `--diff-base=main` flag (or auto-detected in CI via env vars).

Pipeline:
1. `git diff --name-only <base>...HEAD` — changed files
2. `bazel query "rdeps(//..., set(<changed_files>))"` — affected targets
3. Run aspects only on affected targets
4. All findings from affected targets = "new scope"

"New" means: "finding in a target affected by your changes."

**Advantages over file-only diff:**
- Understands the dependency graph — changing a library flags all consumers
- Faster — only runs aspects on what changed, not the whole monorepo
- No server, no persistence, no network
- Natural fit for PR-based CI workflows

**Limitations:**
- A preexisting finding in a file you touched appears as "new"
- Not precise at the fingerprint level — scope-based, not identity-based
- Requires a git merge-base (CI usually has it; local worktrees have it)

**Design questions:**
- [ ] What Bazel query variant? `rdeps` on source files vs `rdeps` on packages
- [ ] How to handle generated files in the diff?
- [ ] Should the quality gate have separate thresholds for "diff scope" vs "full scope"?
- [ ] Auto-detect CI env vars (`GITHUB_BASE_REF`, `CI_MERGE_REQUEST_TARGET_BRANCH_SHA`, etc.)?

**Effort:** Medium. Git diff + Bazel query plumbing, new runner mode, quality
gate split between scoped/total.

### Strategy 3: Server fingerprint comparison (remote baseline)

**When:** `server.url` configured in `gavel.yaml`.

Pipeline:
1. CLI runs full analysis (all targets, all aspects)
2. CLI uploads CaseFile to server (`POST /api/judgments`)
3. Server compares fingerprints against last successful run on base branch
4. Server classifies findings as new / existing / resolved
5. Quality gate evaluates only new findings against `max_new_findings`

"New" means: "fingerprint not present in the baseline run."

**Advantages:**
- Most precise — identity-based comparison, not scope-based
- Handles all edge cases (moved findings, renamed files, same fingerprint)
- Central source of truth — team sees the same baseline
- Enables trends, history, dashboards (the web UI)
- `max_new_findings` is honest

**Limitations:**
- Requires server deployment and network access
- Requires the `server:` config to be wired (currently dead code, issue #13)
- Server must have a concept of "baseline run" (last successful main)
- Latency: upload + comparison round-trip

**Design questions:**
- [ ] What defines the baseline? Last successful run on default branch?
- [ ] How does the server know which branch is "main"? Per-project config?
- [ ] Should the CLI block waiting for server classification, or fire-and-forget?
- [ ] Offline fallback: if server is unreachable, fall back to strategy 1 or 2?

**Effort:** High. Requires wiring `server:` config, HTTP client in CLI,
classification endpoint on server, baseline tracking in DB.

## Bazel as differentiator

SonarQube doesn't understand the build graph. Gavel does. Strategy 2 is
a unique advantage:

```
changed files ──► affected targets ──► scoped analysis
```

Two approaches for computing affected targets:

### Option A: `bazel query rdeps` (built-in)

```bash
git diff --name-only main...HEAD > changed_files
bazel query "rdeps(//..., set($(cat changed_files)))"
```

Simple, no external dependencies. But only tracks source file changes —
misses attribute changes, toolchain updates, etc.

### Option B: `bazel-diff` (Tinder)

https://github.com/Tinder/bazel-diff

```bash
bazel-diff generate-hashes -w . -b bazel -so starting_hashes.json   # at base rev
bazel-diff generate-hashes -w . -b bazel -so final_hashes.json      # at head rev
bazel-diff get-impacted-targets -sh starting_hashes.json -fh final_hashes.json -o targets.txt
```

Hashes the full dependency graph (rule impl + attributes + source content).
More precise — detects toolchain changes, rule attribute changes, not just
source diffs. Battle-tested at Tinder with tens of thousands of targets.
Trade-off: external JAR dependency.

### Recommendation

Start with `bazel query rdeps` (zero dependencies). Migrate to `bazel-diff`
if precision becomes a problem in large monorepos.

---

This is faster (fewer targets analyzed) AND more precise (dependency-aware)
than file-only diff. It's what makes Gavel worth using over "just run
SonarQube."

Strategy 3 then adds fingerprint precision on top. The two compose:
- Strategy 2 scopes WHAT to analyze (speed)
- Strategy 3 scopes WHAT is new (precision)

In combination: run aspects on affected targets only (fast), upload to server,
server compares fingerprints (precise). Best of both worlds.

## Execution order

```
Phase 1 ─── Strategy 1: rename max_new_findings → max_findings    ✅ DONE
            Stop lying. Trivial change.

Phase 1b ── Local snapshots: fingerprint-based progress tracking   ✅ DONE
            Run-over-run delta (new/fixed/existing).
            .gavel/snapshots/<project>.json persistence.

Phase 2 ─── Strategy 2: committed baseline files + default-on       ✅ DONE
            See incrementality-decision.md.
            rdeps scoping eliminated; Bazel cache is the scoping mechanism.

Phase 3 ─── Strategy 3: server fingerprint comparison
            Full SonarQube parity. High effort.
            Wire server: config (fixes issue #13).

Phase 2+3 ─ Combined: scoped analysis + server comparison
            Speed of local + precision of server.
```

## Impact on gavel.yaml schema

```yaml
# Phase 1: honest
quality_gate:
  max_findings: 0          # was: max_new_findings

# Phase 2: diff-aware
quality_gate:
  max_findings: 0          # total (no diff-base)
  max_new_findings: 0      # scoped to diff (requires --diff-base)

# Phase 3: server-aware
quality_gate:
  max_findings: 0          # total (fallback)
  max_new_findings: 0      # server-classified new findings
```

## Impact on CLI flags

```
# Phase 1: no new flags
gavel judge                          # total findings only

# Phase 2
gavel judge --diff-base=main         # Bazel-scoped diff
gavel judge --diff-base=auto         # auto-detect from CI env

# Phase 3
gavel judge                          # auto: server if configured, else total
gavel judge --diff-base=main         # scoped analysis + server comparison
```
