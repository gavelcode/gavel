package json_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	outputjson "github.com/usegavel/gavel/core/userinterface/cli/judge/output/json"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
)

func TestLoadReadsVerdictsWrittenByWriteCache(t *testing.T) {
	workspace := t.TempDir()
	result := pipeline.Result{
		Name:            "core",
		Verdict:         "pass",
		CommitSHA:       "abc123",
		Branch:          "main",
		FindingsCount:   1,
		CoveragePercent: 94.7,
		Findings: []evidencedto.Finding{{
			Tool:          "golangci-lint",
			RuleID:        "errcheck",
			Severity:      "error",
			FilePath:      "core/x.go",
			Line:          10,
			Message:       "unchecked error",
			FingerprintID: "fp1",
		}},
		Delta: pipeline.Delta{
			HasPrevious:     true,
			NewCount:        1,
			NewFingerprints: map[string]bool{"fp1": true},
		},
	}
	require.NoError(t, outputjson.WriteCache(workspace, []pipeline.Result{result}))

	verdicts, _, err := outputjson.Load(workspace)
	require.NoError(t, err)
	require.Len(t, verdicts, 1)

	got := verdicts[0]
	assert.Equal(t, "core", got.Name)
	assert.Equal(t, "pass", got.Verdict)
	assert.Equal(t, "abc123", got.CommitSHA)
	assert.Equal(t, "main", got.Branch)
	require.NotNil(t, got.CoveragePercent)
	assert.InDelta(t, 94.7, *got.CoveragePercent, 0.01)

	require.Len(t, got.Findings, 1)
	finding := got.Findings[0]
	assert.Equal(t, "golangci-lint", finding.Tool)
	assert.Equal(t, "errcheck", finding.RuleID)
	assert.Equal(t, "error", finding.Severity)
	assert.Equal(t, "core/x.go", finding.FilePath)
	assert.Equal(t, 10, finding.Line)
	assert.Equal(t, "unchecked error", finding.Message)
	assert.Equal(t, "fp1", finding.FingerprintID)
	assert.Equal(t, "new", finding.Status)

	require.NotNil(t, got.Delta)
	assert.Equal(t, 1, got.Delta.NewCount)
}

func TestLoadParsesExistingStatusAndAbsentCoverage(t *testing.T) {
	workspace := t.TempDir()
	writeVerdictFixture(t, workspace, "web", `{
      "name": "web",
      "verdict": "fail",
      "commit_sha": "deadbeef",
      "branch": "feature",
      "findings": [
        {"tool":"eslint","rule_id":"no-unused","severity":"warning",
         "file_path":"apps/web/x.ts","line":5,"message":"unused",
         "fingerprint":"fp9","status":"existing"}
      ],
      "delta": {"has_previous": true, "new_count": 0, "fixed_count": 2, "existing_count": 1}
    }`)

	verdicts, _, err := outputjson.Load(workspace)
	require.NoError(t, err)
	require.Len(t, verdicts, 1)

	got := verdicts[0]
	assert.Equal(t, "web", got.Name)
	assert.Equal(t, "fail", got.Verdict)
	assert.Nil(t, got.CoveragePercent, "absent coverage_percent must read as nil")
	require.Len(t, got.Findings, 1)
	assert.Equal(t, "existing", got.Findings[0].Status)
	require.NotNil(t, got.Delta)
	assert.Equal(t, 2, got.Delta.FixedCount)
}

func TestLoadReturnsErrNoResultsWhenDirAbsent(t *testing.T) {
	_, _, err := outputjson.Load(t.TempDir())
	assert.ErrorIs(t, err, outputjson.ErrNoResults)
}

func TestLoadSkipsUnparseableFilesAndReportsThem(t *testing.T) {
	workspace := t.TempDir()
	writeVerdictFixture(t, workspace, "core", `{"name":"core","verdict":"pass"}`)
	writeVerdictFixture(t, workspace, "web", `{ not json`)

	verdicts, skipped, err := outputjson.Load(workspace)
	require.NoError(t, err)
	require.Len(t, verdicts, 1)
	assert.Equal(t, "core", verdicts[0].Name)
	require.Len(t, skipped, 1)
	assert.Contains(t, skipped[0], "web")
}

func TestLoadReturnsErrNoResultsWhenEveryFileCorrupt(t *testing.T) {
	workspace := t.TempDir()
	writeVerdictFixture(t, workspace, "web", `{ not json`)

	_, skipped, err := outputjson.Load(workspace)
	assert.ErrorIs(t, err, outputjson.ErrNoResults)
	assert.Len(t, skipped, 1)
}

func TestLoadSkipsNonDirectoryEntriesAndEmptyProjects(t *testing.T) {
	workspace := t.TempDir()
	results := filepath.Join(workspace, ".gavel", "results")
	require.NoError(t, os.MkdirAll(filepath.Join(results, "empty"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(results, "stray.txt"), []byte("x"), 0o644))
	writeVerdictFixture(t, workspace, "core", `{"name":"core","verdict":"pass"}`)

	verdicts, skipped, err := outputjson.Load(workspace)
	require.NoError(t, err)
	require.Len(t, verdicts, 1)
	assert.Equal(t, "core", verdicts[0].Name)
	assert.Empty(t, skipped)
}

func TestLoadMapsViolations(t *testing.T) {
	workspace := t.TempDir()
	writeVerdictFixture(t, workspace, "core", `{"name":"core","verdict":"fail","commit_sha":"c1",
		"violations":[{"rule":"deny","source_pkg":"a","target_pkg":"b","message":"forbidden","status":"new"}]}`)

	verdicts, _, err := outputjson.Load(workspace)
	require.NoError(t, err)
	require.Len(t, verdicts, 1)
	require.Len(t, verdicts[0].Violations, 1)
	assert.Equal(t, "deny", verdicts[0].Violations[0].Rule)
	assert.Equal(t, "forbidden", verdicts[0].Violations[0].Message)
}

func TestLoadErrorsWhenResultsPathIsNotADirectory(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, ".gavel"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, ".gavel", "results"), []byte("x"), 0o644))

	_, _, err := outputjson.Load(workspace)
	require.Error(t, err)
	assert.NotErrorIs(t, err, outputjson.ErrNoResults)
}

func TestLoadErrorsWhenVerdictFileCannotBeRead(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, ".gavel", "results", "core")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "verdict.json"), 0o755))

	_, _, err := outputjson.Load(workspace)
	require.Error(t, err)
	assert.NotErrorIs(t, err, outputjson.ErrNoResults)
}

func writeVerdictFixture(t *testing.T, workspace, project, body string) {
	t.Helper()
	dir := filepath.Join(workspace, ".gavel", "results", project)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "verdict.json"), []byte(body), 0o644))
}
