package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const bridgeFindingsJSON = `{"projects":[{"name":"payments","findings":[
	{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"a.go","line":5,"message":"unchecked","fingerprint":"f1","provenance":"catalog"},
	{"tool":"shellcheck","rule_id":"SC2086","severity":"warning","file_path":"a.sh","line":2,"message":"quote it","fingerprint":"f2","provenance":"bridge"}
]}]}`

func TestRunFindings_MarksBridgeFindingsAsNotGuaranteed(t *testing.T) {
	result, err := runFindings(context.Background(), fakeCLI(t, bridgeFindingsJSON), FindingsInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "a.sh:2 [warning] quote it (shellcheck:SC2086) [bridge — not guaranteed by gavel]")
	assert.NotContains(t, result, "unchecked (golangci-lint:errcheck) [bridge")
}

func TestRunLintFile_MarksBridgeFindingsAsNotGuaranteed(t *testing.T) {
	result, err := runLintFile(context.Background(), fakeCLI(t, bridgeFindingsJSON), LintFileInput{File: "a.sh"})

	require.NoError(t, err)
	assert.Contains(t, result, "(shellcheck:SC2086) [bridge — not guaranteed by gavel]")
}

func TestRunJudge_ListsToolFailuresAndMarksBridgeOnes(t *testing.T) {
	judgeJSON := `{"projects":[{"name":"payments","verdict":"fail","findings_count":0,"violations_count":0,"rulings":[],
		"tool_failures":[
			{"tool":"shellcheck","reason":"the build produced no SARIF files","degraded":false,"provenance":"bridge"},
			{"tool":"ErrorProne","reason":"partial","degraded":true,"provenance":"catalog"}]}]}`

	result, err := runJudge(context.Background(), fakeCLIWithExitCode(t, judgeJSON, 1), JudgeInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "Tool failures:")
	assert.Contains(t, result, "shellcheck [bridge — not guaranteed by gavel]: the build produced no SARIF files")
	assert.Contains(t, result, "ErrorProne (incomplete): partial")
	assert.NotContains(t, result, "ErrorProne [bridge")
}
