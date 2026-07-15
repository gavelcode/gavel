package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

func fakeCLI(t *testing.T, jsonOutput string) *executor.CLI {
	t.Helper()
	tmpDir := t.TempDir()
	script := filepath.Join(tmpDir, "fake-gavel")
	content := "#!/bin/sh\ncat <<'ENDJSON'\n" + jsonOutput + "\nENDJSON\n"
	require.NoError(t, os.WriteFile(script, []byte(content), 0o755))
	return executor.NewWithBinary(script, tmpDir)
}

func fakeCLIWithExitCode(t *testing.T, jsonOutput string, exitCode int) *executor.CLI {
	t.Helper()
	tmpDir := t.TempDir()
	script := filepath.Join(tmpDir, "fake-gavel")
	content := "#!/bin/sh\ncat <<'ENDJSON'\n" + jsonOutput + "\nENDJSON\nexit " + fmt.Sprintf("%d", exitCode) + "\n"
	require.NoError(t, os.WriteFile(script, []byte(content), 0o755))
	return executor.NewWithBinary(script, tmpDir)
}

func TestRunJudge_PassingProject(t *testing.T) {
	judgeJSON := `{"projects":[{"name":"core","verdict":"pass","findings_count":0,"violations_count":0,"coverage_percent":85.3,"rulings":[{"subtype":"findings","passed":true,"detail":"0 findings"}]}]}`
	cli := fakeCLI(t, judgeJSON)

	result, err := runJudge(context.Background(), cli, JudgeInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "core — PASS")
	assert.Contains(t, result, "Coverage: 85.3%")
}

func TestRunJudge_WithProjectFilter(t *testing.T) {
	judgeJSON := `{"projects":[{"name":"cli","verdict":"pass","findings_count":0,"violations_count":0,"rulings":[]}]}`
	cli := fakeCLI(t, judgeJSON)

	result, err := runJudge(context.Background(), cli, JudgeInput{Project: "cli"})

	require.NoError(t, err)
	assert.Contains(t, result, "cli — PASS")
}

func TestRunJudge_FailingVerdict(t *testing.T) {
	judgeJSON := `{"projects":[{"name":"core","verdict":"fail","findings_count":5,"violations_count":0,"rulings":[{"subtype":"findings","passed":false,"detail":"5 > 0"}]}]}`
	cli := fakeCLIWithExitCode(t, judgeJSON, 1)

	result, err := runJudge(context.Background(), cli, JudgeInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "FAIL")
	assert.Contains(t, result, "Quality gate FAILED")
}

func TestRunCoverage_ShowsPercentage(t *testing.T) {
	coverageJSON := `{"projects":[{"name":"core","verdict":"pass","coverage_percent":91.5,"rulings":[{"subtype":"coverage","passed":true,"detail":"91.5% coverage (min 90.0%)"}],"coverage_by_file":[{"file_path":"main.go","covered_lines":10,"total_lines":11,"percent":90.9}]}]}`
	cli := fakeCLI(t, coverageJSON)

	result, err := runCoverage(context.Background(), cli, CoverageInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "91.5%")
	assert.Contains(t, result, "main.go")
}

func TestRunCoverage_WithFileFilter(t *testing.T) {
	coverageJSON := `{"projects":[{"name":"core","verdict":"pass","coverage_percent":80.0,"rulings":[],"coverage_by_file":[{"file_path":"a.go","covered_lines":5,"total_lines":10,"percent":50.0,"covered":[1,2,3,4,5],"uncovered":[6,7,8,9,10]},{"file_path":"b.go","covered_lines":10,"total_lines":10,"percent":100.0}]}]}`
	cli := fakeCLI(t, coverageJSON)

	result, err := runCoverage(context.Background(), cli, CoverageInput{Files: []string{"a.go"}})

	require.NoError(t, err)
	assert.Contains(t, result, "a.go")
	assert.Contains(t, result, "uncovered:")
	assert.NotContains(t, result, "b.go")
}

func TestRunCoverage_WithProjectFilter(t *testing.T) {
	coverageJSON := `{"projects":[{"name":"web","verdict":"pass","coverage_percent":60.0,"rulings":[]}]}`
	cli := fakeCLI(t, coverageJSON)

	result, err := runCoverage(context.Background(), cli, CoverageInput{Project: "web"})

	require.NoError(t, err)
	assert.Contains(t, result, "web")
	assert.Contains(t, result, "60.0%")
}

func TestRunLintFile_WithFindings(t *testing.T) {
	lintJSON := `{"projects":[{"name":"core","findings":[{"tool":"golangci-lint","rule_id":"unused","severity":"warning","file_path":"main.go","line":10,"message":"unused var","fingerprint":"fp1"},{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"other.go","line":5,"message":"unchecked","fingerprint":"fp2"}]}]}`
	cli := fakeCLI(t, lintJSON)

	result, err := runLintFile(context.Background(), cli, LintFileInput{File: "main.go"})

	require.NoError(t, err)
	assert.Contains(t, result, "main.go")
	assert.Contains(t, result, "unused var")
	assert.NotContains(t, result, "other.go")
}

func TestRunLintFile_NoFindings(t *testing.T) {
	lintJSON := `{"projects":[{"name":"core","findings":[{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"other.go","line":5,"message":"unchecked","fingerprint":"fp2"}]}]}`
	cli := fakeCLI(t, lintJSON)

	result, err := runLintFile(context.Background(), cli, LintFileInput{File: "clean.go"})

	require.NoError(t, err)
	assert.Contains(t, result, "No findings in clean.go")
}

func TestRunLintFile_NewFindingsMarked(t *testing.T) {
	lintJSON := `{"projects":[{"name":"core","findings":[{"tool":"golangci-lint","rule_id":"govet","severity":"warning","file_path":"main.go","line":20,"message":"shadow","fingerprint":"fp3","status":"new"}]}]}`
	cli := fakeCLI(t, lintJSON)

	result, err := runLintFile(context.Background(), cli, LintFileInput{File: "main.go"})

	require.NoError(t, err)
	assert.Contains(t, result, "NEW")
}

func TestRunTrends_WithData(t *testing.T) {
	trendsJSON := `[{"commit_sha":"abc1234567","branch":"main","coverage_percent":92.0,"total_findings":10,"new_findings":0,"resolved_findings":2,"verdict_outcome":"pass","created_at":"2026-06-14T10:00:00Z"}]`
	cli := fakeCLI(t, trendsJSON)

	result, err := runTrends(context.Background(), cli, TrendsInput{Project: "core"})

	require.NoError(t, err)
	assert.Contains(t, result, "core")
	assert.Contains(t, result, "abc1234")
	assert.Contains(t, result, "92.0%")
}

func TestRunTrends_EmptyHistory(t *testing.T) {
	cli := fakeCLI(t, `[]`)

	result, err := runTrends(context.Background(), cli, TrendsInput{Project: "core"})

	require.NoError(t, err)
	assert.Contains(t, result, "No analysis history")
}

func TestRunTrends_WithBranchFilter(t *testing.T) {
	trendsJSON := `[{"commit_sha":"abc","branch":"feat","total_findings":5,"new_findings":1,"resolved_findings":0,"verdict_outcome":"pass","created_at":"2026-06-14T10:00:00Z"}]`
	cli := fakeCLI(t, trendsJSON)

	result, err := runTrends(context.Background(), cli, TrendsInput{Project: "core", Branch: "feat"})

	require.NoError(t, err)
	assert.Contains(t, result, "core")
}

func TestTrendArrow_PositiveDelta(t *testing.T) {
	assert.Equal(t, "▲", trendArrow(5.0))
}

func TestTrendArrow_NegativeDelta(t *testing.T) {
	assert.Equal(t, "▼", trendArrow(-3.0))
}

func TestTrendArrow_NearZero(t *testing.T) {
	assert.Equal(t, "→", trendArrow(0.005))
	assert.Equal(t, "→", trendArrow(-0.005))
}

func TestTrendTimeAgo_JustNow(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	result := trendTimeAgo(now)
	assert.Contains(t, result, "just now")
}

func TestTrendTimeAgo_HoursAgo(t *testing.T) {
	twoHoursAgo := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	result := trendTimeAgo(twoHoursAgo)
	assert.Contains(t, result, "2h ago")
}

func TestTrendTimeAgo_DaysAgo(t *testing.T) {
	threeDaysAgo := time.Now().Add(-72 * time.Hour).Format(time.RFC3339)
	result := trendTimeAgo(threeDaysAgo)
	assert.Contains(t, result, "3d ago")
}

func TestTrendTimeAgo_InvalidFormat(t *testing.T) {
	result := trendTimeAgo("not-a-date")
	assert.Equal(t, "not-a-date", result)
}

func TestCompressRanges_Empty(t *testing.T) {
	assert.Equal(t, "", compressRanges(nil))
}

func TestCompressRanges_SingleLine(t *testing.T) {
	assert.Equal(t, "5", compressRanges([]int{5}))
}

func TestCompressRanges_ConsecutiveRange(t *testing.T) {
	assert.Equal(t, "1-5", compressRanges([]int{1, 2, 3, 4, 5}))
}

func TestCompressRanges_MixedRangesAndSingles(t *testing.T) {
	assert.Equal(t, "1-3, 7, 10-12", compressRanges([]int{1, 2, 3, 7, 10, 11, 12}))
}

func TestFormatSeverities_AllPresent(t *testing.T) {
	result := formatSeverities(map[string]int{"error": 2, "warning": 3, "note": 1})
	assert.Equal(t, "2 error, 3 warning, 1 note", result)
}

func TestFormatSeverities_OnlyError(t *testing.T) {
	result := formatSeverities(map[string]int{"error": 5})
	assert.Equal(t, "5 error", result)
}

func TestFormatSeverities_Empty(t *testing.T) {
	result := formatSeverities(map[string]int{})
	assert.Equal(t, "", result)
}

func TestRunInit_Success(t *testing.T) {
	cli := fakeCLI(t, "Gavel initialized successfully.")

	result, err := runInit(context.Background(), cli, InitInput{From: "gavel.yaml"})

	require.NoError(t, err)
	assert.Contains(t, result, "initialized")
}

func TestRunInit_Failure(t *testing.T) {
	cli := fakeCLIWithExitCode(t, "missing gavel.yaml", 1)

	result, err := runInit(context.Background(), cli, InitInput{From: "missing.yaml"})

	require.NoError(t, err)
	assert.Contains(t, result, "failed")
	assert.Contains(t, result, "missing gavel.yaml")
}

func TestRunValidate_Valid(t *testing.T) {
	cli := fakeCLI(t, "all checks passed")

	result, err := runValidate(context.Background(), cli, ValidateInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "valid")
}

func TestRunValidate_Invalid(t *testing.T) {
	cli := fakeCLIWithExitCode(t, "missing .bazelrc include", 1)

	result, err := runValidate(context.Background(), cli, ValidateInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "issues")
	assert.Contains(t, result, "missing .bazelrc include")
}

func TestFormatTrendsOutput_WithMultipleEntries(t *testing.T) {
	cov1 := 95.0
	cov2 := 90.0
	data := fmt.Sprintf(`[
		{"commit_sha":"aaabbb1234567","branch":"main","coverage_percent":%f,"total_findings":3,"new_findings":0,"resolved_findings":1,"verdict_outcome":"pass","created_at":"2026-06-15T10:00:00Z"},
		{"commit_sha":"cccdddd1234567","branch":"main","coverage_percent":%f,"total_findings":5,"new_findings":2,"resolved_findings":0,"verdict_outcome":"fail","created_at":"2026-06-14T08:00:00Z"}
	]`, cov1, cov2)

	result, err := formatTrendsOutput([]byte(data), "core")

	require.NoError(t, err)
	assert.Contains(t, result, "Trends — core")
	assert.Contains(t, result, "aaabbb1")
	assert.Contains(t, result, "cccdddd")
	assert.Contains(t, result, "Coverage:")
	assert.Contains(t, result, "Findings:")
	assert.Contains(t, result, "Pass rate: 1/2")
}

func failingCLI(t *testing.T) *executor.CLI {
	t.Helper()
	tmpDir := t.TempDir()
	script := filepath.Join(tmpDir, "failing-gavel")
	content := "#!/bin/sh\necho 'command failed' >&2\nexit 1\n"
	require.NoError(t, os.WriteFile(script, []byte(content), 0o755))
	return executor.NewWithBinary(script, tmpDir)
}

func TestRunJudge_WithAffectedAndQuick(t *testing.T) {
	judgeJSON := `{"projects":[{"name":"core","verdict":"pass","findings_count":0,"violations_count":0,"rulings":[]}]}`
	cli := fakeCLI(t, judgeJSON)

	result, err := runJudge(context.Background(), cli, JudgeInput{Affected: true, Quick: true})

	require.NoError(t, err)
	assert.Contains(t, result, "core — PASS")
}

func TestRunJudge_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := runJudge(context.Background(), cli, JudgeInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execute gavel judge")
}

func TestRunCoverage_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := runCoverage(context.Background(), cli, CoverageInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execute gavel judge")
}

func TestRunCoverage_ParseError(t *testing.T) {
	cli := fakeCLI(t, "not json")
	_, err := runCoverage(context.Background(), cli, CoverageInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse judge output")
}

func TestRunCoverage_NoCoveragePercent(t *testing.T) {
	coverageJSON := `{"projects":[{"name":"core","verdict":"pass","rulings":[]}]}`
	cli := fakeCLI(t, coverageJSON)

	result, err := runCoverage(context.Background(), cli, CoverageInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "not collected")
}

func TestRunLintFile_WithProjectFilter(t *testing.T) {
	lintJSON := `{"projects":[{"name":"web","findings":[{"tool":"eslint","rule_id":"no-unused-vars","severity":"warning","file_path":"app.ts","line":5,"message":"unused","fingerprint":"fp1"}]}]}`
	cli := fakeCLI(t, lintJSON)

	result, err := runLintFile(context.Background(), cli, LintFileInput{File: "app.ts", Project: "web"})

	require.NoError(t, err)
	assert.Contains(t, result, "app.ts")
}

func TestRunLintFile_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := runLintFile(context.Background(), cli, LintFileInput{File: "main.go"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execute gavel judge")
}

func TestRunLintFile_ParseError(t *testing.T) {
	cli := fakeCLI(t, "not json")
	_, err := runLintFile(context.Background(), cli, LintFileInput{File: "main.go"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse judge output")
}

func TestRunTrends_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := runTrends(context.Background(), cli, TrendsInput{Project: "core"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execute gavel trends")
}

func TestTrendTimeAgo_MinutesAgo(t *testing.T) {
	thirtyMinAgo := time.Now().Add(-30 * time.Minute).Format(time.RFC3339)
	result := trendTimeAgo(thirtyMinAgo)
	assert.Contains(t, result, "m ago")
}

func TestRunInit_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := runInit(context.Background(), cli, InitInput{From: "gavel.yaml"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execute gavel init")
}

func TestRunInit_WithGavelspace(t *testing.T) {
	cli := fakeCLI(t, "Initialized.")
	result, err := runInit(context.Background(), cli, InitInput{From: "gavel.yaml", Gavelspace: t.TempDir()})
	require.NoError(t, err)
	assert.Contains(t, result, "Initialized")
}

func TestRunValidate_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := runValidate(context.Background(), cli, ValidateInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execute gavel validate")
}

func TestRunValidate_WithGavelspace(t *testing.T) {
	cli := fakeCLI(t, "all checks passed")
	result, err := runValidate(context.Background(), cli, ValidateInput{Gavelspace: t.TempDir()})
	require.NoError(t, err)
	assert.Contains(t, result, "valid")
}

func TestRunJudge_ShowsViolationDetails(t *testing.T) {
	judgeJSON := `{"projects":[{"name":"core","verdict":"fail","findings_count":0,"violations_count":2,"rulings":[{"subtype":"architecture","passed":false,"detail":"2 violations"}],"violations":[{"rule":"no_infra_in_domain","source_pkg":"domain/foo","target_pkg":"infra/bar","message":"forbidden import"},{"rule":"no_ui_in_app","source_pkg":"application/baz","target_pkg":"userinterface/qux","message":"layer violation"}]}]}`
	cli := fakeCLIWithExitCode(t, judgeJSON, 1)

	result, err := runJudge(context.Background(), cli, JudgeInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "no_infra_in_domain: domain/foo → infra/bar")
	assert.Contains(t, result, "no_ui_in_app: application/baz → userinterface/qux")
}

func TestRunJudge_ShowsNewViolationMarker(t *testing.T) {
	judgeJSON := `{"projects":[{"name":"core","verdict":"fail","findings_count":0,"violations_count":1,"rulings":[{"subtype":"architecture","passed":false,"detail":"1 violation"}],"violations":[{"rule":"no_infra_in_domain","source_pkg":"domain/x","target_pkg":"infra/y","message":"bad import","status":"new"}]}]}`
	cli := fakeCLIWithExitCode(t, judgeJSON, 1)

	result, err := runJudge(context.Background(), cli, JudgeInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "NEW")
	assert.Contains(t, result, "no_infra_in_domain: domain/x → infra/y")
}

func TestRunArch_NoViolations(t *testing.T) {
	archJSON := `{"projects":[{"name":"core","verdict":"pass","findings_count":0,"violations_count":0,"rulings":[{"subtype":"architecture","passed":true,"detail":"0 violations"}]}]}`
	cli := fakeCLI(t, archJSON)

	result, err := runArch(context.Background(), cli, ArchInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "No architecture violations found")
}

func TestRunArch_WithViolations(t *testing.T) {
	archJSON := `{"projects":[{"name":"core","verdict":"fail","findings_count":0,"violations_count":2,"rulings":[{"subtype":"architecture","passed":false,"detail":"2 violations"}],"violations":[{"rule":"no_infra_in_domain","source_pkg":"domain/foo","target_pkg":"infra/bar","message":"forbidden import"},{"rule":"no_ui_in_app","source_pkg":"application/baz","target_pkg":"userinterface/qux","message":"layer violation","status":"new"}]}]}`
	cli := fakeCLIWithExitCode(t, archJSON, 1)

	result, err := runArch(context.Background(), cli, ArchInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "core")
	assert.Contains(t, result, "no_infra_in_domain: domain/foo → infra/bar")
	assert.Contains(t, result, "no_ui_in_app: application/baz → userinterface/qux")
	assert.Contains(t, result, "NEW")
}

func TestRunArch_WithProjectFilter(t *testing.T) {
	archJSON := `{"projects":[{"name":"server","verdict":"pass","findings_count":0,"violations_count":0,"rulings":[]}]}`
	cli := fakeCLI(t, archJSON)

	result, err := runArch(context.Background(), cli, ArchInput{Project: "server"})

	require.NoError(t, err)
	assert.Contains(t, result, "No architecture violations found")
}

func TestRunArch_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := runArch(context.Background(), cli, ArchInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execute gavel judge")
}

func TestFormatTrendsOutput_NoCoverage(t *testing.T) {
	data := `[{"commit_sha":"abc","branch":"main","total_findings":0,"new_findings":0,"resolved_findings":0,"verdict_outcome":"pass","created_at":"2026-06-15T10:00:00Z"}]`

	result, err := formatTrendsOutput([]byte(data), "core")

	require.NoError(t, err)
	assert.NotContains(t, result, "Coverage:")
	assert.Contains(t, result, "Findings:")
}

func fakeCLICapturingArgs(t *testing.T, jsonOutput string) (*executor.CLI, func() string) {
	t.Helper()
	tmpDir := t.TempDir()
	argsFile := filepath.Join(tmpDir, "args")
	script := filepath.Join(tmpDir, "fake-gavel")
	content := "#!/bin/sh\necho \"$@\" > " + argsFile + "\ncat <<'ENDJSON'\n" + jsonOutput + "\nENDJSON\n"
	require.NoError(t, os.WriteFile(script, []byte(content), 0o755))
	read := func() string {
		b, _ := os.ReadFile(argsFile)
		return string(b)
	}
	return executor.NewWithBinary(script, tmpDir), read
}

func TestRunFindings_PassesNoBaselineUpdate(t *testing.T) {
	cli, args := fakeCLICapturingArgs(t, `{"projects":[]}`)
	_, err := runFindings(context.Background(), cli, FindingsInput{})
	require.NoError(t, err)
	assert.Contains(t, args(), "--no-baseline-update")
}

func TestRunLintFile_PassesNoBaselineUpdate(t *testing.T) {
	cli, args := fakeCLICapturingArgs(t, `{"projects":[]}`)
	_, err := runLintFile(context.Background(), cli, LintFileInput{File: "main.go"})
	require.NoError(t, err)
	assert.Contains(t, args(), "--no-baseline-update")
}

func TestRunCoverage_PassesNoBaselineUpdate(t *testing.T) {
	cli, args := fakeCLICapturingArgs(t, `{"projects":[]}`)
	_, err := runCoverage(context.Background(), cli, CoverageInput{})
	require.NoError(t, err)
	assert.Contains(t, args(), "--no-baseline-update")
}

func TestRunArch_PassesNoBaselineUpdate(t *testing.T) {
	cli, args := fakeCLICapturingArgs(t, `{"projects":[]}`)
	_, err := runArch(context.Background(), cli, ArchInput{})
	require.NoError(t, err)
	assert.Contains(t, args(), "--no-baseline-update")
}
