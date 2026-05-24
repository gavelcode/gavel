//go:build e2e

package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJudge_LocalMode(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	result := runGavel(t, "judge", "--json", "--project=api-gateway")

	assert.Equal(t, 1, result.ExitCode, "gate should fail (findings > 0, coverage < 90%%)")

	var response struct {
		Projects []struct {
			Name          string   `json:"name"`
			Verdict       string   `json:"verdict"`
			FindingsCount int      `json:"findings_count"`
			CoveragePct   *float64 `json:"coverage_percent"`
		} `json:"projects"`
	}
	require.NoError(t, json.Unmarshal([]byte(result.Stdout), &response), "stdout must be valid JSON")
	require.Len(t, response.Projects, 1)

	proj := response.Projects[0]

	t.Run("verdict", func(t *testing.T) {
		assert.Equal(t, "fail", proj.Verdict)
	})

	t.Run("project_name", func(t *testing.T) {
		assert.Equal(t, "api-gateway", proj.Name)
	})

	t.Run("findings", func(t *testing.T) {
		assert.Greater(t, proj.FindingsCount, 0)
	})

	t.Run("coverage", func(t *testing.T) {
		require.NotNil(t, proj.CoveragePct)
		assert.Greater(t, *proj.CoveragePct, 0.0)
	})

	t.Run("baseline_persisted", func(t *testing.T) {
		baselinePath := filepath.Join(examplesGoRepo(t), ".gavel", "baseline", "api-gateway", "findings")
		data, err := os.ReadFile(baselinePath)
		require.NoError(t, err, "baseline findings file should exist after judge")
		assert.Greater(t, len(data), 0, "baseline should contain fingerprints")
	})

}

func TestJudge_QuickMode(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	result := runGavel(t, "judge", "--json", "--quick", "--project=api-gateway")

	var response struct {
		Projects []struct {
			CoveragePct *float64 `json:"coverage_percent"`
		} `json:"projects"`
	}
	require.NoError(t, json.Unmarshal([]byte(result.Stdout), &response))
	require.Len(t, response.Projects, 1)
	assert.Nil(t, response.Projects[0].CoveragePct, "quick mode should skip coverage")
}

func TestJudge_ServerMode(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	sf := startServer(t)

	result := runGavel(t, "judge", "--project=api-gateway",
		"--server="+sf.URL, "--token="+sf.Token)

	require.True(t, result.ExitCode == 0 || result.ExitCode == 1,
		"judge should run; stderr: %s", result.Stderr)

	t.Run("server_submission_succeeds", func(t *testing.T) {
		assert.NotContains(t, result.Stderr, "server submission failed")
	})

	t.Run("trends_shows_history", func(t *testing.T) {
		trendsResult := runGavel(t, "trends", "--project=api-gateway",
			"--server="+sf.URL, "--token="+sf.Token)
		assert.Equal(t, 0, trendsResult.ExitCode,
			"trends should succeed; stderr: %s", trendsResult.Stderr)
		assert.Contains(t, trendsResult.Stdout, "api-gateway")
	})

	t.Run("trends_json_output", func(t *testing.T) {
		trendsResult := runGavel(t, "trends", "--project=api-gateway",
			"--server="+sf.URL, "--token="+sf.Token, "--json")
		require.Equal(t, 0, trendsResult.ExitCode)

		var entries []map[string]any
		require.NoError(t, json.Unmarshal([]byte(trendsResult.Stdout), &entries))
		require.NotEmpty(t, entries)
		assert.Contains(t, entries[0], "commit_sha")
		assert.Contains(t, entries[0], "verdict_outcome")
		assert.Contains(t, entries[0], "total_findings")
	})
}

func TestJudge_ServerMode_AutoCreateProject(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	sf := startServer(t)

	result := runGavel(t, "judge", "--quick", "--project=api-gateway",
		"--server="+sf.URL, "--token="+sf.Token)

	require.True(t, result.ExitCode == 0 || result.ExitCode == 1,
		"judge should run; stderr: %s", result.Stderr)
	assert.NotContains(t, result.Stderr, "server submission failed",
		"project should be auto-created on server")
}

func TestJudge_ServerMode_GracefulFallback(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	result := runGavel(t, "judge", "--quick", "--project=api-gateway",
		"--server=http://127.0.0.1:1", "--token=fake")

	assert.True(t, result.ExitCode == 0 || result.ExitCode == 1,
		"should produce a local verdict (0 or 1), not a server error")
	assert.Contains(t, result.Stderr, "server",
		"stderr should warn about server failure")
}

func TestJudge_ServerMode_RequireSubmitFails(t *testing.T) {
	cleanWorkspace(t)
	t.Cleanup(func() { cleanWorkspace(t) })

	result := runGavel(t, "judge", "--quick", "--project=api-gateway",
		"--server=http://127.0.0.1:1", "--token=fake", "--require-submit")

	assert.NotEqual(t, 0, result.ExitCode)
	combined := result.Stdout + result.Stderr
	assert.Contains(t, combined, "submit")
}
