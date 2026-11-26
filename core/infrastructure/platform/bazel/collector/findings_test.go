package collector_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/collector"
)

var _ collectevidence.FindingsCollector = (*collector.BazelFindingsCollector)(nil)

func TestBazelFindingsCollector_NoReports_ReturnsEmpty(t *testing.T) {
	runner := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{}}
	parser := &fakeFindingsParser{}
	c := collector.NewBazelFindingsCollector(runner, parser)

	evidences, rawFiles, buildWarning, unanalyzed, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}}, nil)

	require.NoError(t, err)
	assert.Empty(t, evidences)
	assert.Empty(t, rawFiles)
	assert.Empty(t, buildWarning)
	assert.NotEmpty(t, unanalyzed,
		"a configured tool that produced no SARIF must be reported as having analyzed nothing, not silently omitted")
}

func TestBazelFindingsCollector_ParsesReports(t *testing.T) {
	sarif := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`)
	runner := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{"go_golangci_lint_submission_aspect": {sarif}}}
	parser := &fakeFindingsParser{returnEmpty: true}
	c := collector.NewBazelFindingsCollector(runner, parser)

	evidences, rawFiles, buildWarning, _, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}}, nil)

	require.NoError(t, err)
	assert.Len(t, evidences, 1)
	assert.Len(t, rawFiles, 1)
	assert.Equal(t, "sarif", rawFiles[0].Format)
	assert.Empty(t, buildWarning)
}

func TestBazelFindingsCollector_NoAspectsForEmptySelection(t *testing.T) {
	r := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{}}
	parser := &fakeFindingsParser{}
	c := collector.NewBazelFindingsCollector(r, parser)

	evidences, rawFiles, buildWarning, _, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{}, nil)

	require.NoError(t, err)
	assert.Nil(t, evidences)
	assert.Nil(t, rawFiles)
	assert.Empty(t, buildWarning)
}

func TestBazelFindingsCollector_RunnerError(t *testing.T) {
	r := &fakeAnalysisRunner{err: fmt.Errorf("bazel build failed")}
	parser := &fakeFindingsParser{}
	c := collector.NewBazelFindingsCollector(r, parser)

	_, _, _, _, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "run analysis")
}

func TestBazelFindingsCollector_ParserError(t *testing.T) {
	sarif := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`)
	r := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{"go_golangci_lint_submission_aspect": {sarif}}}
	parser := &fakeFindingsParser{err: fmt.Errorf("parse failed")}
	c := collector.NewBazelFindingsCollector(r, parser)

	_, _, _, _, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse failed")
}

func TestBazelFindingsCollector_EmptySARIFData(t *testing.T) {
	r := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{"go_golangci_lint_submission_aspect": {[]byte{}}}}
	parser := &fakeFindingsParser{}
	c := collector.NewBazelFindingsCollector(r, parser)

	_, _, _, _, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}}, nil)

	require.Error(t, err)
}

func TestBazelFindingsCollector_PropagatesBuildWarning(t *testing.T) {
	sarif := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`)
	runner := &fakeAnalysisRunner{
		sarifFiles:   map[string][][]byte{"go_golangci_lint_submission_aspect": {sarif}},
		buildWarning: fmt.Errorf("bazel build had failures"),
	}
	parser := &fakeFindingsParser{returnEmpty: true}
	c := collector.NewBazelFindingsCollector(runner, parser)

	evidences, rawFiles, buildWarning, _, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}}, nil)

	require.NoError(t, err)
	assert.Len(t, evidences, 1)
	assert.Len(t, rawFiles, 1)
	assert.Equal(t, "bazel build had failures", buildWarning)
}

func writeModule(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "MODULE.bazel"), []byte(content), 0o644))
	return root
}

func TestBazelFindingsCollector_RulesRsConsumer_RunsTheVariantAspect(t *testing.T) {
	runner := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{}}
	c := collector.NewBazelFindingsCollector(runner, &fakeFindingsParser{})
	workspace := writeModule(t, `bazel_dep(name = "rules_rs", version = "0.2.4")`+"\n")

	_, _, _, _, err := c.CollectFindings(context.Background(), workspace, []string{"//pkg:lib"}, map[string][]string{"rust": {"clippy"}}, nil)

	require.NoError(t, err)
	require.Len(t, runner.config.Aspects, 1)
	assert.Equal(t, "gavel_tools_rust_rs", runner.config.Aspects[0].Module,
		"judge must resolve the same variant init wired")
}

func TestBazelFindingsCollector_ConsumerWithoutRulesRs_RunsTheBaseAspect(t *testing.T) {
	runner := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{}}
	c := collector.NewBazelFindingsCollector(runner, &fakeFindingsParser{})
	workspace := writeModule(t, `bazel_dep(name = "rules_rust", version = "0.60.0")`+"\n")

	_, _, _, _, err := c.CollectFindings(context.Background(), workspace, []string{"//pkg:lib"}, map[string][]string{"rust": {"clippy"}}, nil)

	require.NoError(t, err)
	require.Len(t, runner.config.Aspects, 1)
	assert.Equal(t, "gavel_tools_rust", runner.config.Aspects[0].Module)
}
