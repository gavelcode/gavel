---
title: Incrementality decision record
type: explanation
description: Why Gavel relies on Bazel's action cache for incrementality instead of rdeps or diff-based scoping.
---

# Incrementality Decision Record

## Status

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | Baseline comparison via `--diff-base` | DONE |
| Phase 2 | Make baseline the default + committed baseline files | DONE |
| Phase 3 | Server-based baseline (shared/CI) | DONE (CLI fetches via `--server`) |

> **Implementation note (updated):** two decisions below were reversed once
> the feature landed. Coverage **is** stored in the baseline today — an overall
> `coverage` value plus a per-file `coverage.json` (see [baseline.md](../baseline.md))
> — used for the coverage delta. And "coverage on new code" is **implemented**
> (the `new_code_coverage` gate rule). The original rationale (avoid an escapable
> window) is preserved for the *gate* threshold, which remains absolute; the
> stored numbers feed the delta/diff, not the pass/fail gate.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Target scoping (rdeps) | Eliminated | Benchmarks: 71s query > 30-60s Bazel cache verification |
| Baseline default | Always on (unless `--absolute`) | Developer shouldn't need flags for the common case |
| Baseline storage | `.gavel/baseline/<project>/` committed to git | Team shares same baseline (ArchUnit pattern) |
| Baseline format | Plain text, one ID per line, sorted | Maximizes git auto-merge |
| Findings fingerprint | From SARIF (or fallback `sha256(tool:rule:file:line)`) | Already implemented in SARIF parser |
| Architecture violation ID | `rule:sourcePkg:targetPkg` | Same stability guarantee as findings fallback |
| Coverage baseline | None — uses `git merge-base` at runtime | Coverage is a gradient, not a set; storing % creates escapable window |
| Coverage on new code | `git diff --unified=0` intersected with LCOV `DA:` records | SonarQube pattern; ~300 lines of implementation |
| Scalar metrics (coverage %, counts) | Not stored in baseline | Derivable at runtime; avoids merge conflicts |
| CI baseline | Deferred to Phase 3 (server) | Current use is local-only |

## The core insight: Bazel's cache IS the scoping mechanism

### Benchmark: mm-monorepo (136,644 targets)

| Operation | Time | Notes |
|-----------|------|-------|
| `bazel query //...` (cold) | 46s | Full graph load, one-time cost |
| `bazel query //...` (warm) | 5.5s | Cached graph |
| `bazel query "rdeps(//mas-business/..., set(file))"` | 71s | Single file, scoped to one project (190 targets) |
| `bazel build --aspects` (warm cache, no changes) | 30-60s | Cache verification only |
| `bazel coverage` (cold, monolithic project) | ~20 min | Full test execution + instrumentation |
| `bazel coverage` (warm cache, no changes) | ~30-60s | All cache hits, no test execution |

### Why rdeps is counterproductive

1. rdeps query costs 71s for a single project scope (190 targets)
2. Warm Bazel cache verifies those 190 targets in 30-60s (cache hit = no work)
3. Net result: rdeps ADDS overhead vs running the full pattern with cache

rdeps only helps in cold-cache scenarios (CI without remote cache), but:
- Cold cache is a one-time event
- The query cost partially offsets savings
- With remote cache (standard in CI), cold starts are rare

**Conclusion**: Bazel's action cache IS the scoping mechanism. No rdeps needed.

## Baseline format

### Directory structure

```
.gavel/baseline/<project>/
├── findings        # one fingerprint per line, sorted
└── architecture    # one violation ID per line, sorted
```

Committed to git. Team shares the same baseline.

### `findings` file

One fingerprint per line, lexicographically sorted:

```
0a1b2c3d4e5f6789...
1f2e3d4c5b6a7890...
a0b1c2d3e4f56789...
```

Source: SARIF `fingerprints`/`partialFingerprints` fields when available.
Fallback: `sha256(tool:ruleID:filePath:line)` (computed in `core/infrastructure/casefile/sarif/parser.go`).

### `architecture` file

One violation ID per line, format `rule:sourcePkg:targetPkg`, sorted:

```
layer_violation:com.example.infra.persistence:com.example.domain.model
layer_violation:com.example.interfaces.http:com.example.infrastructure
```

Stability: breaks on package rename. Acceptable — a rename is a conscious
architectural decision that merits re-evaluation of violations.

### Why no `meta.json`

- `findings_count` = `wc -l findings`
- `coverage_percent` = computed from LCOV at runtime
- `commit` = derivable from `git log` on the baseline files

No stored value = no merge conflict on scalars.

## Coverage: two independent mechanisms

### Quality gate: overall coverage threshold

```yaml
quality_gate:
  min_coverage: 60    # applies to total LCOV percentage
```

Already implemented. No baseline needed — it's an absolute threshold.

### Quality gate: coverage on new code

```yaml
quality_gate:
  min_new_code_coverage: 80    # applies to changed lines only
```

**Implemented** (the `new_code_coverage` gate rule). Algorithm:

1. `git diff --unified=0 $(git merge-base <default-branch> HEAD)...HEAD` → changed lines per file
2. LCOV `DA:<line>,<hitcount>` records → coverable lines per file with hit counts
3. Intersection: changed lines that appear in LCOV = "new coverable lines"
4. Result: `covered / coverable * 100`

Key design choice: **diff base is `git merge-base`, NOT the baseline files.**
This prevents the "escape by re-running" problem — the window is anchored to
the branch point, not to the last gavel run.

LCOV handles "what's coverable" automatically: lines without `DA:` records
(imports, comments, declarations) are skipped. No heuristics needed.

### Implementation estimate: ~300 lines

| Component | Lines | Location |
|-----------|-------|----------|
| Extend LCOV parser with `DA:` records | ~30 | `core/infrastructure/casefile/lcov/` |
| Git diff parser (hunk headers → line numbers) | ~60 | `core/infrastructure/platform/git/` |
| Intersection logic | ~30 | New application use case |
| Types / value objects | ~20 | |
| Tests | ~150 | |

## CLI interface

```bash
# Default: baseline mode (evaluates only NEW findings/violations)
gavel judge
gavel judge --project=payments

# Absolute mode: evaluates ALL findings (release gates, nightly)
gavel judge --absolute

# Composable
gavel judge --quick                # findings only, baseline mode
gavel judge --quick --absolute     # findings only, absolute mode
```

### First run (no baseline exists)

```
$ gavel judge
  analyzing: payments (//payments/...)
  · no previous baseline — evaluating all findings
  findings: 142 errors, 38 warnings
  verdict: FAIL (max_error: 0, found: 142)
  baseline saved: .gavel/baseline/payments/ (180 fingerprints)
```

### Subsequent run (baseline active)

```
$ gavel judge
  analyzing: payments (//payments/...)
  · baseline: 180 fingerprints
  findings: 143 errors (+1 new), 38 warnings
  verdict: FAIL (max_error: 0, new errors: 1)
  baseline updated: .gavel/baseline/payments/ (181 fingerprints)
```

### All new findings fixed

```
$ gavel judge
  analyzing: payments (//payments/...)
  · baseline: 181 fingerprints
  findings: 142 errors (0 new, 1 fixed), 38 warnings
  verdict: PASS (max_error: 0, new errors: 0)
  baseline updated: .gavel/baseline/payments/ (180 fingerprints)
```

## Two worlds: CLI vs Server

### World 1: CLI (current, local + committed baseline)

- Baseline source: `.gavel/baseline/` committed to git
- Shared: yes, via git (team sees same baseline)
- Coverage on new code: `git merge-base` at runtime
- Suitable for: local development, teams without server

### World 2: Server (implemented)

- Baseline source: database (last successful run on default branch)
- Shared: yes, authoritative
- Suitable for: CI gates, dashboards, cross-repo visibility
- The CLI fetches the baseline via `--server URL --token TOKEN` and submits
  results back; it falls back to the committed local baseline when the server
  is unreachable.

## Supersedes

This record replaced two earlier design notes (since removed):

- **Scoped analysis** (rdeps / bazel-diff target scoping) — eliminated; the
  benchmarks above proved it counterproductive, and Bazel's action cache
  provides sufficient incrementality.
- **Baseline strategy** — its "Bazel-aware git diff" phase became "committed
  baseline files + default baseline mode" (now the live behaviour; see
  [baseline.md](../baseline.md)).
