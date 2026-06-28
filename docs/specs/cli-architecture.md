---
title: CLI architecture
type: reference
description: Current structure of the CLI — command wiring, pipelines, and Bazel and git integration.
---

# CLI Architecture

> This documents the current CLI structure. It is a working implementation,
> not necessarily the final design.

Scope: the `gavel` CLI — its composition root (`apps/cli/`) plus the command,
Bazel, and git code it wires from `core/`.

## Done

Cobra commands are thin (parse flags, wire, delegate) and live in `core/`, not
`apps/cli/`. `apps/cli/cmd/gavel/main.go` is wiring only. The judge pipeline
orchestrates lint + coverage + archtest + submit. Server mode is optional with
offline fallback.

---

## Role

The CLI is a developer tool that runs analysis locally via Bazel and optionally
publishes results to a server. `apps/cli/` is a thin composition root; the
command logic, Bazel integration, and git adapters all live in `core/` so the
server can reuse them.

## Structure

```
apps/cli/
└── cmd/gavel/main.go          # Composition root — wiring only, zero business logic

core/userinterface/cli/        # Cobra commands (no core/domain or core/infrastructure imports)
├── initgavel/                 # gavel init
├── judge/                     # gavel judge (main pipeline: pipeline/local.go + pipeline/server.go)
├── validate/                  # gavel validate
├── watch/                     # gavel watch (JSONL event stream)
├── config/  · projects/       # read-only views over loadgavelspace
├── mcp/                       # gavel mcp — MCP server (subprocess wrapper over the CLI)
└── ui/                        # terminal output formatting

core/infrastructure/platform/  # CLI-facing adapters, shared with the server
├── bazel/{catalog,runner,installer,collector}/   # aspect mapping · aspect+coverage runners · bazelrc/MODULE gen · evidence collectors
└── git/                       # commit SHA, branch, diff/changed lines

core/userinterface/api/v1/client/   # HTTP client for server mode
```

## Key decisions

- **Commands are thin and live in `core/`.** Each command parses flags, takes
  dependencies injected from `main.go`, and delegates to `core/` use cases.
  Zero imports from `core/domain/` or `core/infrastructure/` across the whole
  `userinterface/cli/` tree.
- **judge is the main pipeline.** `pipeline/local.go` runs the full core
  pipeline (submit → finalize) locally; `pipeline/server.go` delegates to the
  API client. It orchestrates lint aspects → coverage → archtest → submit →
  baseline comparison → verdict rendering.
- **Baseline is local-first.** `.gavel/baseline/<project>/` committed to git.
  Server baseline is fetched when `--server` is configured, with fallback to
  local if unreachable.
- **Server mode is optional.** `--server` + `--token` (or env vars) enable
  fetching baseline from server and submitting results back. Without these,
  the CLI works fully offline.
- **Bazel integration lives in `core/infrastructure/platform/bazel/`**, not in
  `apps/cli/`, so both the CLI and (where relevant) the server consume the same
  aspect catalog, runner, and installer.
