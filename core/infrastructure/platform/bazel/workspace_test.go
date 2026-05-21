package bazel_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel"
)

func TestWorkspaceDir_RejectsNonBazelDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BUILD_WORKSPACE_DIRECTORY", dir)

	_, err := bazel.WorkspaceDir()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a Bazel workspace")
}

func TestWorkspaceDir_AcceptsModuleBazel(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "MODULE.bazel"), []byte(""), 0o644))
	t.Setenv("BUILD_WORKSPACE_DIRECTORY", dir)

	got, err := bazel.WorkspaceDir()

	require.NoError(t, err)
	assert.Equal(t, dir, got)
}

func TestWorkspaceDir_AcceptsWorkspaceFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "WORKSPACE"), []byte(""), 0o644))
	t.Setenv("BUILD_WORKSPACE_DIRECTORY", dir)

	got, err := bazel.WorkspaceDir()

	require.NoError(t, err)
	assert.Equal(t, dir, got)
}

func TestWorkspaceDir_AcceptsWorkspaceBazelFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "WORKSPACE.bazel"), []byte(""), 0o644))
	t.Setenv("BUILD_WORKSPACE_DIRECTORY", dir)

	got, err := bazel.WorkspaceDir()

	require.NoError(t, err)
	assert.Equal(t, dir, got)
}

func TestWorkspaceDir_FallsBackToGetwd(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "MODULE.bazel"), []byte(""), 0o644))
	t.Setenv("BUILD_WORKSPACE_DIRECTORY", "")
	t.Chdir(dir)

	got, err := bazel.WorkspaceDir()

	require.NoError(t, err)
	assert.Equal(t, dir, got)
}
