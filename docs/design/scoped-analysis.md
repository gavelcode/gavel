---
title: Scoped analysis and the hybrid scoping model
type: explanation
description: Superseded by the incrementality decision — benchmarks showed rdeps scoping is counterproductive on large monorepos.
---

# Scoped Analysis: `--diff-base` and the Hybrid Scoping Model

> **Superseded by `incrementality-decision.md`.** Benchmarks on a 136k-target
> monorepo proved that `rdeps` scoping is counterproductive — Bazel's action
> cache provides sufficient incrementality. The baseline comparison approach
> (committed fingerprint files + delta computation) was implemented instead.

## Problem

`gavel judge` analyzes every target in the project's `pattern` field. In a
monorepo with 12,000 targets, this means:

| Phase | Command | What happens | Time |
|-------|---------|-------------|------|
| Findings | `bazel build --aspects` per language | Runs PMD/SpotBugs/etc. on every source file | 5-10 min |
| Coverage | `bazel coverage --combined_report=lcov` | Compiles + instruments + executes every test | 15-30 min |
| Architecture | `bazel test //arch:arch_test` | Single test target | <10 sec |

Findings benefit from Bazel's action cache: unchanged sources produce cache
hits, so incremental runs are fast. But **coverage always executes tests** —
there is no cache shortcut for test execution when inputs haven't changed
and the test binary needs to run for instrumentation. On a PR that touches
3 files, running coverage for 12,000 targets is 20+ minutes of waste.

SonarQube solves this with file-level diff ("new code period"). But it doesn't
understand the build graph — changing a shared library doesn't flag its
consumers. Gavel can do better.

## Solution: scoped analysis via `--diff-base`

```
gavel judge --diff-base=main
```

When `--diff-base` is set, Gavel restricts the analysis to targets affected
by the changes between the base ref and HEAD. The pipeline becomes:

```
git diff --name-only main...HEAD
        │
        ▼
   changed files
        │
        ▼
   target scoper  ◄── pluggable: rdeps (built-in) or bazel-diff or aspect impacted
        │
        ▼
   affected targets
        │
        ├──► bazel build --aspects (findings) ──► only affected targets
        ├──► bazel coverage (coverage)         ──► only affected test targets
        └──► bazel test (arch)                 ──► unchanged (single target)
```

Instead of `//payments/...` (thousands of targets), Gavel runs aspects and
coverage on perhaps 50-200 targets. A 20-minute run becomes 2 minutes.

## Hybrid scoping model: pluggable `TargetScoper`

Three implementations, detected automatically with manual override:

### Implementation 1: `rdeps` (built-in, zero dependencies)

```bash
git diff --name-only main...HEAD > /tmp/changed_files
bazel query "rdeps(//..., set($(cat /tmp/changed_files)))" 2>/dev/null
```

**How it works:** `bazel query rdeps(universe, set)` walks the dependency
graph upward from the changed source files, returning every target that
transitively depends on them. If you change `shared/auth/token.go`, it
returns `shared/auth`, `payments/api`, `billing/service`, etc. — every
package that imports `shared/auth`.

**Advantages:**
- Zero external dependencies. Ships with Bazel.
- Understands transitive dependencies (unlike file-only diff).
- Fast — `bazel query` reads the cached action graph, no builds needed.

**Limitations:**
- Only detects source file changes. Misses: rule attribute changes (e.g.,
  adding a `dep`), toolchain version bumps, `.bzl` macro changes.
- `rdeps(//..., ...)` loads the entire dependency graph. In very large
  monorepos (100k+ targets), this can take 30-60 seconds and significant
  memory.

**When to use:** default fallback. Good enough for most monorepos.

### Implementation 2: `bazel-diff` (Tinder, external JAR)

```bash
# At base revision:
bazel-diff generate-hashes -w . -b bazel -so /tmp/start.json

# At HEAD:
bazel-diff generate-hashes -w . -b bazel -so /tmp/end.json

# Diff:
bazel-diff get-impacted-targets \
    -sh /tmp/start.json -fh /tmp/end.json -o /tmp/targets.txt
```

**How it works:** hashes every node in the dependency graph — source
content, rule implementation, attributes, toolchain versions. Compares
hashes between two revisions. A changed hash means the target is impacted,
even if no source file changed (e.g., a toolchain version bump).

**Advantages:**
- Most precise source-change detection available.
- Catches attribute changes, toolchain bumps, macro changes, generated
  file changes — everything `rdeps` misses.
- Battle-tested at Tinder with tens of thousands of targets.

**Limitations:**
- External dependency: JAR file (requires JVM on the build machine).
- Two full graph hashes per invocation. In very large repos, hashing takes
  60-90 seconds (but still cheaper than running all tests).
- Not included with Bazel — user must install it separately.

**When to use:** large monorepos (10k+ targets) where precision matters,
or when toolchain/macro changes are frequent.

### Implementation 3: `aspect impacted` (Aspect CLI, AXL extension)

```bash
aspect impacted
```

**How it works:** wraps `bazel-diff` in an AXL extension for the Aspect
CLI. Same hash-based approach, but integrated into the `aspect` CLI
workflow. Available at https://github.com/aspect-extensions/impacted.

**Advantages:**
- Same precision as `bazel-diff`, better developer UX.
- Part of the Aspect ecosystem (works with `aspect lint`, `aspect test`).
- Starlark-based — no separate JAR management.

**Limitations:**
- Requires Aspect CLI (`aspect` binary) instead of plain `bazel`.
- Experimental (early-stage, API not yet stable).
- Adds a dependency on Aspect Build's tooling.

**When to use:** teams already using Aspect CLI.

### Detection and override

Gavel detects which scoper is available at runtime:

```
1. Is `--diff-base` set?
   └── No  → full analysis (current behavior, no scoping)
   └── Yes → continue

2. Is `--scoper` flag explicitly set?
   └── Yes → use that scoper (error if not available)
   └── No  → auto-detect:
       a. Is `aspect` binary in PATH? → use aspect impacted
       b. Is `bazel-diff` binary in PATH? → use bazel-diff
       c. Fallback → use rdeps
```

Override via flag:

```bash
gavel judge --diff-base=main                     # auto-detect scoper
gavel judge --diff-base=main --scoper=rdeps      # force rdeps
gavel judge --diff-base=main --scoper=bazel-diff  # force bazel-diff
gavel judge --diff-base=main --scoper=aspect      # force aspect impacted
```

## Coverage: the expensive phase

Findings aspects benefit from Bazel's action cache even without scoping.
Coverage does not — `bazel coverage` must execute tests for instrumentation.
This is the primary motivation for scoped analysis.

### Current behavior

```go
runner.RunCoverage(ctx, workspace, targets)
// targets = []string{"//payments/..."}
// Runs: bazel coverage //payments/... --combined_report=lcov
```

All test targets under the pattern are executed, instrumented, and merged
into a single LCOV report.

### Scoped behavior

```go
runner.RunCoverage(ctx, workspace, affectedTestTargets)
// affectedTestTargets = []string{"//payments/api:api_test", "//shared/auth:auth_test"}
// Runs: bazel coverage //payments/api:api_test //shared/auth:auth_test --combined_report=lcov
```

Only tests that transitively depend on changed files are executed. The LCOV
report contains coverage only for affected code. This is correct for diff
analysis — we want to know "are the changes tested?", not "what's the
total project coverage?".

### Filtering test targets from affected targets

`bazel query rdeps` returns all affected targets — libraries, binaries,
and tests. Coverage needs only test targets. Filter:

```bash
bazel query "kind('.*_test', rdeps(//..., set(<changed_files>)))"
```

The `kind('.*_test', ...)` filter selects only targets whose rule type
ends in `_test` (e.g., `go_test`, `java_test`, `py_test`).

For `bazel-diff` and `aspect impacted`, the output includes all target
types. Gavel filters post-hoc:

```bash
bazel query "kind('.*_test', set(<impacted_targets>))"
```

### Future: `aspect coverage` (incremental, git-aware)

Aspect Build's experimental coverage extension
(https://github.com/aspect-extensions/coverage) provides:

- **Git-aware incremental coverage**: shows how many of the added/edited
  lines are tested, rather than total coverage percentage.
- **No analysis cache invalidation**: unlike `bazel coverage`, it doesn't
  bust the analysis cache with coverage-specific flags.
- **LCOV output**: CI-friendly, same format Gavel already parses.

When `aspect coverage` stabilizes, Gavel can delegate coverage collection
to it (same pattern as `aspect impacted` for scoping). The detection logic
extends naturally:

```
Is `aspect` binary available AND aspect-coverage extension installed?
  └── Yes → use aspect coverage (incremental, git-aware)
  └── No  → use bazel coverage on scoped targets (current approach)
```

This is NOT a dependency — it's an optional enhancement. `bazel coverage`
on scoped targets remains the default.

## Interaction with existing features

### Quality gate evaluation

Two modes:

| Mode | Scope | Quality gate evaluates | Use case |
|------|-------|----------------------|----------|
| Full (`gavel judge`) | All targets | Total findings | Nightly CI, release gate |
| Scoped (`gavel judge --diff-base=main`) | Affected targets | Scoped findings | PR check, local dev |

The quality gate rules (`max_error`, `max_warning`, etc.) apply to whichever
scope is active. In scoped mode, 0 errors in affected targets = pass, even
if unaffected code has errors.

This is intentional: a PR check should answer "did this PR introduce
problems?", not "does the entire monorepo have zero issues?".

### Snapshot delta

Snapshots track total project state (all fingerprints). In scoped mode:

- **Findings count**: scoped (only affected targets).
- **Fingerprints**: scoped (only from affected targets). The snapshot is
  NOT updated in scoped mode — it would lose fingerprints from unscoped
  targets.
- **Delta display**: shows new/fixed within scope, not project-wide.

Full-scope runs (`gavel judge` without `--diff-base`) continue to update
the snapshot. This preserves the "project health over time" story.

### `--quick` flag

`--quick` skips coverage and architecture. `--diff-base` scopes what gets
analyzed. They compose:

```bash
gavel judge --quick                       # all targets, findings only
gavel judge --diff-base=main              # affected targets, full analysis
gavel judge --diff-base=main --quick      # affected targets, findings only
```

`--diff-base=main --quick` is the fastest possible analysis: findings
aspects on affected targets only. Sub-minute on most PRs.

### `--project` flag

`--project` selects which project from `gavel.yaml` to analyze. `--diff-base`
scopes within that project's target pattern:

```bash
gavel judge --project=payments --diff-base=main
```

This intersects the project's targets (`//payments/...`) with the affected
targets. Only targets that are both under `//payments/` AND affected by the
diff are analyzed.

### JSON output

JSON output includes the scope metadata when `--diff-base` is active:

```json
{
  "projects": [{
    "name": "payments",
    "scope": {
      "mode": "diff",
      "base_ref": "main",
      "scoper": "rdeps",
      "affected_targets": 47,
      "affected_test_targets": 12,
      "changed_files": ["payments/api/handler.go", "shared/auth/token.go"]
    },
    "verdict": "pass",
    "findings_count": 3,
    ...
  }]
}
```

### CI auto-detection

`--diff-base=auto` detects the base ref from CI environment variables:

| CI system | Environment variable | Value |
|-----------|---------------------|-------|
| GitHub Actions | `GITHUB_BASE_REF` | `main` |
| GitLab CI | `CI_MERGE_REQUEST_TARGET_BRANCH_SHA` | commit SHA |
| Buildkite | `BUILDKITE_PULL_REQUEST_BASE_BRANCH` | `main` |
| Jenkins | `CHANGE_TARGET` | `main` |
| Generic | `GAVEL_DIFF_BASE` | any ref |

Detection order:
1. `GAVEL_DIFF_BASE` (explicit, takes precedence)
2. CI-specific variable (based on which CI is detected)
3. Error if `auto` but no CI variable found

## Implementation plan

### Interface: `TargetScoper`

```go
// cli/internal/bazel/scoper/scoper.go

type ScopedTargets struct {
    AllTargets      []string  // all affected targets (libraries + binaries + tests)
    TestTargets     []string  // only test targets (for coverage)
    ChangedFiles    []string  // git diff output
    BaseRef         string    // the base ref used
    Method          string    // "rdeps", "bazel-diff", or "aspect"
}

type TargetScoper interface {
    Scope(ctx context.Context, workspace string, baseRef string) (ScopedTargets, error)
}
```

The scoper computes ALL affected targets across the entire repo (`//...`).
Per-project filtering happens in `runProject()` via `filterTargetsByPattern()`.
This keeps the scoper simple — it doesn't need to know about Gavel's project
concept.

### Implementation: `RdepsScoper`

```go
// cli/internal/bazel/scoper/rdeps.go

type RdepsScoper struct{}

func (s *RdepsScoper) Scope(ctx context.Context, workspace, baseRef string) (ScopedTargets, error) {
    // 1. git diff --name-only <baseRef>...HEAD
    changedFiles, err := gitDiffFiles(ctx, workspace, baseRef)

    // 2. Empty diff = no affected targets (all projects will skip with pass)
    if len(changedFiles) == 0 {
        return ScopedTargets{BaseRef: baseRef, Method: "rdeps"}, nil
    }

    // 3. bazel query "rdeps(//..., set(<changedFiles>))"
    allTargets, err := bazelQueryRdeps(ctx, workspace, "//...", changedFiles)

    // 4. bazel query "kind('.*_test', set(<allTargets>))"
    testTargets, err := bazelQueryKindTest(ctx, workspace, allTargets)

    return ScopedTargets{
        AllTargets:   allTargets,
        TestTargets:  testTargets,
        ChangedFiles: changedFiles,
        BaseRef:      baseRef,
        Method:       "rdeps",
    }, nil
}
```

### Implementation: `BazelDiffScoper`

```go
// cli/internal/bazel/scoper/bazeldiff.go

type BazelDiffScoper struct{}

func (s *BazelDiffScoper) Scope(ctx context.Context, workspace, baseRef string) (ScopedTargets, error) {
    // 1. Generate hashes at base ref
    //    OPEN QUESTION: bazel-diff standard workflow requires checkout.
    //    Options: (a) git worktree for base revision hashing (safe),
    //    (b) bazel-diff -br flag if supported in current version,
    //    (c) git stash + checkout + restore (fragile, avoid).
    //    Decision deferred to Phase C implementation.
    startHashes, err := generateHashes(ctx, workspace, baseRef)

    // 2. Generate hashes at HEAD
    endHashes, err := generateHashes(ctx, workspace, "HEAD")

    // 3. Get impacted targets
    //    bazel-diff get-impacted-targets -sh start.json -fh end.json -o targets.txt
    allTargets, err := getImpactedTargets(ctx, startHashes, endHashes)

    // 4. Filter test targets
    testTargets, err := bazelQueryKindTest(ctx, workspace, allTargets)

    // 5. Get changed files (for JSON output metadata only — not used for scoping)
    changedFiles, err := gitDiffFiles(ctx, workspace, baseRef)

    return ScopedTargets{
        AllTargets:   allTargets,
        TestTargets:  testTargets,
        ChangedFiles: changedFiles,
        BaseRef:      baseRef,
        Method:       "bazel-diff",
    }, nil
}
```

### Implementation: `AspectScoper`

```go
// cli/internal/bazel/scoper/aspect.go

type AspectScoper struct{}

func (s *AspectScoper) Scope(ctx context.Context, workspace, baseRef string) (ScopedTargets, error) {
    // 1. aspect impacted (outputs impacted targets to stdout)
    //    NOTE: aspect impacted auto-detects the base ref from git.
    //    The baseRef parameter is passed if the CLI supports --base,
    //    but may be ignored by older versions. If aspect impacted
    //    does not support custom base refs, it silently uses its
    //    own detection — which may differ from the user's --diff-base.
    allTargets, err := runAspectImpacted(ctx, workspace, baseRef)

    // 2. Filter test targets
    testTargets, err := bazelQueryKindTest(ctx, workspace, allTargets)

    // 3. Get changed files (for JSON output metadata only — not used for scoping)
    changedFiles, err := gitDiffFiles(ctx, workspace, baseRef)

    return ScopedTargets{
        AllTargets:   allTargets,
        TestTargets:  testTargets,
        ChangedFiles: changedFiles,
        BaseRef:      baseRef,
        Method:       "aspect",
    }, nil
}
```

### Detection: `DetectScoper`

```go
// cli/internal/bazel/scoper/detect.go

func DetectScoper(explicit string) (TargetScoper, error) {
    if explicit != "" {
        switch explicit {
        case "rdeps":
            return &RdepsScoper{}, nil
        case "bazel-diff":
            if !binaryExists("bazel-diff") {
                return nil, fmt.Errorf("bazel-diff not found in PATH")
            }
            return &BazelDiffScoper{}, nil
        case "aspect":
            if !binaryExists("aspect") {
                return nil, fmt.Errorf("aspect CLI not found in PATH")
            }
            return &AspectScoper{}, nil
        default:
            return nil, fmt.Errorf("unknown scoper: %s", explicit)
        }
    }

    if binaryExists("aspect") {
        return &AspectScoper{}, nil
    }
    if binaryExists("bazel-diff") {
        return &BazelDiffScoper{}, nil
    }
    return &RdepsScoper{}, nil
}
```

### CI auto-detection: `ResolveBaseRef`

```go
// cli/internal/bazel/scoper/cidetect.go

func ResolveBaseRef(flagValue string) (string, error) {
    if flagValue != "auto" {
        return flagValue, nil
    }

    envVars := []struct {
        name     string
        ciSystem string
    }{
        {"GAVEL_DIFF_BASE", "gavel"},
        {"GITHUB_BASE_REF", "github-actions"},
        {"CI_MERGE_REQUEST_TARGET_BRANCH_SHA", "gitlab"},
        {"BUILDKITE_PULL_REQUEST_BASE_BRANCH", "buildkite"},
        {"CHANGE_TARGET", "jenkins"},
    }

    for _, ev := range envVars {
        if val := os.Getenv(ev.name); val != "" {
            return val, nil
        }
    }

    return "", fmt.Errorf("--diff-base=auto: no CI environment variable found; set GAVEL_DIFF_BASE explicitly")
}
```

### Changes to `judge.go`

#### New fields in `options`

```go
type options struct {
    // ... existing fields ...
    diffBase string  // --diff-base flag: "main", "auto", or any git ref
    scoper   string  // --scoper flag: "rdeps", "bazel-diff", "aspect", or "" (auto)
}
```

#### New flags

```go
cmd.Flags().StringVar(&opts.diffBase, "diff-base", "", "Scope analysis to targets affected by changes since this ref (e.g., main, auto)")
cmd.Flags().StringVar(&opts.scoper, "scoper", "", "Target scoping method: rdeps, bazel-diff, aspect (auto-detected if omitted)")
```

#### Modified `run()` flow

```go
const scopedAnalysisThreshold = 500

func run(cmd *cobra.Command, opts options, d deps) error {
    // ... existing: workspace, config, projects, git info ...

    // NEW: --scoper without --diff-base is an error
    if opts.scoper != "" && opts.diffBase == "" {
        return fmt.Errorf("--scoper requires --diff-base")
    }

    // NEW: resolve scoping if --diff-base is set
    var scope *scoper.ScopedTargets
    if opts.diffBase != "" {
        baseRef, err := scoper.ResolveBaseRef(opts.diffBase)
        if err != nil {
            return err
        }
        s, err := scoper.DetectScoper(opts.scoper)
        if err != nil {
            return err
        }
        // scope is resolved once, before the project loop
        // each project filters to its own pattern inside runProject
        resolved, err := s.Scope(ctx, workspace, baseRef)
        if err != nil {
            return fmt.Errorf("scope analysis: %w", err)
        }

        // Large diff fallback: if the scoped analysis returns more targets
        // than the threshold, the overhead of listing individual targets
        // exceeds the cost of a full pattern scan. Fall back to full
        // analysis with a warning.
        if len(resolved.AllTargets) > scopedAnalysisThreshold {
            d.log.Warn("diff affects too many targets, falling back to full analysis",
                "affected", len(resolved.AllTargets),
                "threshold", scopedAnalysisThreshold,
            )
            // scope remains nil → full analysis
        } else {
            scope = &resolved
            d.log.Debug("scoped analysis",
                "method", resolved.Method,
                "changed_files", len(resolved.ChangedFiles),
                "affected_targets", len(resolved.AllTargets),
                "affected_tests", len(resolved.TestTargets),
            )
        }
    }

    // ... existing: project loop, but pass scope to runProject ...
}
```

#### Modified `runProject()` flow

```go
func runProject(
    ctx context.Context,
    w io.Writer,
    d deps,
    workspace string,
    project projectmodel.Project,
    commitSHA, branch string,
    startedAt time.Time,
    silent bool,
    quick bool,
    scope *scoper.ScopedTargets,  // NEW parameter, nil = full analysis
) (projectResult, error) {

    targets := []string{project.TargetPattern()}

    // NEW: if scoped, replace targets with affected subset
    if scope != nil {
        targets = filterTargetsByPattern(scope.AllTargets, project.TargetPattern())
        if len(targets) == 0 {
            // No affected targets in this project — skip entirely
            return projectResult{
                name:    project.Name(),
                verdict: "pass",
            }, nil
        }
    }

    // ... existing: collectFindings with (possibly scoped) targets ...

    if !quick {
        coverageTargets := targets
        if scope != nil {
            coverageTargets = filterTargetsByPattern(scope.TestTargets, project.TargetPattern())
        }
        // ... existing: collectCoverage with coverageTargets ...
        // ... existing: collectArchtest (unchanged — single target, not scoped) ...
    }

    // ... existing: submit, snapshot ...

    // NEW: do NOT update snapshot in scoped mode (would lose unscoped fingerprints)
    if scope == nil {
        if err := saveSnapshot(workspace, project.Name(), current); err != nil {
            log.Warn("failed to save snapshot", "error", err)
        }
    }

    return pr, nil
}
```

### Helper: `filterTargetsByPattern`

```go
// cli/internal/bazel/scoper/filter.go

func filterTargetsByPattern(targets []string, pattern string) []string {
    // pattern is like "//payments/..." — extract the package prefix
    // "//payments/..." → "//payments/"
    prefix := strings.TrimSuffix(pattern, "...")

    var filtered []string
    for _, t := range targets {
        if strings.HasPrefix(t, prefix) {
            filtered = append(filtered, t)
        }
    }
    return filtered
}
```

### Helper: `gitDiffFiles`

```go
// cli/internal/bazel/scoper/git.go

func gitDiffFiles(ctx context.Context, workspace, baseRef string) ([]string, error) {
    var output bytes.Buffer
    cmd := exec.CommandContext(ctx, "git", "diff", "--name-only", baseRef+"...HEAD")
    cmd.Dir = workspace
    cmd.Stdout = &output
    cmd.Stderr = &output
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("git diff: %w\n%s", err, output.String())
    }

    lines := strings.Split(strings.TrimSpace(output.String()), "\n")
    var files []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" {
            files = append(files, line)
        }
    }
    return files, nil
}
```

### Helper: `bazelQueryRdeps`

```go
// cli/internal/bazel/scoper/rdeps.go (inside RdepsScoper file)

func bazelQueryRdeps(ctx context.Context, workspace, universe string, files []string) ([]string, error) {
    if len(files) == 0 {
        return nil, nil
    }

    // Build the set expression from file paths
    setExpr := strings.Join(files, " ")
    query := fmt.Sprintf("rdeps(%s, set(%s))", universe, setExpr)

    var output bytes.Buffer
    cmd := exec.CommandContext(ctx, "bazel", "query", query, "--keep_going", "--noshow_progress")
    cmd.Dir = workspace
    cmd.Stdout = &output
    cmd.Stderr = io.Discard  // bazel query emits warnings to stderr
    if err := cmd.Run(); err != nil {
        // bazel query returns exit code 3 for partial results with --keep_going
        var exitErr *exec.ExitError
        if !errors.As(err, &exitErr) || exitErr.ExitCode() != 3 {
            return nil, fmt.Errorf("bazel query rdeps: %w", err)
        }
    }

    return parseQueryOutput(output.String()), nil
}

func bazelQueryKindTest(ctx context.Context, workspace string, targets []string) ([]string, error) {
    if len(targets) == 0 {
        return nil, nil
    }

    setExpr := strings.Join(targets, " ")
    query := fmt.Sprintf("kind('.*_test', set(%s))", setExpr)

    var output bytes.Buffer
    cmd := exec.CommandContext(ctx, "bazel", "query", query, "--noshow_progress")
    cmd.Dir = workspace
    cmd.Stdout = &output
    cmd.Stderr = io.Discard
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("bazel query kind test: %w", err)
    }

    return parseQueryOutput(output.String()), nil
}

func parseQueryOutput(output string) []string {
    lines := strings.Split(strings.TrimSpace(output), "\n")
    var targets []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" && strings.HasPrefix(line, "//") {
            targets = append(targets, line)
        }
    }
    return targets
}
```

## File plan

| File | Action | Description |
|------|--------|-------------|
| `cli/internal/bazel/scoper/scoper.go` | Create | `TargetScoper` interface + `ScopedTargets` struct |
| `cli/internal/bazel/scoper/rdeps.go` | Create | `RdepsScoper` + `gitDiffFiles` + `bazelQueryRdeps` + `bazelQueryKindTest` |
| `cli/internal/bazel/scoper/bazeldiff.go` | Create | `BazelDiffScoper` |
| `cli/internal/bazel/scoper/aspect.go` | Create | `AspectScoper` |
| `cli/internal/bazel/scoper/detect.go` | Create | `DetectScoper` auto-detection |
| `cli/internal/bazel/scoper/cidetect.go` | Create | `ResolveBaseRef` CI env detection |
| `cli/internal/bazel/scoper/filter.go` | Create | `filterTargetsByPattern` |
| `cli/internal/command/judge/judge.go` | Modify | Add `--diff-base`, `--scoper` flags; pass scope to `runProject` |
| `cli/internal/bazel/scoper/*_test.go` | Create | Unit tests for each scoper, filter, CI detection |
| `cli/internal/command/judge/judge_test.go` | Modify | Tests for scoped runProject behavior |

## Execution phases

### Phase A: `rdeps` scoper + `--diff-base` flag (MVP)

Implement `RdepsScoper`, `DetectScoper` (rdeps-only), `ResolveBaseRef`,
wire `--diff-base` into judge command. This alone delivers the core value:
scoped analysis with zero external dependencies.

**Deliverable:** `gavel judge --diff-base=main` works with `rdeps`.

### Phase B: CI auto-detection

Implement `--diff-base=auto` with CI environment variable detection.

**Deliverable:** `gavel judge --diff-base=auto` works in GitHub Actions,
GitLab CI, Buildkite, Jenkins.

### Phase C: `bazel-diff` scoper

Implement `BazelDiffScoper`, extend `DetectScoper` to prefer it when
available.

**Deliverable:** `gavel judge --diff-base=main` auto-detects `bazel-diff`
and uses it for more precise scoping.

### Phase D: `aspect impacted` scoper

Implement `AspectScoper`, extend `DetectScoper` to prefer it when Aspect
CLI is available. Monitor `aspect-extensions/coverage` stability for
future coverage delegation.

**Deliverable:** `gavel judge --diff-base=main` auto-detects Aspect CLI
and uses `aspect impacted` for scoping.

## Edge cases

### Zero files changed

`git diff` returns empty → scoper returns `ScopedTargets` with empty
target lists → every project skips with `verdict: "pass"` (no affected
targets in any project's pattern). This is correct: no changes, no
findings, pass. The terminal output should indicate this:

```
  payments — no affected targets, skipped
```

### Large diff (merge commit, major refactor)

If `rdeps` returns more targets than `scopedAnalysisThreshold` (default
500), Gavel falls back to full analysis with a log warning. Scoped
analysis with 5,000 explicit targets is slower than a pattern scan
(`//...`) due to:
- Long command lines (OS argument length limits on some platforms)
- `bazel build` optimizes pattern expansion internally
- The query overhead itself consumed time for minimal benefit

### Cross-project impacts

A change in `shared/auth/` affects targets in `payments/`, `billing/`,
and `users/`. The scoper runs `rdeps(//..., ...)` which returns targets
across all projects. `runProject()` filters to each project's pattern,
so each project only sees its own affected subset. This is correct — the
impact is distributed automatically.

## Open questions

- [ ] How to handle `bazel query` failures gracefully? If the query
  times out in a very large repo, should Gavel fall back to full analysis
  with a warning?
- [ ] Should `--diff-base` affect architecture tests? The arch test is a
  single target and runs in seconds — scoping it adds complexity for
  negligible time savings.
- [ ] How does `--diff-base` interact with `gavel.yaml` `default_branch`
  field? Should `default_branch` be the implicit `--diff-base` value?
- [ ] `bazel-diff` revision handling: does `bazel-diff generate-hashes`
  support a `-br` flag for hashing at a different revision without
  checkout? If not, Gavel needs to use `git worktree` for safe base-ref
  hashing. This must be resolved before Phase C implementation.
- [ ] Does `aspect impacted` support a `--base` flag to override its
  auto-detected base ref? If not, `--diff-base=release/v2` would be
  silently ignored when using the aspect scoper. Verify before Phase D.
- [ ] What is the right `scopedAnalysisThreshold`? 500 is a guess.
  Profiling on a real monorepo would give a better number. Consider
  making it configurable via `gavel.yaml` or env var.

## Relationship to other design docs

- **baseline-strategy.md**: `--diff-base` is Strategy 2. This document
  details the implementation. Strategy 3 (server fingerprint comparison)
  remains independent and can layer on top.
- **rules-lint-integration.md**: the `SARIFCollector` / findings source
  abstraction is orthogonal to scoping. In integrated mode (reading
  rules_lint SARIF), Gavel would scope which reports to read based on
  affected targets, not which aspects to run.
