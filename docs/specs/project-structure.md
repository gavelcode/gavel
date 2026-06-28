---
title: Project structure
type: spec
description: Product boundaries, the four-layer Vernon architecture, package organization rules, and the canonical directory tree.
---

# Project Structure

Product boundaries, four-layer architecture, package organization rules, and
the canonical directory tree (at the end of this spec).

Scope: all Go code in `core/`, `apps/cli/`, `apps/server/`.
Web follows Feature-Sliced Design (see spec 16).

## Done

Every package lives in the correct layer. Dependencies point inward only.
No layer imports a package from an outer layer. Package names are short,
lowercase, singular. Bounded contexts are organized by domain concept.
Composition roots contain only wiring.

---

## Product boundaries

- `core/` — canonical Vernon-strict DDD shared library. Domain aggregates,
  application use cases, infrastructure adapters per BC, the userinterface
  feature folders, and the cross-cutting platform/shared/supporting wrappers.
  Consumed by both CLI and server.
- `apps/cli/` — developer CLI. Must not import `apps/server/internal`.
- `apps/server/` — backend composition root. Wires the DI graph and hosts
  the chi router. No domain, application, or HTTP code lives here.
- `apps/web/` — browser UI (Feature-Sliced Design, not DDD layers).

Shared contracts live in `core/` — not implicit access to another app's
internals.

## Four-layer architecture

```
userinterface (HTTP/CLI/UI) ─► application (use cases) ─► domain (business logic)
                                                          ▲
infrastructure (adapters)  ────────────────────────────────┘
```

- **domain/** — Pure business logic. Zero external dependencies. The most
  stable code in the project.
- **application/** — Orchestrates use cases. Translates between DTOs and
  domain types. Owns transactional boundaries.
- **infrastructure/** — Implements domain interfaces. Knows about domain
  types, never imports application or userinterface.
- **userinterface/** — Entry points (HTTP handlers, CLI). Delegates to
  application. Never imports domain or infrastructure.

Why separate userinterface from infrastructure: userinterface translates
**incoming** traffic; infrastructure fulfills **outgoing** needs. Collapsing
them leaks HTTP concerns into persistence code.

## Layer dependency rules

### Rule 1: Domain imports nothing external

`domain/` must not import `database/sql`, `net/http`, any ORM or framework,
or any package from `infrastructure/`, `userinterface/`, or `application/`.

`domain/` may import stdlib packages for basic operations (`fmt`, `errors`,
`time`, `strings`, `crypto/rand`) and other `domain/` packages.

### Rule 2: Repository interfaces in domain, Finder ports in application

```go
// domain/casefile/service/repository.go — write-side port (aggregate persistence)
type CaseFileRepository interface {
    Save(ctx context.Context, c casefile.CaseFile) error
    FindByID(ctx context.Context, id casefile.CaseFileID) (casefile.CaseFile, error)
}

// infrastructure/casefile/postgres/case_file_repository.go — Postgres implementation
type Repository struct { db *database.DB }

// application/casefile/get/finder.go — read-side port (Simple CQRS query)
type Finder interface {
    GetByID(ctx context.Context, id string) (*CaseFileDetail, error)
}

// infrastructure/casefile/postgres/case_file_finder.go — Postgres implementation
type CaseFileFinder struct { db *database.DB }
```

The write side (Repository) loads aggregates. The read side (Finder) returns
read projections optimized for the UI without reconstructing aggregates.
Both live in the same BC infrastructure package.

### Rule 3: Application depends on domain only

Application receives infrastructure as interfaces through constructors.
Never instantiates infrastructure directly. Owns transactional boundaries.

### Rule 4: Userinterface depends on application only

Userinterface calls application use cases. It never imports domain packages —
not aggregates, not value objects, not domain events. If userinterface needs
domain-originated data, application translates it to DTOs first.

### Rule 5: Infrastructure implements domain interfaces

Infrastructure knows about domain types but never imports application or
userinterface. Adapters are organized per BC under
`core/infrastructure/<bc>/{memory,postgres,...}/`.

## Package naming rules

- Short, lowercase, single word. No underscores, no camelCase, no plurals.
- Name after what the package does, not what it contains.
- Generic `utils`, `helpers`, `common`, `base`, `misc` are forbidden.
- `pkg/` only when code is genuinely reusable by external projects.

### The shared/, supporting/, platform/ exception

Inside each of `core/application/`, `core/infrastructure/`, and
`core/userinterface/api/v1/` we use three named wrappers with explicit
Vernon vocabulary, **not** generic dumping grounds:

- `shared/` — **Shared Kernel** (Vernon IDDD ch. 14): cross-BC contracts that
  every BC depends on. E.g. `application/shared/event/` (Event DTO). Domain has
  its own `core/domain/shared/` with `event/` + `failure/`.
- `supporting/` — **Supporting Subdomain** (Vernon/Evans): cross-BC use cases or
  endpoints with no owning aggregate. E.g. `application/supporting/search/`,
  `infrastructure/supporting/search/`, `userinterface/api/v1/supporting/{search,source}/`.
- `platform/` — platform infrastructure with no domain semantics. E.g.
  `infrastructure/platform/{bazel/catalog,database,sourceblob}/` and
  `userinterface/api/v1/platform/{middleware,bootstrap,router,spa}/`.

These three names are reserved for those meanings. Do not use them as a
miscellany. If a piece does not fit a BC, shared, supporting, or platform,
stop and revisit the design.

## Bounded context organization

Organize by domain concept, not by technical type:

```go
// Right
core/domain/casefile/
core/domain/project/
core/application/casefile/
core/infrastructure/casefile/{memory,postgres,sarif,lcov}/
core/userinterface/api/v1/casefile/

// Wrong
core/domain/entities/
core/domain/valueobjects/
core/domain/repositories/
core/userinterface/api/v1/{dto,handlers}/   # flat by tech type
```

BCs are as independent as possible. Direct imports between BCs are allowed but
minimized. For low coupling, prefer domain events. Never share a SQL table
between BCs.

### Feature Folders in userinterface

The userinterface layer uses **Feature Folders** (Bogard, Microsoft eShopOn
Containers) — each BC has its own folder keeping its HTTP handlers and JSON
DTOs together, not split by technical type:

```
userinterface/api/v1/casefile/
├── analysis_handler.go
├── baseline_handler.go
├── case_file_handler.go
├── finding_handler.go
├── evidence.go              # DTO + mapper
├── finding.go               # DTO + mapper
├── verdict.go               # DTO + mapper
└── ...
```

## Composition roots

`apps/cli/cmd/gavel/main.go` and `apps/server/cmd/gavel-server/main.go` are the
composition roots. Wiring only. No business logic. No conditionals beyond
feature selection by config.

## Build system: Bazel

- `MODULE.bazel` at root (bzlmod, not WORKSPACE).
- `go.mod` at root — Gazelle reads it for dependency resolution.
- `BUILD.bazel` files generated by Gazelle: `bazel run //:gazelle`.
- All build, run, test via `bazel build`, `bazel run`, `bazel test`.

## Canonical directory tree

```
gavel/
├── MODULE.bazel · BUILD.bazel · go.mod       # bzlmod module · gazelle target · root Go module
├── .gavel/                                    # gavel home — config + generated artifacts (committed)
│   ├── gavel.yaml · architecture.yml          #   project config · DDD layer rules
│   ├── gavel.bazelrc · gavel.MODULE.bazel      #   generated aspect/tool registrations
│   └── baseline/<project>/{findings,architecture,coverage}
├── openapi/v1/                                # Published Language — split YAML API contract
│   ├── openapi.yaml · shared/common.yaml      #   consumed by server (Go gen) + web (TS gen)
│   └── <bc>/<bc>.yaml                          #   iam · gavelspace · project · pleading · casefile · supporting · platform
├── core/                                      # CANONICAL Vernon DDD — shared library
│   ├── domain/<bc>/{model,service}/           #   aggregates + VOs (model) · repository ports (service); shared/{event,failure}
│   ├── application/<bc>/<usecase>/            #   Simple CQRS: Command/Query + Handler + finder.go; shared/ + supporting/
│   ├── infrastructure/<bc>/{memory,postgres,…}/  # adapters; platform/{bazel,git,database,sourceblob}; supporting/
│   └── userinterface/
│       ├── api/v1/                            #   OpenAPI-first HTTP: gen/ (generated) + per-BC handlers + shared/
│       └── cli/                               #   Cobra: judge · init · validate · watch · config · projects · mcp
├── clispec/v1/clispec.yaml                    # CLI spec — commands · flags · exit codes
├── apps/                                      # composition roots (wiring only)
│   ├── cli/cmd/gavel/                         #   CLI entry — wires core deps + commands
│   ├── server/                               #   internal/api/v1 (Server + mux) + internal/platform
│   └── web/                                   #   React + Vite + TypeScript (Feature-Sliced Design)
├── tools/                                     # gavel-specific Bazel tooling kept in-repo
│   ├── clispec-gen/                           #   CLI-spec codegen
│   └── spectest/                              #   OpenAPI $ref resolvability test
└── docs/                                      # knowledge base (see docs/index.md)
```

The lint aspects (per-language analyzers → SARIF) and the `web_project` build
macro live in the external **`gavel_tools`** Bazel module (consumed via
`bazel_dep`), not in this repo — the interface is SARIF files on disk.
