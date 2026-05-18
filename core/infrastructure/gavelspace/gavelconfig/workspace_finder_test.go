package gavelconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/gavelspace/gavelconfig"
)

func TestWorkspaceFinderLoadFromConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "gavel.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
name: test-mono
projects:
  - name: api
    pattern: //api/...
`), 0644))

	finder := gavelconfig.NewWorkspaceFinder()
	gs, projects, err := finder.LoadFromConfig(configPath)

	require.NoError(t, err)
	assert.Equal(t, "test-mono", gs.ID().String())
	require.Len(t, projects, 1)
	assert.Equal(t, "api", projects[0].Name())
	assert.Equal(t, "//api/...", projects[0].TargetPattern())
}

func TestWorkspaceFinderLoadFromConfigReturnsErrorForMissingFile(t *testing.T) {
	finder := gavelconfig.NewWorkspaceFinder()

	_, _, err := finder.LoadFromConfig("/nonexistent/gavel.yaml")

	require.Error(t, err)
	assert.ErrorIs(t, err, gavelconfig.ErrReadConfig)
}
