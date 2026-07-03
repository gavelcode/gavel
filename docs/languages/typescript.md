---
title: TypeScript
type: reference
description: TypeScript and JavaScript analysis via ESLint — syntactic or type-aware — with a composite vitest + LCOV coverage strategy.
resource: https://github.com/gavelcode/gavel-tools/tree/main/lint/lang/typescript
tags: [typescript, eslint, analyzers]
---

# TypeScript

Gavel analyzes TypeScript and JavaScript with **ESLint**, run as a hermetic Bazel
aspect from [gavel-tools](https://github.com/gavelcode/gavel-tools), plus a
composite coverage strategy for vitest projects.

| | |
|---|---|
| **Lint tool** | ESLint |
| **Aspect** | `typescript_eslint_submission_aspect` |
| **Archtest** | `typescript_archtest_submission_aspect` |
| **Bazel targets** | anything carrying `JsInfo` (`js_library`, `ts_project`) |
| **SARIF suffix** | `.eslint.sarif` |
| **Coverage** | vitest JSON → LCOV, with a Bazel LCOV fallback (composite) |

## How ESLint runs

The aspect runs gavel's pinned ESLint over a target's TypeScript sources inside
the Bazel sandbox, resolving the project's own ESLint plugins from `JsInfo`. The
mechanics — plugin-closure harvesting, the pnpm-store repair that keeps it
hermetic, and the contract for bumping `rules_js` / ESLint — live in gavel-tools,
so this page points there rather than restating them:

- [web_project macro](https://github.com/gavelcode/gavel-tools/blob/main/docs/web-project.md) — how a frontend exposes its sources, config and plugins to the aspect.
- [The hermetic analyzer driver](https://github.com/gavelcode/gavel-tools/blob/main/docs/tier-model.md) — why the run is sandboxed, and the ESLint maintenance contract.

## Configuring ESLint

ESLint reads a standard flat config (`eslint.config.js`) from the project. Two
modes:

- **Syntactic** — recommended rules that need no type information; the default.
- **Type-aware** — the `@typescript-eslint` rules that need types
  (`no-floating-promises`, `no-unsafe-*`, `no-misused-promises`, …). Extend
  `tseslint.configs.recommendedTypeChecked` and set `parserOptions.projectService:
  true` with `tsconfigRootDir: import.meta.dirname`. The aspect already carries the
  tsconfig and type deps into the sandbox, so no extra wiring is needed.

> [!IMPORTANT]
> Under type-aware config, give test files their own **syntactic-only** block.
> `projectService` builds its program from the app tsconfig; any file that tsconfig
> excludes (typically `*.test.*`) errors as "not found by the project service".
> Match `**/*.test.{ts,tsx}` to a block that does not set `projectService`.

## Coverage: composite strategy

TypeScript coverage uses a two-tier approach in
`core/infrastructure/platform/bazel/collector/composite/coverage.go`:

1. **Primary** — vitest coverage JSON, converted to LCOV by
   `core/infrastructure/platform/bazel/runner/jscoverage.go`, which runs
   `npx vitest run --coverage` in the project.
2. **Fallback** — Bazel's `--combined_report=lcov`.

vitest produces more accurate JS coverage than Bazel's LCOV instrumentation, and
`bazel coverage` with `rules_js` is not always reliable — so the collector prefers
vitest output and detects it automatically. No user configuration needed.

> [!NOTE]
> The primary path shells out to `npx vitest`, so wherever `gavel judge` runs — CI
> included — it needs Node on `PATH` and the project's dev dependencies installed
> (e.g. `pnpm install`). Without them coverage silently falls to 0%.
