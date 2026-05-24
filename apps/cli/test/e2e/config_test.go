//go:build e2e

package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_OutputsJSON(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	result := runGavel(t, "config")
	require.Equal(t, 0, result.ExitCode,
		"config should exit 0\nstdout: %s\nstderr: %s", result.Stdout, result.Stderr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Stdout), &parsed),
		"stdout should be valid JSON, got:\n%s", result.Stdout)

	t.Run("gavelspace_name", func(t *testing.T) {
		assert.Equal(t, "example-go-repo", parsed["gavelspace"])
	})

	t.Run("projects_array", func(t *testing.T) {
		projects, ok := parsed["projects"].([]any)
		require.True(t, ok && len(projects) > 0)

		first, ok := projects[0].(map[string]any)
		require.True(t, ok)
		assert.NotEmpty(t, first["name"])

		langs, ok := first["languages"].([]any)
		assert.True(t, ok && len(langs) > 0)
	})

	t.Run("config_path", func(t *testing.T) {
		configPath, ok := parsed["config_path"].(string)
		assert.True(t, ok)
		assert.True(t, strings.Contains(configPath, "gavel.yaml"))
	})
}

func TestProjects_OutputsJSON(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	result := runGavel(t, "projects")
	require.Equal(t, 0, result.ExitCode,
		"projects should exit 0\nstderr: %s", result.Stderr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Stdout), &parsed))

	t.Run("has_projects", func(t *testing.T) {
		projects, ok := parsed["projects"].([]any)
		require.True(t, ok && len(projects) > 0)

		first, ok := projects[0].(map[string]any)
		require.True(t, ok)
		assert.NotEmpty(t, first["name"])

		gate, ok := first["quality_gate"].(map[string]any)
		assert.True(t, ok)
		rules, ok := gate["rules"].([]any)
		assert.True(t, ok && len(rules) > 0)
	})

	t.Run("has_architecture", func(t *testing.T) {
		arch, ok := parsed["architecture"].(map[string]any)
		require.True(t, ok)

		layers, ok := arch["layers"].([]any)
		assert.True(t, ok && len(layers) > 0)

		denyRules, ok := arch["deny_rules"].([]any)
		assert.True(t, ok && len(denyRules) > 0)
	})
}

func TestProjects_BaselineFilePersistedAfterJudge(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	runGavel(t, "judge", "--project=api-gateway")

	baselinePath := filepath.Join(examplesGoRepo(t), ".gavel", "baseline", "api-gateway", "findings")
	data, err := os.ReadFile(baselinePath)
	require.NoError(t, err, "baseline file should exist after judge")

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Greater(t, len(lines), 0, "baseline should contain fingerprints")

	result := runGavel(t, "projects")
	require.Equal(t, 0, result.ExitCode)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Stdout), &parsed))

	projects, ok := parsed["projects"].([]any)
	require.True(t, ok && len(projects) > 0)
	proj, ok := projects[0].(map[string]any)
	require.True(t, ok)
	_, hasBaseline := proj["baseline"]
	assert.True(t, hasBaseline, "projects output should include baseline key")
}
