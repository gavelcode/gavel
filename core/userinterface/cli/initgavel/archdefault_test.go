package initgavel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteArchitectureConfig_CreatesFile(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, ".gavel"), 0o755))

	err := writeArchitectureConfig(workspace, []string{"go"})

	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(workspace, architectureConfigFile))
	require.NoError(t, err)
	assert.Contains(t, string(data), "layers:")
	assert.Contains(t, string(data), "domain")
	assert.Contains(t, string(data), "domain-imports-nothing")
}

func TestWriteArchitectureConfig_FileExistsSkipsWrite(t *testing.T) {
	workspace := t.TempDir()
	archPath := filepath.Join(workspace, architectureConfigFile)
	require.NoError(t, os.MkdirAll(filepath.Dir(archPath), 0o755))
	require.NoError(t, os.WriteFile(archPath, []byte("existing"), 0o644))

	err := writeArchitectureConfig(workspace, []string{"go"})

	require.NoError(t, err)
	data, err := os.ReadFile(archPath)
	require.NoError(t, err)
	assert.Equal(t, "existing", string(data))
}

func TestWriteArchitectureConfig_NoLayersSkipsWrite(t *testing.T) {
	workspace := t.TempDir()

	err := writeArchitectureConfig(workspace, []string{"unknown-lang"})

	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(workspace, architectureConfigFile))
	assert.True(t, os.IsNotExist(err))
}

func TestPickLayers_Go(t *testing.T) {
	layers := pickLayers([]string{"go"})
	require.NotNil(t, layers)
	assert.Contains(t, layers, "domain")
}

func TestPickLayers_Java(t *testing.T) {
	layers := pickLayers([]string{"java"})
	require.NotNil(t, layers)
	assert.Contains(t, layers["domain"][0], "main/java")
}

func TestPickLayers_UnknownLanguage(t *testing.T) {
	layers := pickLayers([]string{"cobol"})
	assert.Nil(t, layers)
}

func TestPickLayers_EmptyTooling(t *testing.T) {
	layers := pickLayers(nil)
	assert.Nil(t, layers)
}

func TestRenderArchitectureYAML_ContainsAllLayers(t *testing.T) {
	layers := layersByLanguage["go"]

	result := renderArchitectureYAML(layers)

	assert.Contains(t, result, "layers:")
	assert.Contains(t, result, "domain:")
	assert.Contains(t, result, "application:")
	assert.Contains(t, result, "infrastructure:")
	assert.Contains(t, result, "userinterface:")
	assert.Contains(t, result, "rules:")
	assert.Contains(t, result, "domain-imports-nothing")
	assert.Contains(t, result, "detect_cycles: true")
}

func TestRenderArchitectureYAML_PartialLayers(t *testing.T) {
	layers := map[string][]string{
		"domain":      {"src/domain/..."},
		"application": {"src/app/...", "src/service/..."},
	}

	result := renderArchitectureYAML(layers)

	assert.Contains(t, result, "domain:")
	assert.Contains(t, result, "application:")
	assert.NotContains(t, result, "infrastructure:")
	assert.Contains(t, result, "src/app/...\", \"src/service/...")
}
