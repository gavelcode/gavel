---
title: Implementation status
type: reference
description: What is built today across core/, server, and cli — bounded contexts, use cases, adapters, and the IAM context shape. A snapshot of the build, not the design contract.
---

# Implementation status

Snapshot of what exists today. The design contract lives elsewhere —
[`model/`](model/domain-model.md) is the canonical domain/application model and
[`specs/`](specs/01-project-structure.md) the structure rules. This file records
the *current build* against that contract, so it ages and is updated as the code
moves.

## `core/` (canonical, Vernon-strict)

- **Domain**: 5 bounded contexts — `casefile`, `project`, `gavelspace`,
  `pleading`, `iam` — with sub-model packages where needed
  (`casefile/model/evidence/`, `project/model/qualitygate/`); shared
  `event/` (DomainEvent interface) and `failure/` (Sentinel + Kind
  classifier).
- **Entity identity VOs**: `ProjectID`, `CaseFileID`, `EvidenceID`,
  `GavelspaceID`, `PleadingID`, plus IAM IDs (`TenantID`, `UserID`,
  `APITokenID`) — no primitive obsession.
- **Application**: 30+ use cases grouped per BC (Simple CQRS — commands
  + queries in the same BC tree):
  `casefile/{judge,submit,classify,classifyarch,ingestfindings,
  ingestcoverage,ingestncc,collectevidence,createcasefile,
  ingestevidence,finalize,get,list,listfindings,evidencedto}`,
  `gavelspace/{create,registerproject,removeproject,
  loadgavelspace,get,list}`,
  `pleading/{file,resolve,get,list,gateview}`,
  `project/{create,updatequalitygate,updatelanguages,
  updatetargetpattern,preparebaseline,get,getbykey,list,projectview}`,
  `iam/{login,changepassword,createuser,deactivateuser,issuetoken,
  revoketoken,resolveprincipal,listmytokens,tenant/{create,suspend}}`,
  plus `shared/event/` (Shared Kernel) and
  `supporting/{search,analyzetarget}` (Supporting Subdomain cross-BC).
  Commands use `Command` + `Execute`; queries use `Query` + `Execute` +
  `finder.go` with a single-method `type Finder interface`.
  **`submit`** wraps the full judge pipeline (create → ingest →
  finalize) as a single use case. **`finalize`** computes delta against
  project baseline and updates it when verdict passes.
  **`loadgavelspace`** optionally applies arch policy and saves projects
  to the repo (via injected ports). **`preparebaseline`** fetches
  baselines from the server and seeds the CaseFile repo.
  **`analyzetarget`** runs lint aspects on a single Bazel target
  (supporting command for watch).
- **Infrastructure**: adapters organized per BC under
  `core/infrastructure/<bc>/`:
  - `casefile/{memory,postgres,sarif,lcov}/` — in-memory + Postgres
    repository, SARIF + LCOV parsers (both implement
    `application/casefile/ingest*.Parser`).
  - `gavelspace/{memory,postgres,gavelconfig}/` — in-memory + Postgres
    repos, `gavel.yaml` parser.
  - `iam/{memory,postgres,argon2,crypto}/` — in-memory repos + fakes,
    Postgres repos, Argon2id PasswordHasher, crypto/rand SecretGenerator.
  - `pleading/postgres/` — Postgres repo (no in-memory yet).
  - `project/{memory,postgres,archconfig}/` — in-memory + Postgres repos,
    architecture-policy YAML parser.
  - Cross-cutting in `platform/`: `bazel/catalog/` (language → aspects
    mapping), `bazel/runner/` (aspect + coverage + JS coverage runners,
    plus `BazelTargetAnalyzer` and `BazelTargetResolver` implementing
    `analyzetarget` ports), `bazel/installer/` (.bazelrc + MODULE
    generation), `bazel/collector/` (evidence collection adapters
    implementing `collectevidence` ports, `composite/` for
    Bazel+vitest coverage fallback), `git/` (commit SHA, branch,
    diff/changed lines), `database/` (dbkit + pgx Open/Migrate +
    testcontainer helper), `sourceblob/` (blob storage without a domain
    aggregate).
  - Cross-cutting in `supporting/`: `search/` (cross-BC Query Model
    Finder implementation).
- **Userinterface** (`core/userinterface/api/v1/`): OpenAPI-first HTTP
  layer with three top-level packages. `gen/` is the generated package
  (DTOs + the generated `StrictServerInterface` + chi wrappers, from the
  `openapi/v1/` split YAML source). `server/<bc>/` (`casefile/`,
  `gavelspace/`, `iam/`, `pleading/`, `project/`, `supporting/`, `ops/`)
  holds the handwritten handlers + mappers, and `server/httpx/` carries
  the RFC 7807 problem constructors, pagination + client-IP + body-size
  helpers, and `MustUUID`. `client/` wraps the generated OpenAPI client
  for CLI server mode. Composition into a `Server` satisfying the
  generated `StrictServerInterface` lives in
  `apps/server/internal/api/v1/`, not here.
- `bazel test //core/...` green across every package.

## IAM bounded context (`core/domain/iam/`)

The `apps/server/internal/platform/identity/` package — flat CRUD wrapping
`User`, `Session`, `APIToken` — was deleted and replaced by a Vernon-strict
bounded context in `core/`.

- **Aggregates**: `Tenant`, `User`, `Session`, `APIToken`. Each owns its
  own invariants (status transitions, expiry, scope checks) and emits
  domain events. `User` and `APIToken` carry `TenantID`; downstream
  judicial aggregates (Gavelspace/Project/CaseFile/Pleading) intentionally
  do **not** — multi-tenancy enablement is a separate refactor.
- **Value Objects**: `Email`, `Role`, `Scope`, `PasswordHash`,
  `SessionToken` + `SessionTokenHash`, `APITokenSecret` + `APITokenHash`,
  `TenantSlug`, `TenantStatus`, plus opaque IDs (`UserID`, `TenantID`,
  `APITokenID`). All construction paths validate; persisted state
  re-validates via `Reconstitute*`.
- **Repository ports**: `TenantRepository`, `UserRepository`,
  `SessionRepository`, `APITokenRepository` in
  `core/domain/iam/service/`. Domain services (`PasswordHasher`,
  `SecretGenerator`) are ports there too.
- **Application** (`core/application/iam/`): nine use cases — `login`,
  `changepassword`, `createuser`, `deactivateuser`, `issuetoken`,
  `revoketoken`, `resolveprincipal`, `tenant/create`, `tenant/suspend`,
  plus a `listmytokens` read.
- **Infrastructure**: `core/infrastructure/iam/argon2/` (PasswordHasher),
  `core/infrastructure/iam/crypto/` (SecretGenerator),
  `core/infrastructure/iam/memory/` (in-memory repos + `FakeHasher` +
  `FakeSecretGenerator` for tests), and
  `core/infrastructure/iam/postgres/` (PostgreSQL repos against the
  `iam_tenants` / `iam_users` / `iam_sessions` / `iam_api_tokens`
  tables; `iam_sessions` keys on a `SessionID` UUID with a UNIQUE
  `token_hash` for the cookie lookup, mirroring `iam_api_tokens`).
- **HTTP**: `core/userinterface/api/v1/iam/` carries the IAM endpoints
  (`auth.go`, `tokens.go`, `users.go`, plus the request-context
  middleware). DTOs are generated from `openapi/iam/iam.yaml` into the
  `oapi` package. `core/userinterface/api/v1/platform/middleware/` hosts
  the `AuthMiddleware` (backed by `resolveprincipal`), the
  `SessionCookie` helper, and the body-size middleware.

Bootstrap (`apps/server/cmd/gavel-server/main.go`) creates the `default`
tenant + first admin via use cases on a fresh database.

> Pull requests took the same path: once the model grew real behaviour
> (status transitions with rules, invariants on filing, domain events) the
> "no domain logic" carve-out no longer held, so they were promoted from a
> flat `apps/server/` CRUD struct to the **Pleading** aggregate in
> `core/domain/pleading/`. IAM was promoted for the same reason.

## `apps/server/` (composition root)

- Backend starts with `bazel run //apps/server/cmd/gavel-server -- serve`.
- Auto-migrations on startup (single bootstrap SQL, applied once).
- Bootstrap creates `default` tenant and the first admin user (random
  password to stdout) on a fresh database via `core/iam` use cases.
- The server is a composition root in three layers:
  - `cmd/gavel-server/main.go` — instantiates IAM + core repositories
    against PostgreSQL, wires every application handler, mounts the
    v1 mux under `/api/v1` and the SPA at `/`. Owns the session
    cleanup goroutine and signal-driven shutdown.
  - `internal/api/v1/` — defines `Server` as a struct that embeds eight
    per-BC handlers (via public type aliases) to satisfy the union of
    `oapi.StrictServerInterface`. No constructor: `main.go` and the
    integration fixture build the per-BC handlers themselves and
    assemble the `&apiv1.Server{...}` struct literal. `NewMux` builds
    the chi router and attaches the scope / role middleware groups.
  - `internal/platform/{config,frontend,spa}/` — env-var loading,
    embedded frontend FS loader, SPA fallback handler.
- Persistence adapters live in `core/infrastructure/<bc>/postgres/` and
  implement the domain Repository ports plus the application Finder
  ports. The dbkit wrapper, pgx connection, bootstrap SQL, and
  testcontainer helper live in `core/infrastructure/platform/database/`.
- All HTTP paths are declared in `openapi/v1/*.yaml` and served under
  `/api/v1`. Highlights:
  - **Auth**: `POST /sessions`, `DELETE /sessions/current`, `GET /me`,
    `POST /me/password`.
  - **Tokens**: `GET /me/tokens`, `POST /me/tokens`, `DELETE /me/tokens/{id}`.
  - **Read**: `GET /projects`, `GET /projects/{key}`, `GET /casefiles`,
    `GET /casefiles/{id}`, `GET /findings`, `GET /search`,
    `GET /pleadings`, `GET /pleadings/{id}`, `GET /gavelspaces`,
    `GET /gavelspaces/{name}`, `GET /projects/{key}/casefiles`,
    `GET /projects/{key}/pleadings`, `GET /projects/{key}/baseline`,
    `GET /projects/{key}/source`.
  - **Ingest** (requires `ingest` scope): the three-step submit
    `POST /casefiles` → `POST /casefiles/{id}/evidence` →
    `POST /casefiles/{id}/finalize`.
  - **Admin** (requires `admin` role): `POST /admin/users`,
    `POST /projects`, `PUT /projects/{key}/quality-gate`,
    `PUT /projects/{key}/languages`, `POST /gavelspaces`,
    `POST /gavelspaces/{name}/projects`,
    `DELETE /gavelspaces/{name}/projects/{project_id}`,
    `POST /projects/{key}/pleadings`, `PATCH /pleadings/{id}`.
- **Middleware**: `Authenticate`, `RequireRole`, `RequireScope`,
  `iam.AttachRequestMiddleware` (puts `*http.Request` on context so
  login can read remote IP + User-Agent).
- **Health**: `GET /api/v1/health` (in the spec) plus `GET /healthz`
  (root-level, for k8s probes).
- **Integration tests** for the full assembled HTTP service live in
  `apps/server/test/integration/api/v1/` (package `v1integration`), not
  next to v1 sources. They wire the real application handlers against
  in-memory fakes and exercise the chi mux through httptest.

## `apps/cli/` (composition root)

- CLI entry point: `apps/cli/cmd/gavel/main.go` — **wiring only**. Instantiates
  infrastructure adapters, application handlers, and CLI commands. Zero
  business logic.
- Commands live in `core/userinterface/cli/` (not `apps/cli/`):
  `init`, `judge`, `validate`, `watch`, `config`, `projects`, `mcp`.
- **All CLI commands are Vernon-strict clean**: zero imports from
  `core/domain/` or `core/infrastructure/` across the entire
  `userinterface/cli/` tree. Every command receives dependencies
  injected from `main.go` — no command creates infrastructure
  adapters or touches domain aggregates directly.
  - `judge/`: orchestration via `submit` → `finalize` use cases.
    `pipeline/local.go` calls `submit.Handler` and maps the result.
    `pipeline/server.go` delegates to the API client. Workspace
    detection injected as `WorkspaceResolver` function.
  - `watch/`: uses `analyzetarget.Handler` + `TargetResolver` ports.
  - `config/`, `projects/`: receive `loadgavelspace.Handler` injected.
  - `initgavel/`, `validate/`: receive `WorkspaceResolver` +
    `StructureVerifier` via local interfaces.
- Bazel integration lives in `core/infrastructure/platform/bazel/`:
  aspect catalog, runner (aspect + coverage + JS coverage + target
  analysis), installer (bazelrc/MODULE generation), collector adapters
  (implement `collectevidence` ports).
- Git integration: `core/infrastructure/platform/git/` — commit SHA,
  branch detection, diff/changed lines.
- MCP server: `core/userinterface/cli/mcp/` — `gavel mcp` subcommand.
  Stdio transport, delegates everything to the CLI binary via subprocess.
  Zero `core/` imports in MCP — pure subprocess wrapper.
- Server mode: `--server URL` + `--token TOKEN` delegates to the server's
  Open Host Service. HTTP client in `core/userinterface/api/v1/client/`
  wraps the generated OpenAPI client.
- CLI specification: `clispec/v1/clispec.yaml` — source of truth for all
  commands, flags, exit codes, output schemas.
