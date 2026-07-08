---
title: Gavel documentation
type: index
description: Index of the Gavel knowledge base, grouped by kind.
tags: [index]
---

# Gavel documentation

The knowledge base for what Gavel *is* and how it works. Each concept is a
single markdown file with an OKF `type` / `title` / `description` header;
this index maps them. Start here and follow the links.

## Start here

* [Quickstart](quickstart.md) — From zero to a first quality report in five minutes: install, init, judge.

## Reference — look things up

* [Configuration reference](configuration.md) — The `gavel.yaml` and `architecture.yml` schema: projects, quality-gate rules, DDD layer definitions.
* [Domain model](model/domain-model.md) — The canonical aggregates, value objects, invariants, identity types, and events (Vernon IDDD strict).
* [Application model](model/application-model.md) — The Simple CQRS commands and queries that orchestrate the aggregates.
* [Project structure](specs/project-structure.md) — Product boundaries, the four-layer architecture, package rules, and the canonical directory tree.
* [Server architecture](specs/server-architecture.md) · [CLI architecture](specs/cli-architecture.md) · [Web architecture](specs/web-architecture.md) — Current structure of each composition root.
* [Language support](languages/index.md) — Languages, tools per language, how lint aspects flow to SARIF, coverage. Per-language: [Go](languages/go.md) · [Java](languages/java.md) · [Python](languages/python.md) · [TypeScript](languages/typescript.md) · [Rust](languages/rust.md).
* [Implementation status](status.md) — A snapshot of what is built today across `core/`, server, and CLI.

## Guides — get a task done

* [CI integration](ci.md) — Running `gavel judge` and `gavel report` in CI pipelines.
* [Server deployment](deployment.md) — Building, configuring, and running `gavel-server`.
* [IDE integration](ide/vscode.md) — View findings inline: [VS Code](ide/vscode.md) · [IntelliJ IDEA](ide/intellij.md) · [MCP for agents](ide/mcp.md).

## Explanation — understand the why

* [Baseline](baseline.md) — How Gavel separates new findings from existing debt; the pass/ratchet update rules.
* [Spec design](specs/spec-design.md) — How to write specification documents LLMs can follow reliably.

### Decision records (`design/`)

* [Incrementality decision record](design/incrementality-decision.md) — Why Gavel relies on Bazel's action cache instead of rdeps/diff scoping.
* [Gavel and rules_lint integration](design/rules-lint-integration.md) — Where `aspect_rules_lint` fits relative to the native SARIF aspects.
* [PostgreSQL connection pool tuning](design/postgres-pool-tuning.md) — Sizing the server's pgx pool.
* [Coverage exclusion policy](design/coverage-exclusion-policy.md) — Why Gavel mirrors what the tools measure (no per-line pragma) and where legitimate exclusions belong.
* [GitHub Checks reporting](design/github-checks-reporting.md) — Why `gavel report` is a separate command from `judge`, and why its GitHub client lives in userinterface.
* [Multi-tenancy decision record](design/multi-tenancy.md) — Why tenants are isolated with a shared schema + `tenant_id` carried in each aggregate's identity, not schema-per-tenant, and where RLS fits as a future layer.

## Related

* Lint aspects and the `web_project` build macro live in the external
  **`gavel_tools`** Bazel module
  ([github.com/gavelcode/gavel-tools](https://github.com/gavelcode/gavel-tools)),
  which carries its own OKF docs bundle.
