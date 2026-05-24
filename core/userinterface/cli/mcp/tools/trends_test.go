package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatTrendsOutput_WithData(t *testing.T) {
	input := `[
		{"commit_sha":"abc1234","branch":"main","coverage_percent":92.3,"total_findings":45,"new_findings":0,"resolved_findings":2,"verdict_outcome":"pass","created_at":"2026-06-14T10:00:00Z"},
		{"commit_sha":"def5678","branch":"main","coverage_percent":91.8,"total_findings":47,"new_findings":3,"resolved_findings":1,"verdict_outcome":"fail","created_at":"2026-06-13T10:00:00Z"},
		{"commit_sha":"ghi9012","branch":"main","coverage_percent":92.1,"total_findings":45,"new_findings":0,"resolved_findings":0,"verdict_outcome":"pass","created_at":"2026-06-12T10:00:00Z"}
	]`

	result, err := formatTrendsOutput([]byte(input), "core")

	require.NoError(t, err)
	assert.Contains(t, result, "core")
	assert.Contains(t, result, "abc1234")
	assert.Contains(t, result, "92.3%")
	assert.Contains(t, result, "pass")
	assert.Contains(t, result, "fail")
	assert.Contains(t, result, "Pass rate: 2/3")
}

func TestFormatTrendsOutput_EmptyData(t *testing.T) {
	result, err := formatTrendsOutput([]byte("[]"), "core")

	require.NoError(t, err)
	assert.Contains(t, result, "No analysis history")
	assert.Contains(t, result, "core")
}

func TestFormatTrendsOutput_ComputesCoverageTrend(t *testing.T) {
	input := `[
		{"commit_sha":"a","branch":"main","coverage_percent":95.0,"total_findings":10,"new_findings":0,"resolved_findings":0,"verdict_outcome":"pass","created_at":"2026-06-14T10:00:00Z"},
		{"commit_sha":"b","branch":"main","coverage_percent":90.0,"total_findings":15,"new_findings":0,"resolved_findings":0,"verdict_outcome":"pass","created_at":"2026-06-13T10:00:00Z"}
	]`

	result, err := formatTrendsOutput([]byte(input), "core")

	require.NoError(t, err)
	assert.Contains(t, result, "+5.0%")
	assert.Contains(t, result, "Findings: 10")
}

func TestFormatTrendsOutput_InvalidJSON(t *testing.T) {
	_, err := formatTrendsOutput([]byte("not json"), "core")

	assert.Error(t, err)
}
