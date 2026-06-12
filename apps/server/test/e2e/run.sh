#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../.."

git checkout -- apps/server/test/e2e/BUILD.bazel 2>/dev/null || true

echo "Building server binary, CLI binary, and test binary..."
bazel build //apps/server/cmd/gavel-server //apps/cli/cmd/gavel //apps/server/test/e2e:e2e_test 2>&1 | tail -3

GAVEL_SERVER_BINARY="$(pwd)/$(bazel cquery //apps/server/cmd/gavel-server --output=files 2>/dev/null | tail -1)"
GAVEL_BINARY="$(pwd)/$(bazel cquery //apps/cli/cmd/gavel --output=files 2>/dev/null | tail -1)"
TEST_BINARY="$(pwd)/$(bazel cquery //apps/server/test/e2e:e2e_test --output=files 2>/dev/null | tail -1)"

echo "Server binary: $GAVEL_SERVER_BINARY"
echo "CLI binary:    $GAVEL_BINARY"
echo "Test binary:   $TEST_BINARY"
echo "Running server E2E tests..."
echo ""

export GAVEL_SERVER_BINARY
export GAVEL_BINARY
export BUILD_WORKSPACE_DIRECTORY="$(pwd)"

exec "$TEST_BINARY" -test.v -test.timeout=300s "$@"
