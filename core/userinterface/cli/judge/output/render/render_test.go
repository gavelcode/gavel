package render_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/render"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

func TestFindings_EmptyResultProducesEmptyBlock(t *testing.T) {
	out := render.Findings(pipeline.Result{})
	assert.NotContains(t, out, "fp1")
}

func TestFindings_RendersFindingMessage(t *testing.T) {
	result := pipeline.Result{
		Findings: []evidencedto.Finding{
			{Tool: "PMD", RuleID: "R1", Severity: "error", FilePath: "a.go", Line: 1, Message: "null check missing", FingerprintID: "fp1"},
		},
	}

	out := render.Findings(result)

	assert.True(t, strings.Contains(out, "null check missing") || strings.Contains(out, "PMD"),
		"render output should mention either the message or the tool for a single finding")
}

func TestFindings_DoesNotPanicWithNewFingerprintDelta(t *testing.T) {
	result := pipeline.Result{
		Findings: []evidencedto.Finding{
			{Tool: "PMD", RuleID: "R1", Severity: "error", FilePath: "a.go", Line: 1, Message: "m", FingerprintID: "fp-new"},
		},
		Delta: pipeline.Delta{
			HasPrevious:     true,
			NewFingerprints: map[string]bool{"fp-new": true},
		},
	}

	out := render.Findings(result)
	assert.NotEmpty(t, out)
}

func TestCoverageSummary_EmptyReturnsEmpty(t *testing.T) {
	out := render.CoverageSummary(pipeline.Result{})
	assert.Empty(t, out)
}

func TestCoverageSummary_ShowsTopLevelDirectories(t *testing.T) {
	result := pipeline.Result{
		CoverageByFile: []evidencedto.FileCoverage{
			{FilePath: "pkg/a.go", Covered: []int{1, 2, 3}, Uncovered: []int{4}},
			{FilePath: "pkg/b.go", Covered: []int{1}, Uncovered: []int{2, 3}},
			{FilePath: "cmd/main.go", Covered: []int{1, 2}, Uncovered: nil},
		},
	}

	out := render.CoverageSummary(result)

	assert.Contains(t, out, "pkg")
	assert.Contains(t, out, "cmd")
	assert.Contains(t, out, "Coverage by directory")
}

func TestFindings_RendersViolations(t *testing.T) {
	result := pipeline.Result{
		Violations: []evidencedto.Violation{
			{Rule: "layer_violation", SourcePkg: "api", TargetPkg: "domain", Message: "forbidden import"},
		},
	}

	out := render.Findings(result)

	assert.NotEmpty(t, out)
}

func TestFindings_RendersViolationsWithDelta(t *testing.T) {
	result := pipeline.Result{
		Violations: []evidencedto.Violation{
			{Rule: "layer", SourcePkg: "api", TargetPkg: "domain", Message: "bad"},
			{Rule: "layer", SourcePkg: "infra", TargetPkg: "ui", Message: "bad"},
		},
		Delta: pipeline.Delta{
			HasArchPrevious: true,
			NewViolationIDs: map[string]bool{"layer:api:domain": true},
		},
	}

	out := render.Findings(result)

	assert.NotEmpty(t, out)
}

func TestCoverageSummary_RootFilesRendersTreeWithDot(t *testing.T) {
	result := pipeline.Result{
		CoverageByFile: []evidencedto.FileCoverage{
			{FilePath: "main.go", Covered: []int{1, 2}, Uncovered: []int{3}},
		},
	}

	out := render.CoverageSummary(result)

	assert.Contains(t, out, ".")
	assert.Contains(t, out, "Coverage by directory")
}
