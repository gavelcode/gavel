package collector_test

import (
	"context"
	"fmt"
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

	evidences, rawFiles, buildWarning, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}})

	require.NoError(t, err)
	assert.Empty(t, evidences)
	assert.Empty(t, rawFiles)
	assert.Empty(t, buildWarning)
}

func TestBazelFindingsCollector_ParsesReports(t *testing.T) {
	sarif := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`)
	runner := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{"go_golangci_lint_submission_aspect": {sarif}}}
	parser := &fakeFindingsParser{returnEmpty: true}
	c := collector.NewBazelFindingsCollector(runner, parser)

	evidences, rawFiles, buildWarning, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}})

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

	evidences, rawFiles, buildWarning, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{})

	require.NoError(t, err)
	assert.Nil(t, evidences)
	assert.Nil(t, rawFiles)
	assert.Empty(t, buildWarning)
}

func TestBazelFindingsCollector_RunnerError(t *testing.T) {
	r := &fakeAnalysisRunner{err: fmt.Errorf("bazel build failed")}
	parser := &fakeFindingsParser{}
	c := collector.NewBazelFindingsCollector(r, parser)

	_, _, _, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "run analysis")
}

func TestBazelFindingsCollector_ParserError(t *testing.T) {
	sarif := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`)
	r := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{"go_golangci_lint_submission_aspect": {sarif}}}
	parser := &fakeFindingsParser{err: fmt.Errorf("parse failed")}
	c := collector.NewBazelFindingsCollector(r, parser)

	_, _, _, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse failed")
}

func TestBazelFindingsCollector_EmptySARIFData(t *testing.T) {
	r := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{"go_golangci_lint_submission_aspect": {[]byte{}}}}
	parser := &fakeFindingsParser{}
	c := collector.NewBazelFindingsCollector(r, parser)

	_, _, _, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}})

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

	evidences, rawFiles, buildWarning, err := c.CollectFindings(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"golangci-lint"}})

	require.NoError(t, err)
	assert.Len(t, evidences, 1)
	assert.Len(t, rawFiles, 1)
	assert.Equal(t, "bazel build had failures", buildWarning)
}
