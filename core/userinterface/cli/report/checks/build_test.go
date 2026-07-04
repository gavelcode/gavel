package checks_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	outputjson "github.com/usegavel/gavel/core/userinterface/cli/judge/output/json"
	"github.com/usegavel/gavel/core/userinterface/cli/report/checks"
)

func floatPtr(f float64) *float64 { return &f }

func TestConclusionIsFailureWhenAnyProjectFails(t *testing.T) {
	run := checks.Build([]outputjson.Verdict{
		{Name: "core", Verdict: "pass"},
		{Name: "web", Verdict: "fail"},
	}, checks.Options{})
	assert.Equal(t, checks.ConclusionFailure, run.Conclusion)
}

func TestConclusionIsSuccessWhenEveryProjectPasses(t *testing.T) {
	run := checks.Build([]outputjson.Verdict{
		{Name: "core", Verdict: "pass"},
		{Name: "cli", Verdict: "pass"},
	}, checks.Options{})
	assert.Equal(t, checks.ConclusionSuccess, run.Conclusion)
}

func TestConclusionIsFailureForUnknownVerdict(t *testing.T) {
	run := checks.Build([]outputjson.Verdict{
		{Name: "core", Verdict: "pass"},
		{Name: "weird", Verdict: "error"},
	}, checks.Options{})
	assert.Equal(t, checks.ConclusionFailure, run.Conclusion,
		"a non-pass verdict must fail the check, not silently go green")
}

func TestSeverityMapsToAnnotationLevel(t *testing.T) {
	v := outputjson.Verdict{Name: "core", Verdict: "fail", Findings: []outputjson.VerdictFinding{
		{Severity: "error", FilePath: "a.go", Line: 1, Status: "new"},
		{Severity: "WARNING", FilePath: "b.go", Line: 2, Status: "new"},
		{Severity: "note", FilePath: "c.go", Line: 3, Status: "new"},
	}}
	run := checks.Build([]outputjson.Verdict{v}, checks.Options{})
	require.Len(t, run.Annotations, 3)
	assert.Equal(t, checks.LevelFailure, run.Annotations[0].Level)
	assert.Equal(t, checks.LevelWarning, run.Annotations[1].Level)
	assert.Equal(t, checks.LevelNotice, run.Annotations[2].Level)
}

func TestNewOnlySkipsExistingFindings(t *testing.T) {
	v := outputjson.Verdict{Name: "core", Verdict: "fail", Findings: []outputjson.VerdictFinding{
		{Severity: "error", FilePath: "a.go", Line: 1, Status: "new"},
		{Severity: "error", FilePath: "b.go", Line: 2, Status: "existing"},
	}}
	run := checks.Build([]outputjson.Verdict{v}, checks.Options{NewOnly: true})
	require.Len(t, run.Annotations, 1)
	assert.Equal(t, "a.go", run.Annotations[0].Path)
}

func TestNewOnlyKeepsUnknownStatusFindings(t *testing.T) {
	v := outputjson.Verdict{Name: "core", Verdict: "fail", Findings: []outputjson.VerdictFinding{
		{Severity: "error", FilePath: "a.go", Line: 1, Status: ""},
	}}
	run := checks.Build([]outputjson.Verdict{v}, checks.Options{NewOnly: true})
	assert.Len(t, run.Annotations, 1, "no-baseline findings must survive new-only")
}

func TestWithoutNewOnlyKeepsExistingFindings(t *testing.T) {
	v := outputjson.Verdict{Name: "core", Verdict: "fail", Findings: []outputjson.VerdictFinding{
		{Severity: "error", FilePath: "a.go", Line: 1, Status: "new"},
		{Severity: "error", FilePath: "b.go", Line: 2, Status: "existing"},
	}}
	run := checks.Build([]outputjson.Verdict{v}, checks.Options{NewOnly: false})
	assert.Len(t, run.Annotations, 2)
}

func TestAnnotationCarriesFileLineMessageAndTitle(t *testing.T) {
	v := outputjson.Verdict{Name: "core", Verdict: "fail", Findings: []outputjson.VerdictFinding{
		{Tool: "golangci-lint", RuleID: "errcheck", Severity: "error", FilePath: "core/x.go", Line: 10, Message: "unchecked error", Status: "new"},
	}}
	run := checks.Build([]outputjson.Verdict{v}, checks.Options{})
	require.Len(t, run.Annotations, 1)
	annotation := run.Annotations[0]
	assert.Equal(t, "core/x.go", annotation.Path)
	assert.Equal(t, 10, annotation.StartLine)
	assert.Equal(t, 10, annotation.EndLine)
	assert.Equal(t, "unchecked error", annotation.Message)
	assert.Contains(t, annotation.Title, "golangci-lint")
	assert.Contains(t, annotation.Title, "errcheck")
}

func TestZeroLineIsClampedToGitHubMinimum(t *testing.T) {
	v := outputjson.Verdict{Name: "core", Verdict: "fail", Findings: []outputjson.VerdictFinding{
		{Severity: "error", FilePath: "core/x.go", Line: 0, Status: "new"},
	}}
	run := checks.Build([]outputjson.Verdict{v}, checks.Options{})
	require.Len(t, run.Annotations, 1)
	assert.Equal(t, 1, run.Annotations[0].StartLine)
	assert.Equal(t, 1, run.Annotations[0].EndLine)
}

func TestHeadSHADefaultsToFirstVerdictButOptionWins(t *testing.T) {
	verdicts := []outputjson.Verdict{{Name: "core", Verdict: "pass", CommitSHA: "abc123"}}
	assert.Equal(t, "abc123", checks.Build(verdicts, checks.Options{}).HeadSHA)
	assert.Equal(t, "override", checks.Build(verdicts, checks.Options{HeadSHA: "override"}).HeadSHA)
}

func TestCheckNameDefaultsToGavel(t *testing.T) {
	assert.Equal(t, "gavel", checks.Build(nil, checks.Options{}).Name)
	assert.Equal(t, "ci/gavel", checks.Build(nil, checks.Options{CheckName: "ci/gavel"}).Name)
}

func TestSummaryListsProjectsCoverageAndViolations(t *testing.T) {
	verdicts := []outputjson.Verdict{
		{Name: "core", Verdict: "pass", CoveragePercent: floatPtr(94.6), Delta: &outputjson.VerdictDelta{NewCount: 0, FixedCount: 2}},
		{Name: "web", Verdict: "fail", Delta: &outputjson.VerdictDelta{NewCount: 3},
			Violations: []outputjson.VerdictViolation{{Rule: "deny", SourcePkg: "web/a", TargetPkg: "web/b", Message: "forbidden import"}}},
	}
	run := checks.Build(verdicts, checks.Options{})
	assert.Contains(t, run.Summary, "core")
	assert.Contains(t, run.Summary, "web")
	assert.Contains(t, run.Summary, "94.6%")
	assert.Contains(t, run.Summary, "Architecture violations")
	assert.Contains(t, run.Summary, "forbidden import")
}

func TestBatchAnnotationsChunksByMax(t *testing.T) {
	assert.Nil(t, checks.BatchAnnotations(nil, checks.MaxAnnotationsPerRequest))

	fifty := make([]checks.Annotation, 50)
	batches := checks.BatchAnnotations(fifty, checks.MaxAnnotationsPerRequest)
	require.Len(t, batches, 1)
	assert.Len(t, batches[0], 50)

	fiftyOne := make([]checks.Annotation, 51)
	batches = checks.BatchAnnotations(fiftyOne, checks.MaxAnnotationsPerRequest)
	require.Len(t, batches, 2)
	assert.Len(t, batches[0], 50)
	assert.Len(t, batches[1], 1)
}

func TestMaxAnnotationsPerRequestMatchesGitHubLimit(t *testing.T) {
	assert.Equal(t, 50, checks.MaxAnnotationsPerRequest)
}
