# Contributing to Gavel

## Prerequisites

- Go 1.25
- Bazel (bzlmod — `MODULE.bazel`, not WORKSPACE)
- Podman (for integration tests using testcontainers)

## Build and test

```bash
bazel build //...                  # Build everything
bazel test //...                   # Run all tests (must be green)
bazel run //:gazelle               # Regenerate BUILD.bazel files
```

For OpenAPI changes:

```bash
make openapi-gen                   # Bundle + regenerate Go + TS clients
make openapi-check                 # Fail if generated artifacts are stale
```

## Project structure

```
core/           # Shared library — Vernon-strict DDD
  domain/       #   Business logic, aggregates, value objects
  application/  #   Use cases (commands + queries, Simple CQRS)
  infrastructure/ # Adapters (Postgres, SARIF parser, Bazel runner)
  userinterface/  # HTTP handlers, CLI commands
apps/
  cli/          # CLI composition root (wiring only)
  server/       # Server composition root
  web/          # React + Vite frontend
```

Dependency direction is inward only: userinterface → application → domain.
Infrastructure implements domain and application interfaces. See
`docs/specs/01-project-structure.md` for full rules.

## Code conventions

- **TDD**: write the failing test first, then implement
- **Zero mocks** in domain tests — use hand-rolled in-memory fakes
- **Black-box tests**: `package xxx_test` by default
- **No CGO** — keeps the binary portable
- **No new dependencies** without discussion
- `gofmt` and `go vet` clean
- `bazel run //:gazelle` after adding/removing files

The full coding specifications are in `docs/specs/` (20 documents covering
Go style, testing, persistence, security, observability, and more).

## Architecture rules

- Domain imports nothing outside `core/domain/`
- Application imports domain only
- Userinterface imports application only, never domain or infrastructure
- Infrastructure implements domain/application interfaces

These are enforced by `gavel judge` via archtest aspects.

## Testing

- **Unit tests**: next to the code, `*_test.go`. Table-driven preferred
- **Integration tests**: `test/integration/` when spanning packages or
  requiring external systems (Postgres via testcontainers)
- `bazel test //...` must be green after every change — no exceptions

## Submitting changes

1. Verify layer integrity: no cross-layer imports
2. Run `bazel test //...` — all green
3. Run `bazel run //:gazelle` if you added/removed files
4. Keep changes focused — one responsibility per PR
