# Security Policy

## Supported versions

Gavel is **alpha**. Security fixes land on `main`; there are no maintained
release branches yet.

## Reporting a vulnerability

Please **do not open a public issue** for security problems.

Use GitHub's **private vulnerability reporting**: go to the repository's
**Security** tab → **Report a vulnerability**. That opens a private advisory
thread visible only to the maintainers.

Gavel handles source code, runs analyzers through Bazel, and (in server mode)
authenticates users with Argon2id passwords, opaque sessions, and SHA-256 API
tokens. Reports about authentication/session handling, token storage, SARIF
ingestion, or build-time execution are in scope.

You can expect an initial response within a few days.
