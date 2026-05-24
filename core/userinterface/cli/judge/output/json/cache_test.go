package json_test

import (
	encjson "encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/json"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
)

func TestWriteCache_CreatesVerdictFile(t *testing.T) {
	workspace := t.TempDir()
	results := []pipeline.Result{
		{
			Name:            "core",
			Verdict:         "pass",
			CommitSHA:       "abc123",
			Branch:          "main",
			StartedAt:       time.Date(2025, 6, 20, 10, 0, 0, 0, time.UTC),
			FindingsCount:   5,
			CoveragePercent: 85.0,
		},
	}

	err := json.WriteCache(workspace, results)
	require.NoError(t, err)

	path := filepath.Join(workspace, ".gavel", "results", "core", "verdict.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, encjson.Unmarshal(data, &got))
	assert.Equal(t, "core", got["name"])
	assert.Equal(t, "pass", got["verdict"])
	assert.Equal(t, "abc123", got["commit_sha"])
	assert.Equal(t, "main", got["branch"])
	assert.Equal(t, "2025-06-20T10:00:00Z", got["started_at"])
	assert.Equal(t, float64(5), got["findings_count"])
	assert.Equal(t, 85.0, got["coverage_percent"])
}

func TestWriteCache_WritesMultipleProjects(t *testing.T) {
	workspace := t.TempDir()
	results := []pipeline.Result{
		{Name: "core", Verdict: "pass", CommitSHA: "abc123", Branch: "main"},
		{Name: "server", Verdict: "fail", CommitSHA: "abc123", Branch: "main"},
	}

	err := json.WriteCache(workspace, results)
	require.NoError(t, err)

	for _, name := range []string{"core", "server"} {
		path := filepath.Join(workspace, ".gavel", "results", name, "verdict.json")
		_, err := os.Stat(path)
		assert.NoError(t, err, "expected verdict.json for project %s", name)
	}
}

func TestWriteCache_OverwritesExisting(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, ".gavel", "results", "core")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "verdict.json"), []byte(`{"old":true}`), 0o644))

	results := []pipeline.Result{
		{Name: "core", Verdict: "fail", CommitSHA: "def456", Branch: "feat"},
	}

	err := json.WriteCache(workspace, results)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "verdict.json"))
	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(data, &got))
	assert.Equal(t, "fail", got["verdict"])
	assert.Equal(t, "def456", got["commit_sha"])
}

func TestWriteCache_MkdirAllError(t *testing.T) {
	results := []pipeline.Result{
		{Name: "core", Verdict: "pass"},
	}

	err := json.WriteCache("/dev/null/nope", results)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cache core")
}

func TestWriteCache_EmptyResults(t *testing.T) {
	workspace := t.TempDir()

	err := json.WriteCache(workspace, nil)

	require.NoError(t, err)
}

func TestWriteCache_PropagatesFirstError(t *testing.T) {
	results := []pipeline.Result{
		{Name: "a", Verdict: "pass"},
		{Name: "b", Verdict: "pass"},
	}

	err := json.WriteCache("/dev/null/nope", results)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cache a")
}
