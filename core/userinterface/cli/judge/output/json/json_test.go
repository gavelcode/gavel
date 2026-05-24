package json_test

import (
	"bytes"
	encjson "encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	corejudge "github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/json"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
)

func TestWrite_EmptyResults(t *testing.T) {
	var buf bytes.Buffer

	err := json.Write(&buf, nil)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	projects, ok := got["projects"].([]any)
	require.True(t, ok)
	assert.Empty(t, projects)
}

func TestWrite_WithResults(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{Name: "backend", Verdict: "pass", FindingsCount: 0, CoveragePercent: 85.5},
		{Name: "frontend", Verdict: "fail", FindingsCount: 3, CoveragePercent: 40.0},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	projects := got["projects"].([]any)
	assert.Len(t, projects, 2)

	first := projects[0].(map[string]any)
	assert.Equal(t, "backend", first["name"])
	assert.Equal(t, "pass", first["verdict"])
	assert.Equal(t, float64(0), first["findings_count"])
	assert.Equal(t, 85.5, first["coverage_percent"])
}

func TestWrite_IncludesCoverageByFile(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:            "backend",
			Verdict:         "fail",
			CoveragePercent: 50.0,
			CoverageByFile: []evidencedto.FileCoverage{
				{FilePath: "main.go", Covered: []int{1, 2}, Uncovered: []int{3, 4}},
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	first := got["projects"].([]any)[0].(map[string]any)
	files := first["coverage_by_file"].([]any)
	require.Len(t, files, 1)
	file := files[0].(map[string]any)
	assert.Equal(t, "main.go", file["file_path"])
	assert.Equal(t, float64(2), file["covered_lines"])
	assert.Equal(t, float64(4), file["total_lines"])
	assert.Equal(t, 50.0, file["percent"])
	assert.Equal(t, []any{float64(1), float64(2)}, file["covered"])
	assert.Equal(t, []any{float64(3), float64(4)}, file["uncovered"])
}

func TestWrite_OmitsCoverageByFileWhenSkipped(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:            "backend",
			Verdict:         "pass",
			CoverageSkipped: true,
			CoverageByFile: []evidencedto.FileCoverage{
				{FilePath: "main.go", Covered: []int{1}, Uncovered: []int{2}},
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	first := got["projects"].([]any)[0].(map[string]any)
	_, present := first["coverage_by_file"]
	assert.False(t, present)
}

func TestWrite_IncludesRulings(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:            "backend",
			Verdict:         "fail",
			FindingsCount:   3,
			CoveragePercent: 65.0,
			Rulings: []corejudge.RulingView{
				{Subtype: "code_quality", Passed: false, Detail: "3 exceed threshold"},
				{Subtype: "coverage", Passed: false, Detail: "65.0% < 80.0%"},
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)
	rulings := project["rulings"].([]any)
	assert.Len(t, rulings, 2)

	first := rulings[0].(map[string]any)
	assert.Equal(t, "code_quality", first["subtype"])
	assert.Equal(t, false, first["passed"])
	assert.Equal(t, "3 exceed threshold", first["detail"])
}

func TestWrite_IncludesFindings(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:          "backend",
			Verdict:       "fail",
			FindingsCount: 1,
			Findings: []evidencedto.Finding{
				{Tool: "PMD", RuleID: "NullCheck", Severity: "error", FilePath: "api.go", Line: 42, Message: "null check", FingerprintID: "fp1"},
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)
	findings := project["findings"].([]any)
	assert.Len(t, findings, 1)

	first := findings[0].(map[string]any)
	assert.Equal(t, "PMD", first["tool"])
	assert.Equal(t, "NullCheck", first["rule_id"])
	assert.Equal(t, "error", first["severity"])
	assert.Equal(t, "api.go", first["file_path"])
	assert.Equal(t, float64(42), first["line"])
}

func TestWrite_IncludesViolations(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:            "backend",
			Verdict:         "fail",
			ViolationsCount: 1,
			Violations: []evidencedto.Violation{
				{Rule: "layer_violation", SourcePkg: "api", TargetPkg: "domain", Message: "forbidden"},
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)
	violations := project["violations"].([]any)
	assert.Len(t, violations, 1)

	first := violations[0].(map[string]any)
	assert.Equal(t, "layer_violation", first["rule"])
	assert.Equal(t, "api", first["source_pkg"])
	assert.Equal(t, "domain", first["target_pkg"])
}

func TestWrite_IncludesDelta(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:          "backend",
			Verdict:       "fail",
			FindingsCount: 42,
			Findings: []evidencedto.Finding{
				{Tool: "PMD", RuleID: "R1", Severity: "error", FilePath: "a.go", Line: 1, Message: "m", FingerprintID: "fp-new"},
				{Tool: "PMD", RuleID: "R2", Severity: "warning", FilePath: "b.go", Line: 2, Message: "m", FingerprintID: "fp-old"},
			},
			Delta: pipeline.Delta{
				HasPrevious:     true,
				NewCount:        3,
				FixedCount:      8,
				ExistingCount:   31,
				FindingsDelta:   -5,
				CoverageDelta:   5.0,
				NewFingerprints: map[string]bool{"fp-new": true},
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)

	delta := project["delta"].(map[string]any)
	assert.Equal(t, true, delta["has_previous"])
	assert.Equal(t, float64(3), delta["new_count"])
	assert.Equal(t, float64(8), delta["fixed_count"])
	assert.Equal(t, float64(31), delta["existing_count"])
	assert.Equal(t, float64(-5), delta["findings_delta"])
	assert.Equal(t, 5.0, delta["coverage_delta"])

	findings := project["findings"].([]any)
	first := findings[0].(map[string]any)
	assert.Equal(t, "new", first["status"])
	second := findings[1].(map[string]any)
	assert.Equal(t, "existing", second["status"])
}

func TestWrite_IncludesCoverageTree(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:            "backend",
			Verdict:         "pass",
			CoveragePercent: 75.0,
			CoverageByFile: []evidencedto.FileCoverage{
				{FilePath: "pkg/a.go", Covered: []int{1, 2, 3}, Uncovered: []int{4}},
				{FilePath: "pkg/b.go", Covered: []int{1}, Uncovered: []int{2, 3}},
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)

	ct, ok := project["coverage_tree"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(4), ct["covered_lines"])
	assert.Equal(t, float64(7), ct["total_lines"])

	children := ct["children"].([]any)
	require.Len(t, children, 1)
	pkg := children[0].(map[string]any)
	assert.Equal(t, "pkg", pkg["path"])
	assert.Equal(t, float64(4), pkg["covered_lines"])
}

func TestWrite_IncludesFindingsTree(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:          "backend",
			Verdict:       "fail",
			FindingsCount: 2,
			Findings: []evidencedto.Finding{
				{Tool: "PMD", RuleID: "R1", Severity: "error", FilePath: "pkg/a.go", Line: 1, Message: "m", FingerprintID: "fp1"},
				{Tool: "PMD", RuleID: "R2", Severity: "warning", FilePath: "pkg/a.go", Line: 2, Message: "m", FingerprintID: "fp2"},
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)

	ft, ok := project["findings_tree"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(2), ft["count"])

	bySev := ft["by_severity"].(map[string]any)
	assert.Equal(t, float64(1), bySev["error"])
	assert.Equal(t, float64(1), bySev["warning"])
}

func TestWrite_OmitsTreesWhenCoverageSkipped(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:            "backend",
			Verdict:         "pass",
			CoverageSkipped: true,
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)
	_, hasCovTree := project["coverage_tree"]
	assert.False(t, hasCovTree)
}

func TestWrite_IncludesGitMetadata(t *testing.T) {
	var buf bytes.Buffer
	startedAt := time.Date(2025, 6, 20, 10, 0, 0, 0, time.UTC)
	results := []pipeline.Result{
		{
			Name:      "backend",
			Verdict:   "pass",
			CommitSHA: "abc123def456",
			Branch:    "main",
			StartedAt: startedAt,
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)
	assert.Equal(t, "abc123def456", project["commit_sha"])
	assert.Equal(t, "main", project["branch"])
	assert.Equal(t, "2025-06-20T10:00:00Z", project["started_at"])
}

func TestWrite_NoDeltaOnFirstRun(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{Name: "backend", Verdict: "pass", Delta: pipeline.Delta{HasPrevious: false}},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)
	_, hasDelta := project["delta"]
	assert.False(t, hasDelta)
}

func TestWrite_IncludesViolationStatusWithArchDelta(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:            "backend",
			Verdict:         "fail",
			ViolationsCount: 2,
			Violations: []evidencedto.Violation{
				{Rule: "layer", SourcePkg: "api", TargetPkg: "domain", Message: "forbidden"},
				{Rule: "layer", SourcePkg: "infra", TargetPkg: "ui", Message: "bad"},
			},
			Delta: pipeline.Delta{
				HasArchPrevious: true,
				NewViolationIDs: map[string]bool{"layer:api:domain": true},
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)
	violations := project["violations"].([]any)
	require.Len(t, violations, 2)
	assert.Equal(t, "new", violations[0].(map[string]any)["status"])
	assert.Equal(t, "existing", violations[1].(map[string]any)["status"])
}

func TestWrite_IncludesBuildWarning(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{Name: "backend", Verdict: "pass", BuildWarning: "partial failure"},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)
	assert.Equal(t, "partial failure", project["build_warning"])
}

func TestWrite_ArchDeltaOnly(t *testing.T) {
	var buf bytes.Buffer
	results := []pipeline.Result{
		{
			Name:    "backend",
			Verdict: "pass",
			Delta: pipeline.Delta{
				HasArchPrevious:         true,
				NewViolationsCount:      1,
				FixedViolationsCount:    2,
				ExistingViolationsCount: 3,
			},
		},
	}

	err := json.Write(&buf, results)

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &got))
	project := got["projects"].([]any)[0].(map[string]any)
	delta := project["delta"].(map[string]any)
	assert.Equal(t, float64(1), delta["new_violations_count"])
	assert.Equal(t, float64(2), delta["fixed_violations_count"])
	assert.Equal(t, float64(3), delta["existing_violations_count"])
}
