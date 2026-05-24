package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/cli/ui"
)

func TestFindingsBlockEmpty(t *testing.T) {
	got := ui.FindingsBlock(nil)
	assert.Equal(t, "", got)
}

func TestFindingsBlockSingleFile(t *testing.T) {
	findings := []ui.FindingItem{
		{FilePath: "server/api.go", Line: 42, Severity: "error", Message: "null check missing", Tool: "PMD", RuleID: "NullCheck"},
		{FilePath: "server/api.go", Line: 18, Severity: "warning", Message: "exposed field", Tool: "SpotBugs", RuleID: "EI_EXPOSE"},
	}

	got := ui.FindingsBlock(findings)

	assert.Contains(t, got, "server/api.go")
	assert.Contains(t, got, "null check missing")
	assert.Contains(t, got, "exposed field")
	assert.Contains(t, got, "PMD:NullCheck")
	assert.Contains(t, got, "SpotBugs:EI_EXPOSE")
}

func TestFindingsBlockMultipleFiles(t *testing.T) {
	findings := []ui.FindingItem{
		{FilePath: "z/last.go", Line: 1, Severity: "note", Message: "msg-z", Tool: "T", RuleID: "R"},
		{FilePath: "a/first.go", Line: 1, Severity: "note", Message: "msg-a", Tool: "T", RuleID: "R"},
	}

	got := ui.FindingsBlock(findings)

	aPos := indexOf(got, "a/first.go")
	zPos := indexOf(got, "z/last.go")
	assert.Greater(t, zPos, aPos)
}

func TestFindingsBlockSortedByLine(t *testing.T) {
	findings := []ui.FindingItem{
		{FilePath: "main.go", Line: 99, Severity: "error", Message: "late", Tool: "T", RuleID: "R"},
		{FilePath: "main.go", Line: 3, Severity: "warning", Message: "early", Tool: "T", RuleID: "R"},
	}

	got := ui.FindingsBlock(findings)

	earlyPos := indexOf(got, "early")
	latePos := indexOf(got, "late")
	assert.Greater(t, latePos, earlyPos)
}

func TestViolationsBlockEmpty(t *testing.T) {
	got := ui.ViolationsBlock(nil)
	assert.Equal(t, "", got)
}

func TestViolationsBlockSingle(t *testing.T) {
	violations := []ui.ViolationItem{
		{Rule: "LayerRule", SourcePkg: "server/api", TargetPkg: "server/auth", Message: "forbidden import"},
	}

	got := ui.ViolationsBlock(violations)

	assert.Contains(t, got, "arch/violations")
	assert.Contains(t, got, "LayerRule")
	assert.Contains(t, got, "server/api")
	assert.Contains(t, got, "server/auth")
	assert.Contains(t, got, "forbidden import")
}

func TestFindingsBlockNewBadge(t *testing.T) {
	findings := []ui.FindingItem{
		{FilePath: "main.go", Line: 10, Severity: "error", Message: "issue", Tool: "T", RuleID: "R", IsNew: true},
		{FilePath: "main.go", Line: 20, Severity: "warning", Message: "old issue", Tool: "T", RuleID: "R2", IsNew: false},
	}

	got := ui.FindingsBlock(findings)

	assert.Contains(t, got, "NEW")
	newPos := indexOf(got, "NEW")
	oldPos := indexOf(got, "old issue")
	assert.Greater(t, oldPos, newPos)
}

func TestToolSummaryEmpty(t *testing.T) {
	got := ui.ToolSummary(nil)
	assert.Equal(t, "", got)
}

func TestToolSummarySingleTool(t *testing.T) {
	findings := []ui.FindingItem{
		{FilePath: "a.go", Line: 1, Severity: "error", Message: "m", Tool: "PMD", RuleID: "R1"},
		{FilePath: "b.go", Line: 2, Severity: "warning", Message: "m", Tool: "PMD", RuleID: "R2"},
	}

	got := ui.ToolSummary(findings)

	assert.Contains(t, got, "PMD: 2")
}

func TestToolSummaryMultipleToolsSortedAlphabetically(t *testing.T) {
	findings := []ui.FindingItem{
		{FilePath: "a.go", Line: 1, Severity: "error", Message: "m", Tool: "SpotBugs", RuleID: "R"},
		{FilePath: "b.go", Line: 2, Severity: "error", Message: "m", Tool: "PMD", RuleID: "R"},
		{FilePath: "c.go", Line: 3, Severity: "error", Message: "m", Tool: "golangci-lint", RuleID: "R"},
	}

	got := ui.ToolSummary(findings)

	assert.Contains(t, got, "PMD: 1")
	assert.Contains(t, got, "SpotBugs: 1")
	assert.Contains(t, got, "golangci-lint: 1")

	pmdPos := indexOf(got, "PMD")
	spotPos := indexOf(got, "SpotBugs")
	goPos := indexOf(got, "golangci-lint")
	assert.Greater(t, spotPos, pmdPos)
	assert.Greater(t, goPos, spotPos)
}

func TestToolSummaryCountsAccurate(t *testing.T) {
	findings := []ui.FindingItem{
		{FilePath: "a.go", Line: 1, Severity: "error", Message: "m", Tool: "PMD", RuleID: "R1"},
		{FilePath: "b.go", Line: 2, Severity: "error", Message: "m", Tool: "PMD", RuleID: "R2"},
		{FilePath: "c.go", Line: 3, Severity: "error", Message: "m", Tool: "PMD", RuleID: "R3"},
		{FilePath: "d.go", Line: 4, Severity: "error", Message: "m", Tool: "SpotBugs", RuleID: "R1"},
	}

	got := ui.ToolSummary(findings)

	assert.Contains(t, got, "PMD: 3")
	assert.Contains(t, got, "SpotBugs: 1")
}

func TestDeltaSummary_NoPrevious(t *testing.T) {
	got := ui.DeltaSummary(42, 0, 80.0, 0, 0, 0, 0, false, 0, 0, 0, false)
	assert.Equal(t, "", got)
}

func TestDeltaSummary_Improvement(t *testing.T) {
	got := ui.DeltaSummary(42, -5, 80.0, 5.0, 3, 8, 31, true, 0, 0, 0, false)

	assert.Contains(t, got, "42 findings")
	assert.Contains(t, got, "↓5 since last run")
	assert.Contains(t, got, "3 new")
	assert.Contains(t, got, "8 fixed")
	assert.Contains(t, got, "31 existing")
	assert.Contains(t, got, "↑5.0%")
}

func TestDeltaSummary_Regression(t *testing.T) {
	got := ui.DeltaSummary(50, 8, 70.0, -10.0, 8, 0, 42, true, 0, 0, 0, false)

	assert.Contains(t, got, "↑8 since last run")
	assert.Contains(t, got, "↓10.0%")
}

func TestDeltaSummary_NoChange(t *testing.T) {
	got := ui.DeltaSummary(42, 0, 80.0, 0, 0, 0, 42, true, 0, 0, 0, false)

	assert.Contains(t, got, "no change")
	assert.Contains(t, got, "0 new")
	assert.Contains(t, got, "0 fixed")
	assert.Contains(t, got, "42 existing")
}

func TestDeltaSummary_WithArchViolations(t *testing.T) {
	got := ui.DeltaSummary(10, 0, 80.0, 0, 0, 0, 10, true, 2, 1, 3, true)

	assert.Contains(t, got, "5 violations")
	assert.Contains(t, got, "2 new")
	assert.Contains(t, got, "1 fixed")
	assert.Contains(t, got, "3 existing (arch)")
}

func TestDeltaSummary_ArchOnlyNoPrevious(t *testing.T) {
	got := ui.DeltaSummary(10, 0, 0, 0, 0, 0, 0, false, 1, 0, 2, true)

	assert.NotEqual(t, "", got)
	assert.Contains(t, got, "3 violations")
	assert.Contains(t, got, "1 new")
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
