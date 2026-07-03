---
title: GitHub Checks reporting
type: explanation
description: Why delivering a verdict to a PR is a separate `gavel report` command, and why its GitHub client lives in userinterface.
tags: [report, github, checks, ci, design-record]
---

# Delivering a verdict to the pull request

> **Status: implemented.** `gavel report` ships as a separate command. The
> how-to (the GitHub Actions snippet, flag details) lives in
> [ci.md](../ci.md) and [`clispec/v1/clispec.yaml`](../../clispec/v1/clispec.yaml);
> this record keeps only the *why*.

## Why a separate command, not a `judge` flag

SARIF and JSON output are *flags* on `gavel judge` (`--output-sarif`, `--json`),
so the tempting precedent is `gavel judge --report-github`. It is the wrong cut.

The dividing line is not "another output format" — it is **"does this touch an
external system with its own credentials and failure modes?"** SARIF and JSON are
local, pure, synchronous outputs of the same computation. Delivering a verdict to
GitHub is a *delivery* to a remote system: it needs a token, PR context, and it
fails on its own (rate limits, auth, fork restrictions). That is a different
class of concern. So:

- `gavel judge` **computes** the verdict and owns the gate and the exit code.
- `gavel report` **delivers** the verdict to an external sink, and its exit code
  reflects delivery only — the pass/fail rides in the GitHub check-run
  conclusion.

Keeping them separate means a decoration failure never masks the gate, report can
re-run without re-judging, and the gate stays owned by one command. `report`
never re-runs analysis: it reads the verdict `judge` cached under
`.gavel/results/` (see [`output/json`](../../core/userinterface/cli/judge/output/json/))
or fails telling the user to run `judge` first.

This also corrects a coupling already in the tree: server submission was bolted
into `judge` (`pipeline/server.go`) and produced dead code and stale-snapshot
bugs. `report` is the shape that mistake should have taken.

## Why the GitHub client is userinterface, not infrastructure

The client consumes `checks.CheckRun`, a userinterface type. Putting it in
`core/infrastructure/` would break the `userinterface`→ nothing-below rule the
architecture gate enforces (`infrastructure` may not import `userinterface`).

More fundamentally: a CLI's *outbound* HTTP call is a userinterface concern, not
a domain/application port implementation. The repo already models this — the
server-mode client lives in `core/userinterface/api/v1/client/`, not in
infrastructure. So the Checks client lives at
`core/userinterface/cli/report/github/`, beside the command it serves.

The command depends on a local `ChecksPublisher` port and `main.go` injects the
concrete client, so the orchestration is testable with a fake and the real HTTP
path is covered by an httptest integration test.

## Scope boundaries (deliberate, not debt)

- **Fork PRs** get a read-only `GITHUB_TOKEN`, so the check run cannot be created
  from a fork. Same-repo PRs are the zero-infra happy path; forks need a GitHub
  App. Out of the first cut.
- **GitHub Enterprise** (`GITHUB_API_URL`) and non-GitHub sinks (GitLab, …) are
  future sinks behind the `--to` flag; only `github-checks` exists today.
- **Multi-tenancy / SaaS** is explicitly not required here: with no hosted
  offering, each team self-hosts and single-tenant is correct.
