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

	verdicts, err := outputjson.Load(workspace)
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

	verdicts, err := outputjson.Load(workspace)
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
	_, err := outputjson.Load(t.TempDir())
	assert.ErrorIs(t, err, outputjson.ErrNoResults)
}

func writeVerdictFixture(t *testing.T, workspace, project, body string) {
	t.Helper()
	dir := filepath.Join(workspace, ".gavel", "results", project)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "verdict.json"), []byte(body), 0o644))
}
