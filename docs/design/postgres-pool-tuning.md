---
title: PostgreSQL connection pool tuning
type: explanation
description: Decision record for sizing the server's pgx connection pool.
---

# PostgreSQL Connection Pool Tuning

## Decision

Configure the `database/sql` connection pool with these values for the
single-instance Gavel server:

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| `MaxOpenConns` | 25 | Ceiling to prevent database overload. Formula: `max_connections / app_instances - admin_reserve`. |
| `MaxIdleConns` | 10 | ~40% of max open. Keeps warm connections ready without holding resources for nothing. |
| `ConnMaxLifetime` | 30 min | Recycles connections periodically to handle DNS changes, PostgreSQL restarts, and prevent memory leaks in long-lived connections. Lower to ~10 min when a load balancer sits in front of PostgreSQL. |
| `ConnMaxIdleTime` | 5 min | Frees connections that have been idle too long. Prevents stale idle connections from accumulating in the pool. |

Use `PingContext` with a 5-second timeout at startup instead of bare `Ping()`
to fail fast if PostgreSQL is unreachable.

## Context

The initial implementation used `ConnMaxLifetime` of 5 minutes (too aggressive,
causing unnecessary TCP handshake + authentication overhead), `MaxIdleConns` of
5 (too low relative to 25 max open — connections were being closed and reopened
under moderate load), no `ConnMaxIdleTime` (idle connections stayed forever),
and bare `Ping()` without a timeout (could block indefinitely at startup).

## Alternatives considered

**pgxpool (native pgx pool):** Faster than `database/sql` due to binary
protocol and no reflection overhead. However, Gavel uses `database/sql` to
keep repositories driver-agnostic (`dbkit.DB` auto-rebinds `?` → `$N`). The
native pool would tie repos to PostgreSQL, giving up that abstraction.
Acceptable trade-off for now.

**PgBouncer (external pooler):** Offloads pool management to a sidecar. Adds
operational complexity not justified for a single-instance deployment. Revisit
if Gavel scales to multiple instances.

## When to revisit

- Multiple application instances sharing the same PostgreSQL.
- Load balancer in front of PostgreSQL (lower `ConnMaxLifetime`).
- Burst traffic patterns (raise `MaxIdleConns` toward `MaxOpenConns`).
- Monitor `db.Stats()`: if `WaitCount` rises while `InUse` is near
  `MaxOpenConns`, the pool is undersized.

## Sources

- [Go PostgreSQL Connection Pooling (OneUptime, 2026)](https://oneuptime.com/blog/post/2026-01-07-go-postgresql-connection-pooling/view)
- [Go + PostgreSQL: Best Practices (DEV Community)](https://dev.to/mx_tech/go-with-postgresql-best-practices-for-performance-and-safety-47d7)
- [pgx Driver Optimization (Gold Lapel)](https://goldlapel.com/grounds/go-postgres/go-postgresql-optimization)
- [pgxpool package documentation](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool)
