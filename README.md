# Gavel

**Order in the codebase.** Quality governance for Bazel monorepos.

Your monorepo has 12,000 targets across Go, Java, Python, and TypeScript.
SonarQube doesn't understand the build graph. Worktrees cost 14 GB. You have
no idea which packages are healthy and which are rotting.

Gavel runs static analyzers as Bazel aspects, evaluates quality gates defined
as code, and tracks progress run over run — per project, per package, across
the entire monorepo.

```bash
gavel init
gavel judge
```

---

## What it does

**Runs analysis inside Bazel.** Analyzers (PMD, SpotBugs, Error Prone,
golangci-lint, Ruff, Bandit, ESLint) execute as build aspects — they see the
same source tree, dependency graph, and configuration that Bazel sees.

**Normalizes everything to SARIF.** Every analyzer produces findings in a
single format. No more per-tool output parsing.

**Evaluates quality gates as code.** `gavel.yaml` defines thresholds
per severity, coverage minimums, and architecture violation limits — all
per project. Thresholds evaluate only **new** findings and violations
compared to the baseline, so adopting Gavel on existing code never blocks
on legacy debt:

```yaml
projects:
  - name: payments
    pattern: "//payments/..."
    tooling: [go]
    quality_gate:
      findings:
        max_error: 0
      coverage:
        min: 80
      architecture_violations:
        max: 0
```

**Shows what's wrong, not just that something is wrong.** Terminal output
lists every finding grouped by file, sorted by line, with severity coloring
and tool:rule references. Architecture violations shown separately.

**Tracks progress.** Each run saves a fingerprint-based snapshot. Next run
shows the delta: new findings, fixed findings, coverage trend. No server
required — progress is visible from day one.

```
  payments
    src/api/handler.go
        42  error    null check missing              PMD:NullCheck  NEW
        18  warning  exposed field                   SpotBugs:EI_EXPOSE

  42 findings (↓5 since last run)
  coverage: 73.5% (↑5.0%)
  3 new · 8 fixed · 31 existing
```

**Validates architecture.** Define layer/dependency rules in
`.gavel/architecture.yml` and an archtest aspect enforces them on every
`gavel judge` run — new violations block the gate just like findings.

## Why not SonarQube?

| | SonarQube | Gavel |
|---|---|---|
| Build graph awareness | None | Bazel aspects understand target dependencies |
| Monorepo support | Single project or branch-per-project | Hierarchical: per-package, per-project, whole repo |
| Analysis scope | Full scan or file diff | Bazel-aware: changed files + affected targets |
| Local workflow | Requires server round-trip | Fully local, zero network |
| Quality gate | Web UI configuration | Code (`gavel.yaml`), versioned with the repo |
| Progress tracking | Requires server | Local snapshots, no infrastructure |
| Setup | Java server + database + scanner | Single Go binary, runs inside Bazel |

Gavel is not a SonarQube replacement. SonarQube is a mature platform with
15 years of rule development. Gavel solves the problem SonarQube cannot:
health observability at Bazel monorepo scale, where the build graph is the
unit of analysis.

## Status

> **Alpha** — under active development. APIs and config formats may change.

Working today:

- `gavel init` — scaffolds config + Bazel integration
- `gavel judge` — runs analysis, evaluates quality gates, shows findings
- `gavel judge --project <name>` — scoped to a single project
- `gavel judge --json` — structured output for CI
- `gavel judge --absolute` — evaluate all findings (release gates, nightly)
- `gavel judge --findings-source=rules_lint` — read pre-existing SARIF from `bazel-bin/` instead of running Gavel's own aspects (auto-detected when omitted)
- `gavel judge --server URL --token TOKEN` — fetch baseline from server, submit results back
- `gavel validate` — checks Bazel integration health
- `gavel watch` — re-analyzes on file changes, emitting a JSONL event stream
- `gavel mcp` — Model Context Protocol server for editor/agent integration
- Baseline mode (default): fingerprint-based new/fixed/existing classification, committed to git for team sharing
- Analyzers: golangci-lint, PMD, CPD, SpotBugs, Error Prone, Ruff, Bandit, ESLint, Clippy
- Server: web dashboard, centralized history, team baselines, API token auth

Coming next:

- Scoped analysis — Bazel-aware diff (only affected targets)
- SARIF export for IDE integration

## License

[Apache License 2.0](LICENSE)
