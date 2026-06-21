---
title: CLI architecture
type: reference
description: Current structure of the CLI — command wiring, pipelines, and Bazel and git integration.
---

# CLI Architecture

> This documents the current CLI structure. It is a working implementation,
> not necessarily the final design.

Scope: `apps/cli/` — developer tool for Bazel analysis and server submission.

## Done

Cobra commands are thin (parse flags, wire, delegate). The judge pipeline
orchestrates lint + coverage + archtest + submit. Server mode is optional
with offline fallback. Bazel integration is CLI-specific, not in `core/`.

---

## Role

The CLI (`apps/cli/`) is a developer tool that runs analysis locally via
Bazel and optionally publishes results to a server. It consumes `core/` for
domain logic and use cases.

## Structure

```
apps/cli/
├── cmd/
│   ├── gavel/                 # Composition root (wiring only)
│   └── docgen/                # CLI docs generator (cobra/doc)
├── internal/
│   ├── bazel/                 # Bazel integration
│   │   ├── catalog/           # Language → aspects/tools mapping
│   │   ├── installer/         # .bazelrc + MODULE.bazel generation
│   │   └── runner/            # Aspect execution + coverage collection
│   ├── command/               # Cobra commands
│   │   ├── initgavel/         # gavel init
│   │   ├── judge/             # gavel judge (main pipeline)
│   │   ├── validate/          # gavel validate
│   │   ├── gavelspace/        # gavel gavelspace
│   │   └── watch/             # gavel watch (JSONL event stream)
│   ├── git/                   # Git source context (commit, branch)
│   ├── server/                # HTTP client for server mode
│   └── ui/                    # Terminal output formatting
└── test/integration/
```

## Key decisions

- **Cobra commands are thin.** Each command parses flags, wires dependencies,
  and delegates to `core/` use cases or CLI-specific logic.
- **judge is the main pipeline.** It orchestrates: lint aspects → coverage →
  archtest → submit to core → baseline comparison → verdict rendering.
- **Baseline is local-first.** `.gavel/baseline/<project>/` committed to git.
  Server baseline is fetched when `--server` is configured, with fallback to
  local if unreachable.
- **Server mode is optional.** `--server` + `--token` (or env vars) enable
  fetching baseline from server and submitting results back. Without these,
  the CLI works fully offline.
- **Bazel integration is CLI-specific.** Aspect running, coverage collection,
  and target pattern resolution live in `apps/cli/internal/bazel/`, not in `core/`.
