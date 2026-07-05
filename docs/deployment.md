---
title: Server deployment
type: how-to
description: Building, configuring, and running gavel-server — database bootstrap and environment variables.
tags: [deployment, server, operations]
---

# Server Deployment

`gavel-server` is a single static Go binary that serves the API, web
dashboard, and manages centralized baselines for teams.

## Build

```bash
bazel build //apps/server/cmd/gavel-server
# Binary at bazel-bin/apps/server/cmd/gavel-server/gavel-server_/gavel-server
```

## Database

PostgreSQL is the only supported database. Versioned goose migrations apply
automatically on startup — no manual migration step required.

```bash
createdb gavel
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GAVEL_ADDR` | `:8080` | HTTP listen address |
| `GAVEL_DATABASE_URL` | `postgres://localhost:5432/gavel?sslmode=disable` | PostgreSQL connection string |
| `GAVEL_SESSION_TTL_HOURS` | `168` (7 days) | Session expiration |
| `GAVEL_SECURE_COOKIES` | `false` | Set to `true` in production (requires HTTPS) |
| `GAVEL_DATA_DIR` | `./data` | Data directory for file storage |

## Start

```bash
gavel-server serve
```

Or apply migrations only (without starting the server):

```bash
gavel-server migrate
```

## First boot

On a fresh database, `database.Migrate()` applies the versioned migrations
under `migrations/` (starting with `00001_bootstrap.sql`) and then a seed
(`seed.sql`) that creates:

1. The `default` tenant
2. An admin user `admin@gavel.local` with the password `changeme` and
   `must_change_password=true`

**Change the password on first login.** The seeded `changeme` is public; the
security control is the forced password change before the server is exposed.

## Health checks

| Endpoint | Purpose |
|----------|---------|
| `GET /healthz` | Kubernetes liveness probe (root level) |
| `GET /api/v1/health` | Application health (in the API spec) |

## Authentication

- **Web sessions**: Argon2id password hashing, opaque session tokens in
  HttpOnly + SameSite=Lax cookies
- **API tokens**: `gav_` prefixed, SHA-256 hashed in DB, shown once at
  creation. Use for CLI `--token` and CI integration

## Graceful shutdown

The server handles SIGTERM and SIGINT:

1. Stops accepting new connections
2. Drains in-flight requests
3. Closes resources in reverse order (server → database → logger/tracer)

## Frontend

The web dashboard is embedded in the server binary via Bazel runfiles.
It serves as a SPA with `index.html` fallback at the root path `/`.
The API is mounted under `/api/v1`.

## Production checklist

- [ ] Set `GAVEL_SECURE_COOKIES=true` (requires HTTPS)
- [ ] Use a strong `GAVEL_DATABASE_URL` with TLS (`sslmode=require`)
- [ ] Save the initial admin password from first boot
- [ ] Create API tokens for CI pipelines (`POST /api/v1/me/tokens`)
- [ ] Configure health check probes on `/healthz`
