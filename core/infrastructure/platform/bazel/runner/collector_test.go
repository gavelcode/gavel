package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
)

func TestAspectCollector_NilAspects(t *testing.T) {
	collector := runner.NewAspectCollector(nil, runner.NewExecRunner())

	reports, err := collector.Collect(context.Background(), t.TempDir(), []string{"//..."})

	require.NoError(t, err)
	assert.Nil(t, reports)
}

func TestCollectReportsFromDir_FindsMatchingFiles(t *testing.T) {
	dir := t.TempDir()
	sarifData := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg_target.AspectRulesLint.report"), sarifData, 0o644))

	reports, err := runner.CollectReportsFromDir(dir)

	require.NoError(t, err)
	require.Len(t, reports, 1)
	assert.Equal(t, "golangci-lint", reports[0].Source)
	assert.Equal(t, sarifData, reports[0].Data)
}

func TestCollectReportsFromDir_IgnoresNonMatchingFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "regular.sarif"), []byte(`{}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "output.txt"), []byte("hello"), 0o644))

	reports, err := runner.CollectReportsFromDir(dir)

	require.NoError(t, err)
	assert.Nil(t, reports)
}

func TestCollectReportsFromDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	reports, err := runner.CollectReportsFromDir(dir)

	require.NoError(t, err)
	assert.Nil(t, reports)
}

func TestCollectReportsFromDir_NestedDirectories(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "pkg", "sub")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	sarifData := []byte(`{"runs":[{"tool":{"driver":{"name":"PMD"}},"results":[]}]}`)
	require.NoError(t, os.WriteFile(filepath.Join(nested, "sub_target.AspectRulesLint.report"), sarifData, 0o644))

	reports, err := runner.CollectReportsFromDir(dir)

	require.NoError(t, err)
	require.Len(t, reports, 1)
	assert.Equal(t, "PMD", reports[0].Source)
}

func TestCollectReportsFromDir_MultipleReports(t *testing.T) {
	dir := t.TempDir()
	pmdData := []byte(`{"runs":[{"tool":{"driver":{"name":"PMD"}},"results":[]}]}`)
	ruffData := []byte(`{"runs":[{"tool":{"driver":{"name":"ruff"}},"results":[]}]}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.AspectRulesLint.report"), pmdData, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.AspectRulesLint.report"), ruffData, 0o644))

	reports, err := runner.CollectReportsFromDir(dir)

	require.NoError(t, err)
	assert.Len(t, reports, 2)
}

func TestExtractToolName_ValidSARIF(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"SpotBugs"}},"results":[]}]}`)

	got := runner.ExtractToolName(data, "fallback.report")

	assert.Equal(t, "SpotBugs", got)
}

func TestExtractToolName_InvalidJSON(t *testing.T) {
	got := runner.ExtractToolName([]byte("not json"), "myfile.report")

	assert.Equal(t, "myfile.report", got)
}

func TestExtractToolName_EmptyToolName(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":""}},"results":[]}]}`)

	got := runner.ExtractToolName(data, "fallback.report")

	assert.Equal(t, "fallback.report", got)
}

func TestExtractToolName_NoRuns(t *testing.T) {
	data := []byte(`{"runs":[]}`)

	got := runner.ExtractToolName(data, "fallback.report")

	assert.Equal(t, "fallback.report", got)
}

func TestHasRulesLintReports_Found(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg.AspectRulesLint.report"), []byte(`{}`), 0o644))

	got := runner.HasRulesLintReportsInDir(dir)

	assert.True(t, got)
}

func TestHasRulesLintReports_NotFound(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "regular.sarif"), []byte(`{}`), 0o644))

	got := runner.HasRulesLintReportsInDir(dir)

	assert.False(t, got)
}

func TestHasRulesLintReports_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	got := runner.HasRulesLintReportsInDir(dir)

	assert.False(t, got)
}

var _ runner.FindingsCollector = (*runner.AspectCollector)(nil)
var _ runner.FindingsCollector = (*runner.ReportCollector)(nil)
var _ runner.FindingsCollector = (*runner.HybridCollector)(nil)
