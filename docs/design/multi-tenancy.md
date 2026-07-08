---
title: Multi-tenancy decision record
type: explanation
description: Why Gavel isolates tenants with a shared schema plus a tenant_id carried in each aggregate's identity, rather than schema-per-tenant or database-per-tenant.
tags: [multi-tenancy, tenant, isolation, postgres, design-record]
---

# Multi-tenancy Decision Record

## Status

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | `tenant_id` added to the judicial tables (nullable) + backfill to the default tenant | DONE (migration `00002`) |
| Phase 2 | `TenantID` in each judicial aggregate's identity; repos scope writes/reads by tenant | DONE (Gavelspace, Project, CaseFile, Pleading) |
| Phase 3 | `tenant_id NOT NULL` on every judicial table | DONE (migrations `00003`–`00005`) |
| Phase 4 | Row-Level Security as defense-in-depth | NOT STARTED (see [Future](#future-row-level-security)) |

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Isolation model | **Shared schema + `tenant_id` discriminator column** | Industry-recommended default for SaaS: one schema to migrate, no connection-pool blow-up, scales to many tenants |
| Where the tenant lives in the domain | **Aggregate identity** (`NewXxx(tenantID, …)`, `ReconstituteXxx(…, tenantID, …)`, `TenantID()` getter) | Vernon/IDDD Option A. The `Save` writes `tenant_id` straight from the aggregate, so persisting under the wrong tenant is structurally impossible — not just a `WHERE` clause you might forget |
| Read isolation | Every repository/finder takes the tenant and filters on `tenant_id` | Closes the "read leak" side; the aggregate identity closes the "write leak" side |
| Tenant source at the edge | Server: `principal.TenantID` from the authenticated request. CLI: a fixed local sentinel (`tenant.LocalTenantID`), single-tenant | Keeps `userinterface` from importing the domain; the CLI stays a zero-config local tool |
| `NOT NULL` timing | Nullable first (backfill), then tightened per aggregate slice | Lets each aggregate wire the tenant end-to-end before the column becomes mandatory |
| Schema-per-tenant | **Rejected** | See [below](#alternatives-considered) — migration multiplication, catalog bloat, and *no* real DB-level enforcement anyway |
| Database-per-tenant | **Rejected** (kept as a possible future *dedicated* tier) | Connection-pool limits (PgBouncer pools are per-database) and no cross-tenant queries |

The tenant is a value object in the IAM bounded context (`core/domain/iam/model/tenant`); the judicial aggregates reference it by typed ID, never by pointer. The column/constraint story is the migration set (`core/infrastructure/platform/database/migrations/`); the write-path guarantee is the aggregate constructors. This record is the *why*; those are the *what*.

## Why the tenant is in the aggregate identity, not just a column

A bare `tenant_id` column makes isolation a property of **every query** — one forgotten `WHERE tenant_id = ?` is a cross-tenant leak. Putting the tenant in the aggregate's identity moves the guarantee up a level:

- The repository's `Save` reads `tenant_id` from the aggregate, so a write can only ever land under the aggregate's own tenant. There is no code path that writes a row under a caller-supplied tenant that differs from the loaded aggregate.
- Reads still carry the tenant explicitly (repos/finders take it and filter), but the *write* side — the one that corrupts data rather than merely over-reading it — is closed by construction.

This is why the domain change was Option A (tenant in identity) rather than Option B (tenant scoped only in the repository).

## Alternatives considered

Researched against the current Postgres multi-tenancy literature (see [Sources](#sources)). The three canonical models:

| | Shared schema + `tenant_id` (chosen) | Schema-per-tenant | Database-per-tenant |
|---|---|---|---|
| Isolation | Application-level (+ optional RLS) | Logical, but **no DB-level enforcement** — `search_path` can be bypassed in code | Physical, strongest |
| Migrations | **Once**, to one schema | **N times**, one per tenant — on every deploy *and* every tenant signup | N times + `CREATE DATABASE` |
| Connection pooling | Fine | Fine | **Breaks**: PgBouncer pools are per-database; you exhaust `max_connections` |
| Scale ceiling | High | "Won't scale past a few hundred" — catalog bloat slows the planner | Same ceiling, +~8 MB/db |
| Cross-tenant / operator queries | Trivial | Cross-schema joins | Impossible inside Postgres |

**Why schema-per-tenant is worse *for Gavel* specifically:**

1. **Migrations.** Gavel applies one embedded goose migration set once. Schema-per-tenant would run every migration against every schema on every deploy, and turn tenant provisioning from "insert a row in `iam_tenants`" into "create a schema + run all migrations." That ripples into the migration runner, the provisioning flow, and the integration testkit.
2. **No enforcement gain.** Schema-per-tenant has no database-level guarantee either (miss the `search_path` and you read the wrong schema), so it *still* relies on application discipline — which Gavel already has, and more robustly, via aggregate identity.
3. **Scale + operator flows.** A SaaS aiming at many tenants would hit catalog bloat with thousands of schemas, and Gavel's operator/cross-tenant flows would become cross-schema joins.

Database-per-tenant only earns its keep as a **dedicated tier** for enterprise tenants with a hard physical-isolation requirement (a hybrid: most tenants shared, a few premium on dedicated infrastructure). It is not the default.

## Future: Row-Level Security

The one real risk of the shared-schema model is a query that forgets its tenant filter and over-reads. The proportionate mitigation is **not** schema-per-tenant — it is Postgres **Row-Level Security (RLS)** layered on the existing column:

- Keep `tenant_id` (zero model change) and add RLS policies so the engine itself refuses rows for the wrong tenant even when a query omits the filter. Additive defense-in-depth over what already exists.
- Cost: a per-connection/transaction session variable (e.g. `SET app.current_tenant = …`), a small per-query overhead, and careful testing. Opinions differ on pushing security into the database (PlanetScale is cautious; others treat shared-schema + RLS as the 2026 default), which is why it is deferred rather than adopted blindly.

RLS is the natural next capa if isolation guarantees ever need to be raised beyond application discipline. Schema/database-per-tenant stays reserved for a dedicated tier.

## Provisioning & seeding

- **A tenant and its first admin are provisioned as one atomic unit.** A tenant with no admin is unusable, so the domain's `TenantProvisioner` port (mirroring Vernon's `TenantProvisioningService`) commits both aggregates in a single transaction — the Postgres impl runs both repository writes against the same `*Tx`; the in-memory fake rolls the tenant back if the admin write fails. A partial provision (a tenant with no admin) cannot happen. The provisioned admin is forced to change the password on first login.
- **The default tenant/admin identity is defined once** (`core/infrastructure/iam/bootstrap`). First-boot seeding and the integration testkit both seed *this exact identity*, and login-flow tests authenticate against it; a divergence would let tests pass against an admin production never creates.
- **First-boot seeding runs only in `serve`, never in `migrate`** — a migration job must never log a credential. It holds a Postgres advisory lock so replicas booting a fresh database serialize: the winner seeds, the rest see the admin and no-op without paying the Argon2 cost. The admin is (re)created whenever missing, so a deleted admin never locks an operator out permanently; a generated password is logged once, after the write. The password-resolution branch lives in a small testable unit (`apps/server/internal/platform/firstadmin`) kept out of the composition root, so it is covered directly instead of dragging the untestable DI wiring into the coverage denominator.
- **Cross-tenant lifecycle commands are operator-only.** Provisioning/suspending a tenant crosses the tenant boundary, so it belongs to whoever operates the host (the same privilege as `serve`/`migrate`) — an operator CLI subcommand, not an in-tenant admin or an HTTP endpoint. Those commands migrate the database first (like `serve`), so running against a never-migrated database applies the schema instead of failing with a raw "relation does not exist".
- **`tenant.LocalTenantID`** is a fixed, well-known non-nil sentinel under which the CLI runs single-tenant local judgements with in-memory repositories, so aggregates minted locally are never persisted tenant-less.

## Operational notes: migrations & test isolation

- **Migrations hold a Postgres session-level advisory lock**, so concurrent server replicas booting against a fresh database serialize instead of racing on `CREATE TABLE`.
- **The integration testkit** keys its reused Postgres container to a hash of the embedded migrations (a schema change yields a fresh container automatically — see the `test(db)` commit), serializes per-test truncate+seed with an advisory lock so parallel packages don't race the shared schema, caches the Argon2 hash of the seed password once (so seeding each test doesn't pay the deliberately slow hash), and **fails loudly** on a broken migration / bad DSN / seed error — skipping *only* when no container runtime is reachable.

## Sources

- [Approaches to tenancy in Postgres — PlanetScale](https://planetscale.com/blog/approaches-to-tenancy-in-postgres)
- [Building SaaS with PostgreSQL — Multi-Tenancy Patterns Compared — Aditya Agrawal](https://www.adiagr.com/blog/07-saas-postgres-multitenancy-patterns/)
- [Multi-Tenancy Database Patterns: Schema vs Database vs Row-Level — dasroot.net](https://dasroot.net/posts/2026/01/multi-tenancy-database-patterns-schema-database-row-level/)
