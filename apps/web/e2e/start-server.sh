#!/usr/bin/env bash
set -euo pipefail

PORT="${E2E_SERVER_PORT:-3333}"
PG_PORT=5433
PG_CONTAINER="gavel-e2e-postgres"
PG_USER="gavel"
PG_PASS="gavel"
PG_DB="gaveltest"
DATABASE_URL="postgres://${PG_USER}:${PG_PASS}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable"

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
SERVER_BIN="${ROOT_DIR}/bazel-bin/apps/server/cmd/gavel-server/gavel-server_/gavel-server"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

if [[ ! -x "$SERVER_BIN" ]]; then
  echo "ERROR: Server binary not found at $SERVER_BIN" >&2
  echo "Run 'bazel build //apps/server/cmd/gavel-server' first." >&2
  exit 1
fi

if podman inspect "$PG_CONTAINER" &>/dev/null; then
  podman start "$PG_CONTAINER" 2>/dev/null || true
else
  podman run -d --name "$PG_CONTAINER" \
    -e POSTGRES_USER="$PG_USER" \
    -e POSTGRES_PASSWORD="$PG_PASS" \
    -e POSTGRES_DB="$PG_DB" \
    -p "${PG_PORT}:5432" \
    postgres:16-alpine
fi

echo "Waiting for PostgreSQL..." >&2
for i in $(seq 1 30); do
  if podman exec "$PG_CONTAINER" pg_isready -U "$PG_USER" &>/dev/null; then
    break
  fi
  if [[ $i -eq 30 ]]; then
    echo "ERROR: PostgreSQL did not become ready in time" >&2
    exit 1
  fi
  sleep 1
done
echo "PostgreSQL ready on port $PG_PORT" >&2

podman exec "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -c "SELECT 1" &>/dev/null

export GAVEL_DATABASE_URL="$DATABASE_URL"
export GAVEL_ADDR=":${PORT}"
export GAVEL_DATA_DIR="/tmp/gavel-e2e-data"
export GAVEL_SECURE_COOKIES="false"

mkdir -p "$GAVEL_DATA_DIR"

cd "$ROOT_DIR"
exec "$SERVER_BIN" serve
