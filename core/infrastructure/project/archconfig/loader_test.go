package archconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/project/archconfig"
)

func TestNewPolicyLoaderShouldReturnNonNil(t *testing.T) {
	loader := archconfig.NewPolicyLoader()
	require.NotNil(t, loader)
}

func TestLoadPolicyShouldParsePolicyFromWorkspace(t *testing.T) {
	workspace := t.TempDir()
	gavelDir := filepath.Join(workspace, ".gavel")
	require.NoError(t, os.MkdirAll(gavelDir, 0o755))

	content := []byte(`
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
rules:
  - name: domain-purity
    source: domain
    deny: [application]
`)
	require.NoError(t, os.WriteFile(filepath.Join(gavelDir, "architecture.yml"), content, 0o644))

	loader := archconfig.NewPolicyLoader()
	policy, err := loader.LoadPolicy(workspace)
	require.NoError(t, err)

	assert.Len(t, policy.Layers(), 2)
	assert.Len(t, policy.DenyRules(), 1)
}

func TestLoadPolicyShouldReturnErrorForNonExistentWorkspace(t *testing.T) {
	loader := archconfig.NewPolicyLoader()
	_, err := loader.LoadPolicy("/nonexistent/workspace")
	require.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrReadConfig)
}
