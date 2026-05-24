#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../.."

# Restore BUILD.bazel if gazelle wiped it (gotags=e2e confuses gazelle)
git checkout -- apps/cli/test/e2e/BUILD.bazel 2>/dev/null || true

echo "Building gavel binary and test binary..."
bazel build //apps/cli/cmd/gavel //apps/cli/test/e2e:e2e_test 2>&1 | tail -3

GAVEL_BINARY="$(pwd)/$(bazel cquery //apps/cli/cmd/gavel --output=files 2>/dev/null | tail -1)"
TEST_BINARY="$(pwd)/$(bazel cquery //apps/cli/test/e2e:e2e_test --output=files 2>/dev/null | tail -1)"

echo "Gavel binary: $GAVEL_BINARY"
echo "Test binary:  $TEST_BINARY"
echo "Running E2E tests..."
echo ""

export GAVEL_BINARY
export BUILD_WORKSPACE_DIRECTORY="$(pwd)"

exec "$TEST_BINARY" -test.v -test.timeout=900s "$@"
