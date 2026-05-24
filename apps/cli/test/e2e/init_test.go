//go:build e2e

package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_FromExistingConfig(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	result := runGavel(t, "init", "--from=.gavel/gavel.yaml")
	assert.Equal(t, 0, result.ExitCode, "init exit code should be 0\nstdout: %s\nstderr: %s", result.Stdout, result.Stderr)

	workspace := examplesGoRepo(t)

	bazelrcPath := filepath.Join(workspace, ".gavel", "gavel.bazelrc")
	bazelrcContent, err := os.ReadFile(bazelrcPath)
	require.NoError(t, err, ".gavel/gavel.bazelrc should exist after init")
	assert.True(t, strings.Contains(string(bazelrcContent), "aspects"),
		"gavel.bazelrc should contain aspect registrations, got:\n%s", string(bazelrcContent))

	modulePath := filepath.Join(workspace, ".gavel", "gavel.MODULE.bazel")
	moduleContent, err := os.ReadFile(modulePath)
	require.NoError(t, err, ".gavel/gavel.MODULE.bazel should exist after init")
	assert.True(t, strings.Contains(string(moduleContent), "use_repo_rule") || strings.Contains(string(moduleContent), "bazel_dep"),
		"gavel.MODULE.bazel should contain dependency entries, got:\n%s", string(moduleContent))
}

func TestInit_ThenValidate(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	initResult := runGavel(t, "init", "--from=.gavel/gavel.yaml")
	require.Equal(t, 0, initResult.ExitCode, "init should succeed\nstderr: %s", initResult.Stderr)

	validateResult := runGavel(t, "validate")
	assert.Equal(t, 0, validateResult.ExitCode,
		"validate should pass after init\nstdout: %s\nstderr: %s", validateResult.Stdout, validateResult.Stderr)
}
