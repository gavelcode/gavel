---
title: Server architecture
type: reference
description: Current structure of the server composition root — the v1 HTTP layer, mux, and platform glue.
---

# Server Architecture

> This documents the current server structure. It is a working implementation,
> not necessarily the final design.

Scope: `apps/server/` — HTTP API composition root.

## Done

All domain logic, application use cases, persistence adapters, and the HTTP
layer live in `core/`. The server binary is pure wiring: it builds the DI
graph and starts the chi router. PostgreSQL is the only persistence
technology, with a single bootstrap SQL applied on first run.

---

## Role

The server (`apps/server/`) is a thin composition root over `core/`. It owns:

- `cmd/gavel-server/main.go` — DI wiring, config loading, graceful shutdown.
- `internal/platform/config/` — 12-factor env var loading.
- `internal/platform/frontend/` — embedded SPA serving via Bazel runfiles.

Everything else — domain aggregates, use cases, repository implementations,
HTTP handlers, middleware, router, bootstrap seeding — lives in `core/`.

## Structure

```
apps/server/
├── cmd/gavel-server/          # Composition root (wiring only)
└── internal/
    └── platform/
        ├── config/            # Env var loading (12-factor)
        └── frontend/          # Embedded web frontend loader (Bazel runfiles)
```

The Postgres connection, dbkit wrapper, and testcontainer helper live in
`core/infrastructure/platform/database/`. Postgres adapters per BC live
under `core/infrastructure/<bc>/postgres/`. The HTTP router, bootstrap,
SPA fallback, and middleware live under
`core/userinterface/api/v1/platform/`.

## Key decisions

- **No server-specific domain, application, infrastructure, or userinterface
  layers.** Everything is in `core/`. The server is a binary that wires the
  graph and hosts the chi listener.
- **IAM is Vernon-strict, not flat CRUD.** Tenant, User, Session, and
  APIToken are aggregates in `core/domain/iam/`. Application use cases for
  login, token issuance, etc. live in `core/application/iam/`. Postgres
  adapters in `core/infrastructure/iam/postgres/`.
- **Handlers use the Published Language.** HTTP DTOs and the RFC 7807
  problem-details mapper live in `core/userinterface/api/v1/shared/`
  (cross-cutting) and the per-BC feature folders for resource DTOs.
- **PostgreSQL only.** `pgx/v5` via `database/sql`. Connection pool tuning
  documented in [postgres-pool-tuning.md](../design/postgres-pool-tuning.md).
- **Bootstrap migration.** No incremental goose. `bootstrap.sql` in
  `core/infrastructure/platform/database/` creates the full schema on a
  fresh database, idempotent on a non-fresh one.
- **First-run tenant + admin seeding.** `seed.sql` (applied by
  `database.Migrate()` alongside `bootstrap.sql`) creates the `default`
  tenant and admin `admin@gavel.local` (password `changeme`,
  `must_change_password=true`). There is no Go bootstrap function.
