package ui_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/cli/ui"
)

func TestHeaderContainsCommand(t *testing.T) {
	got := ui.Header("JUDGE")
	assert.Contains(t, got, "JUDGE")
}

func TestPhaseHeaderContainsNumberAndName(t *testing.T) {
	got := ui.PhaseHeader(1, 3, "CONFIG", "Writing configuration")
	assert.Contains(t, got, "1/3")
	assert.Contains(t, got, "CONFIG")
	assert.Contains(t, got, "Writing configuration")
}

func TestPhaseItemShowsStatusOK(t *testing.T) {
	got := ui.PhaseItem("gavel.bazelrc", "CREATED", true)
	assert.Contains(t, got, "gavel.bazelrc")
	assert.Contains(t, got, "CREATED")
}

func TestPhaseItemShowsStatusFail(t *testing.T) {
	got := ui.PhaseItem("config", "FAILED", false)
	assert.Contains(t, got, "config")
	assert.Contains(t, got, "FAILED")
}

func TestJudgeVerdictPass(t *testing.T) {
	got := ui.JudgeVerdict("pass", "/tmp/case.json", 0, 0, 85.5, false, 90*time.Second)
	assert.Contains(t, got, "PASS")
	assert.Contains(t, got, "85.5%")
	assert.Contains(t, got, "1m 30s")
}

func TestJudgeVerdictFailWithFindings(t *testing.T) {
	got := ui.JudgeVerdict("fail", "/tmp/case.json", 3, 0, 40.0, false, 5*time.Second)
	assert.Contains(t, got, "FAIL")
	assert.Contains(t, got, "3 findings")
	assert.NotContains(t, got, "violations")
	assert.Contains(t, got, "5s")
}

func TestJudgeVerdictFailWithViolations(t *testing.T) {
	got := ui.JudgeVerdict("fail", "/tmp/case.json", 2, 5, 60.0, false, 3723*time.Second)
	assert.Contains(t, got, "FAIL")
	assert.Contains(t, got, "2 findings")
	assert.Contains(t, got, "5 violations")
	assert.Contains(t, got, "1h 2m 3s")
}

func TestJudgeVerdictCoverageSkipped(t *testing.T) {
	got := ui.JudgeVerdict("pass", "/tmp/case.json", 0, 0, 0, true, 10*time.Second)
	assert.Contains(t, got, "PASS")
	assert.NotContains(t, got, "coverage")
}

func TestFirstRunHint(t *testing.T) {
	got := ui.FirstRunHint()
	assert.Contains(t, got, "first run")
	assert.Contains(t, got, "baseline saved")
	assert.Contains(t, got, "new findings only")
}

func TestExistingConfigShowsPath(t *testing.T) {
	got := ui.ExistingConfig(".gavel/gavel.yaml")
	assert.Contains(t, got, ".gavel/gavel.yaml")
	assert.Contains(t, got, "--force")
}

func TestVerdictShowsConfigPath(t *testing.T) {
	got := ui.Verdict(".gavel/gavel.yaml")
	assert.Contains(t, got, ".gavel/gavel.yaml")
	assert.Contains(t, got, "SO ORDERED")
}

func TestTreeItem(t *testing.T) {
	got := ui.TreeItem("some detail")
	assert.Contains(t, got, "some detail")
	assert.Contains(t, got, "├─")
}

func TestTreeLastItem(t *testing.T) {
	got := ui.TreeLastItem("final item")
	assert.Contains(t, got, "final item")
	assert.Contains(t, got, "└─")
}

func TestRulingLinePassed(t *testing.T) {
	got := ui.RulingLine("code_quality", true, "0 findings", false)
	assert.Contains(t, got, "code_quality")
	assert.Contains(t, got, "PASS")
	assert.Contains(t, got, "0 findings")
	assert.Contains(t, got, "├─")
}

func TestRulingLineFailed(t *testing.T) {
	got := ui.RulingLine("coverage", false, "65.0% < 80.0%", true)
	assert.Contains(t, got, "coverage")
	assert.Contains(t, got, "FAIL")
	assert.Contains(t, got, "65.0% < 80.0%")
	assert.Contains(t, got, "└─")
}

func TestRulingLineEmptyDetail(t *testing.T) {
	got := ui.RulingLine("architecture", true, "", false)
	assert.Contains(t, got, "architecture")
	assert.Contains(t, got, "PASS")
}

func TestSummaryTableMultipleProjects(t *testing.T) {
	projects := []ui.ProjectSummary{
		{Name: "core", Verdict: "fail", Findings: 792, NewFindings: 0, Coverage: 73.1, Violations: 13},
		{Name: "cli", Verdict: "pass", Findings: 667, NewFindings: 0, Coverage: 0.0, Violations: 0},
		{Name: "server", Verdict: "pass", Findings: 794, NewFindings: 0, Coverage: 0.0, Violations: 0},
	}

	got := ui.SummaryTable(projects, 90*time.Second)

	assert.Contains(t, got, "SUMMARY")
	assert.Contains(t, got, "core")
	assert.Contains(t, got, "cli")
	assert.Contains(t, got, "server")
	assert.Contains(t, got, "FAIL")
	assert.Contains(t, got, "PASS")
	assert.Contains(t, got, "73.1%")
	assert.Contains(t, got, "13 violations")
	assert.Contains(t, got, "1/3 FAILED")
	assert.Contains(t, got, "1m 30s")
}

func TestSummaryTableAllPass(t *testing.T) {
	projects := []ui.ProjectSummary{
		{Name: "core", Verdict: "pass", Findings: 10, NewFindings: 0, Coverage: 95.0, Violations: 0},
		{Name: "web", Verdict: "pass", Findings: 5, NewFindings: 0, Coverage: 88.0, Violations: 0},
	}

	got := ui.SummaryTable(projects, 30*time.Second)

	assert.Contains(t, got, "2/2 PASSED")
	assert.NotContains(t, got, "FAILED")
}

func TestSummaryTableLongProjectName(t *testing.T) {
	projects := []ui.ProjectSummary{
		{Name: "very-long-project-name", Verdict: "pass", Findings: 5, NewFindings: 0, Coverage: 90.0, Violations: 0},
		{Name: "short", Verdict: "pass", Findings: 3, NewFindings: 0, Coverage: 85.0, Violations: 0},
	}

	got := ui.SummaryTable(projects, 10*time.Second)

	assert.Contains(t, got, "very-long-project-name")
	assert.Contains(t, got, "2/2 PASSED")
}

func TestSummaryTableSingleProjectReturnsEmpty(t *testing.T) {
	projects := []ui.ProjectSummary{
		{Name: "core", Verdict: "pass", Findings: 10, NewFindings: 0, Coverage: 95.0, Violations: 0},
	}

	got := ui.SummaryTable(projects, 10*time.Second)

	assert.Empty(t, got)
}

func TestTreeLastItemWithStatusOK(t *testing.T) {
	got := ui.TreeLastItemWithStatus("final check", "DONE", true)
	assert.Contains(t, got, "final check")
	assert.Contains(t, got, "DONE")
	assert.Contains(t, got, "└─")
}

func TestTreeLastItemWithStatusFail(t *testing.T) {
	got := ui.TreeLastItemWithStatus("final check", "ERROR", false)
	assert.Contains(t, got, "final check")
	assert.Contains(t, got, "ERROR")
}

func TestPhaseHeaderWithElapsed(t *testing.T) {
	got := ui.PhaseHeaderWithElapsed(2, 3, "BUILD", "Running aspects", "5s")
	assert.Contains(t, got, "2/3")
	assert.Contains(t, got, "BUILD")
	assert.Contains(t, got, "Running aspects")
	assert.Contains(t, got, "5s")
}

func TestServerFallbackWarning(t *testing.T) {
	got := ui.ServerFallbackWarning()
	assert.Contains(t, got, "server unreachable")
	assert.Contains(t, got, "local pipeline")
}

func TestBuildWarning(t *testing.T) {
	got := ui.BuildWarning()
	assert.Contains(t, got, "bazel build had failures")
	assert.Contains(t, got, "unreliable")
}

func TestMissingTargetsWarning(t *testing.T) {
	got := ui.MissingTargetsWarning("web", []string{"eslint", "archtest"})
	assert.Contains(t, got, "web")
	assert.Contains(t, got, "eslint")
	assert.Contains(t, got, "archtest")
	assert.Contains(t, got, "zero findings")
}

func TestCoverageBlockMultipleItems(t *testing.T) {
	items := []ui.CoverageItem{
		{Path: "core/domain/", CoveredLines: 500, TotalLines: 500, Percent: 100.0},
		{Path: "core/infra/", CoveredLines: 80, TotalLines: 100, Percent: 80.0},
	}
	got := ui.CoverageBlock(items)
	assert.Contains(t, got, "core/domain/")
	assert.Contains(t, got, "100.0%")
	assert.Contains(t, got, "core/infra/")
	assert.Contains(t, got, "80.0%")
	assert.Contains(t, got, "├─")
	assert.Contains(t, got, "└─")
}

func TestCoverageBlockEmpty(t *testing.T) {
	got := ui.CoverageBlock(nil)
	assert.Empty(t, got)
}
