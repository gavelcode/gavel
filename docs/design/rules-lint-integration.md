---
title: Gavel and rules_lint integration
type: explanation
description: Where aspect_rules_lint fits relative to Gavel's native SARIF aspects — a breadth add-on, not a substitute.
---

# Integration Strategy: Gavel + rules_lint

## Context

Gavel currently runs its own Bazel aspects to execute linters (PMD, SpotBugs,
golangci-lint, etc.) and then evaluates the results. This works for standalone
usage but creates a problem: monorepos that already use `rules_lint` from
Aspect Build would run the same linters twice — once via `rules_lint` during
build, once via Gavel during `judge`.

## Problem

Duplicating linter execution is:
- **Wasteful** — same analysis runs twice, doubling CI time
- **Confusing** — version skew between Gavel's aspects and rules_lint config
  can produce different results for the same linter
- **A barrier to adoption** — teams already invested in rules_lint won't add
  a tool that fights their existing setup

## Insight: Gavel's value is evaluation, not execution

| Layer | What it does | Who should own it |
|-------|-------------|-------------------|
| **Execution** | Run linters, produce findings | rules_lint, Gavel aspects, or any runner |
| **Normalization** | Convert findings to SARIF | Linter itself or runner |
| **Evaluation** | Aggregate findings, apply quality gate, produce verdict | Gavel |
| **Governance** | Track trends, enforce policy, dashboard | Gavel server |

Gavel's unique value is in the bottom two layers. The top two are commodity —
rules_lint already does them well, with Bazel caching as a bonus.

## rules_lint output format

rules_lint runs linters as Bazel aspects. Linters that support SARIF produce
machine-readable reports stored in `bazel-bin/`. Reports can be found at:

```bash
find $(bazel info bazel-bin) -name "*AspectRulesLint*report"
```

Gavel already has a SARIF 2.1.0 parser (`core/infrastructure/sarif/`). The
normalization layer is already solved.

## Proposed architecture: dual mode

### Mode 1: Standalone (current behavior, fallback)

Gavel runs its own aspects, collects SARIF, evaluates. For teams that don't
use rules_lint and want a zero-config experience.

```
gavel judge
  └── bazel build --aspects @gavel_tools//lint/aspects:defs.bzl%java_pmd_submission_aspect ...
        └── produces SARIF
              └── Gavel parses + evaluates + verdict
```

### Mode 2: Integrated (new, for rules_lint users)

Gavel reads SARIF reports that rules_lint already produced. No duplicate
linter execution.

```
bazel lint //...              (user already runs this — rules_lint)
  └── produces SARIF reports in bazel-bin/

gavel judge                   (reads existing reports)
  └── finds *AspectRulesLint*report in bazel-bin/
        └── Gavel parses + evaluates + verdict
```

### Mode detection

Automatic, based on what's available:

1. Check for existing rules_lint reports in `bazel-bin/`
2. If found → Mode 2 (integrated, skip own aspects)
3. If not found → Mode 1 (standalone, run own aspects)

Override via config or flag:

```yaml
# gavel.yaml
findings_source: auto    # auto | gavel | rules_lint
```

```bash
gavel judge --findings-source=rules_lint   # force integrated mode
gavel judge --findings-source=gavel        # force standalone mode
```

## Tool coverage comparison

| Tool | Language | Gavel aspects | rules_lint |
|------|----------|---------------|------------|
| golangci-lint | Go | Yes | Yes |
| PMD | Java/Kotlin | Yes | Yes |
| SpotBugs | Java/Kotlin | Yes | Yes |
| Error Prone | Java/Kotlin | Yes | No |
| CPD | Java/Kotlin | Yes | No |
| Ruff | Python | Yes | Yes |
| Bandit | Python | Yes | Yes |
| pycompile | Python | Yes | No |
| ESLint | TypeScript | Yes | Yes |
| Clippy | Rust | Yes | Yes |
| Checkstyle | Java | No | Yes |
| ktlint | Kotlin | No | Yes |
| clang-tidy | C/C++ | No | Yes |
| shellcheck | Shell | No | Yes |
| Vale | Markdown | No | Yes |
| RuboCop | Ruby | No | Yes |
| yamllint | YAML | No | Yes |

In integrated mode, Gavel gains access to all rules_lint tools without
maintaining its own aspects for them.

Tools exclusive to Gavel (Error Prone, CPD, pycompile) would still need
Gavel's own aspects as a supplement. These could run alongside rules_lint
without conflict since they cover different tools.

## What Gavel adds that rules_lint cannot

Even in integrated mode, Gavel provides:

- **Quality gate with thresholds** — "allow up to 5 warnings, zero errors"
- **Multi-tool aggregation** — unified verdict across PMD + SpotBugs + ESLint
- **Coverage enforcement** — min coverage thresholds from LCOV
- **Architecture validation** — layer dependency rules
- **Enforcement levels** — blocking vs warning vs advisory per rule
- **Structured verdict** — pass / pass_with_warnings / fail with per-rule rulings
- **Historical tracking** — trends over time (server mode)
- **Dashboard** — web UI for cross-project visibility (server mode)
- **Baseline comparison** — new vs existing findings (see baseline-strategy.md)

rules_lint answers: "does this code have lint errors?"
Gavel answers: "is this code ready to ship?"

## Implementation plan

### Phase 1: Read existing SARIF files

Add a SARIF file discovery mechanism to the judge command. Instead of only
getting SARIF from `runner.RunAspect()`, also accept pre-existing SARIF
files from a directory (e.g., `bazel-bin/`).

Changes:
- New `SARIFCollector` interface with two implementations:
  - `AspectRunner` (current: runs aspects, returns SARIF bytes)
  - `ReportReader` (new: reads SARIF files from bazel-bin)
- `collectFindings` accepts either source
- `--findings-source` flag on judge command

### Phase 2: Auto-detection

Detect rules_lint reports automatically:
- Run `bazel info bazel-bin` to get the output directory
- Glob for `*AspectRulesLint*report` files
- If found and `findings_source: auto`, use them
- Map report filenames to tool names for evidence metadata

### Phase 3: Hybrid mode

For tools only in Gavel (Error Prone, CPD, pycompile):
- Run Gavel aspects only for tools not covered by rules_lint reports
- Merge results from both sources before evaluation

## Open questions

- [ ] Is the `*AspectRulesLint*report` glob pattern stable across rules_lint
      versions? Need to verify with rules_lint maintainers.
- [ ] Do all rules_lint linters produce SARIF, or only some? Linters without
      SARIF support may produce tool-specific formats.
- [ ] Should Gavel trigger `bazel lint //...` itself in integrated mode, or
      require the user to have run it beforehand?
- [ ] How to handle staleness: if rules_lint reports are from a previous
      commit, Gavel would evaluate outdated results.
