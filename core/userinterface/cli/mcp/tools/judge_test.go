package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatJudgeOutput_PassingVerdict(t *testing.T) {
	input := []byte(`{"projects":[{"name":"core","verdict":"pass","findings_count":0,"violations_count":0,"coverage_percent":85.3,"rulings":[{"subtype":"findings","passed":true,"detail":"0 findings"}]}]}`)

	result, err := formatJudgeOutput(input, 0)

	require.NoError(t, err)
	assert.Contains(t, result, "core — PASS")
	assert.Contains(t, result, "Coverage: 85.3%")
	assert.Contains(t, result, "findings: PASS")
}

func TestFormatJudgeOutput_FailingVerdict(t *testing.T) {
	input := []byte(`{"projects":[{"name":"cli","verdict":"fail","findings_count":5,"violations_count":2,"rulings":[{"subtype":"findings","passed":false,"detail":"5 > 0"}]}]}`)

	result, err := formatJudgeOutput(input, 1)

	require.NoError(t, err)
	assert.Contains(t, result, "cli — FAIL")
	assert.Contains(t, result, "Findings: 5")
	assert.Contains(t, result, "Architecture violations: 2")
	assert.Contains(t, result, "Quality gate FAILED")
}

func TestFormatJudgeOutput_WithDelta(t *testing.T) {
	input := []byte(`{"projects":[{"name":"core","verdict":"pass","findings_count":3,"violations_count":0,"delta":{"has_previous":true,"new_count":1,"fixed_count":2}}]}`)

	result, err := formatJudgeOutput(input, 0)

	require.NoError(t, err)
	assert.Contains(t, result, "Delta: 1 new, 2 fixed")
}

func TestFormatJudgeOutput_InvalidJSON(t *testing.T) {
	result, err := formatJudgeOutput([]byte("not json"), 0)

	require.NoError(t, err)
	assert.Equal(t, "not json", result)
}

func TestFormatJudgeOutput_WithTrees(t *testing.T) {
	input := []byte(`{"projects":[{
		"name":"core","verdict":"pass","findings_count":3,"violations_count":0,
		"coverage_percent":80.0,
		"coverage_tree":{
			"path":"","covered_lines":8,"total_lines":10,"percent":80.0,
			"children":[
				{"path":"pkg","covered_lines":8,"total_lines":10,"percent":80.0,
				 "files":[{"name":"a.go","covered_lines":8,"total_lines":10,"percent":80.0}]}
			]
		},
		"findings_tree":{
			"path":"","count":3,"by_severity":{"error":2,"warning":1},
			"children":[
				{"path":"pkg","count":3,"by_severity":{"error":2,"warning":1},
				 "files":[{"name":"a.go","count":3,"by_severity":{"error":2,"warning":1}}]}
			]
		}
	}]}`)

	result, err := formatJudgeOutput(input, 0)

	require.NoError(t, err)
	assert.Contains(t, result, "Coverage by directory:")
	assert.Contains(t, result, "pkg/")
	assert.Contains(t, result, "80.0%")
	assert.Contains(t, result, "Findings by directory:")
	assert.Contains(t, result, "3 findings")
	assert.Contains(t, result, "2 error, 1 warning")
}
