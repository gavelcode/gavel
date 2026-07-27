package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

func TestToolsWithoutSubmissions_FlagsConfiguredAspectThatProducedNoSarif(t *testing.T) {
	result := &AnalysisResult{
		SARIFFiles: map[string][][]byte{
			"clippy": {[]byte("{}")},
		},
	}
	aspects := []catalog.Aspect{
		{Name: "clippy"},
		{Name: "eslint"},
	}

	got := ToolsWithoutSubmissions(result, aspects)

	assert.Equal(t, []string{"eslint"}, got,
		"an enabled aspect that emitted no SARIF attached to zero targets; surfacing it keeps a 0-findings verdict from silently implying the tool ran")
}

func TestToolsWithoutSubmissions_EmptyWhenEveryAspectProducedSarif(t *testing.T) {
	result := &AnalysisResult{
		SARIFFiles: map[string][][]byte{
			"clippy": {[]byte("{}")},
			"eslint": {[]byte("{}")},
		},
	}
	aspects := []catalog.Aspect{{Name: "clippy"}, {Name: "eslint"}}

	assert.Empty(t, ToolsWithoutSubmissions(result, aspects))
}

func TestBuildBazelArgs_TargetsAfterOptionsMarker(t *testing.T) {
	config := AnalysisConfig{
		Targets: []string{"//core/...", "-//core/gen/..."},
		Aspects: []catalog.Aspect{{Path: "p"}},
	}

	args, err := buildBazelArgs(config)
	require.NoError(t, err)

	marker := slices.Index(args, "--")
	require.GreaterOrEqual(t, marker, 0, "args must contain the -- options-end marker")
	assert.Greater(t, slices.Index(args, "//core/..."), marker, "positive target must follow --")
	assert.Greater(t, slices.Index(args, "-//core/gen/..."), marker, "negative target must follow -- so Bazel does not read it as an option")
	for i, a := range args {
		if i > marker && len(a) > 1 && a[0] == '-' && a[1] == '-' {
			t.Fatalf("flag %q must appear before the -- marker", a)
		}
	}
}

func TestBuildBazelArgs_IncludesAspectBuildFlagsBeforeMarker(t *testing.T) {
	config := AnalysisConfig{
		Targets: []string{"//core/..."},
		Aspects: []catalog.Aspect{
			{Path: "golangci", BuildFlags: []string{"--@rules_go//go/config:export_stdlib=True"}},
			{Path: "archtest"},
		},
	}

	args, err := buildBazelArgs(config)
	require.NoError(t, err)

	idx := slices.Index(args, "--@rules_go//go/config:export_stdlib=True")
	require.GreaterOrEqual(t, idx, 0, "golangci's build flag must reach the combined invocation")
	assert.Less(t, idx, slices.Index(args, "--"), "the flag must come before the -- marker")
}

func TestBuildBazelArgs_DeduplicatesSharedBuildFlags(t *testing.T) {
	flag := "--@rules_go//go/config:export_stdlib=True"
	config := AnalysisConfig{
		Targets: []string{"//core/..."},
		Aspects: []catalog.Aspect{
			{Path: "a", BuildFlags: []string{flag}},
			{Path: "b", BuildFlags: []string{flag}},
		},
	}

	args, err := buildBazelArgs(config)
	require.NoError(t, err)

	count := 0
	for _, arg := range args {
		if arg == flag {
			count++
		}
	}
	assert.Equal(t, 1, count, "a flag shared by two aspects must appear once")
}

func TestBuildBazelArgs_CoverageMode(t *testing.T) {
	config := AnalysisConfig{
		Targets:         []string{"//..."},
		Aspects:         []catalog.Aspect{{Path: "@gavel//a:defs.bzl%lint_a"}},
		IncludeCoverage: true,
	}

	args, err := buildBazelArgs(config)
	require.NoError(t, err)

	assert.Equal(t, "coverage", args[0])
	assert.Contains(t, args, "--test_size_filters=small,medium")
	for _, a := range args {
		assert.NotContains(t, a, "combined_report")
		assert.NotContains(t, a, "instrumentation_filter")
	}
}

func TestBuildBazelArgs_CoverageBoundsMemory(t *testing.T) {
	cov, err := buildBazelArgs(AnalysisConfig{Targets: []string{"//..."}, IncludeCoverage: true})
	require.NoError(t, err)
	assert.Contains(t, cov, "--local_resources=memory=HOST_RAM*0.67")

	build, err := buildBazelArgs(AnalysisConfig{Targets: []string{"//..."}, IncludeCoverage: false})
	require.NoError(t, err)
	for _, a := range build {
		assert.NotContains(t, a, "local_resources")
	}
}

func TestBuildBazelArgs_BuildMode(t *testing.T) {
	config := AnalysisConfig{
		Targets:         []string{"//..."},
		Aspects:         []catalog.Aspect{{Path: "@gavel//a:defs.bzl%lint_a"}},
		IncludeCoverage: false,
	}

	args, err := buildBazelArgs(config)
	require.NoError(t, err)

	assert.Equal(t, "build", args[0])
	for _, a := range args {
		assert.NotContains(t, a, "combined_report")
		assert.NotContains(t, a, "test_size_filters")
	}
}

func TestBuildBazelArgs_DefaultTestSizeFilters(t *testing.T) {
	config := AnalysisConfig{
		Targets:         []string{"//..."},
		Aspects:         []catalog.Aspect{{Path: "p"}},
		IncludeCoverage: true,
	}

	args, err := buildBazelArgs(config)
	require.NoError(t, err)

	assert.Contains(t, args, "--test_size_filters=small,medium")
}

func TestBuildBazelArgs_CustomTestSizeFilters(t *testing.T) {
	config := AnalysisConfig{
		Targets:         []string{"//..."},
		Aspects:         []catalog.Aspect{{Path: "p"}},
		IncludeCoverage: true,
		TestSizeFilters: "small",
	}

	args, err := buildBazelArgs(config)
	require.NoError(t, err)

	assert.Contains(t, args, "--test_size_filters=small")
	assert.NotContains(t, args, "--test_size_filters=small,medium")
}

func TestBuildBazelArgs_TestTagFilters(t *testing.T) {
	config := AnalysisConfig{
		Targets:        []string{"//..."},
		Aspects:        []catalog.Aspect{{Path: "p"}},
		TestTagFilters: "-integration,-manual",
	}

	args, err := buildBazelArgs(config)
	require.NoError(t, err)

	assert.Contains(t, args, "--test_tag_filters=-integration,-manual")
}

func TestBuildBazelArgs_BazelConfigsAfterVerbBeforeGavelFlags(t *testing.T) {
	config := AnalysisConfig{
		Targets:      []string{"//..."},
		Aspects:      []catalog.Aspect{{Path: "p"}},
		BazelConfigs: []string{"remote", "ci"},
	}

	args, err := buildBazelArgs(config)

	require.NoError(t, err)
	verbIdx := slices.Index(args, "build")
	require.GreaterOrEqual(t, verbIdx, 0, "args must start with the bazel verb")

	remoteIdx := slices.Index(args, "--config=remote")
	ciIdx := slices.Index(args, "--config=ci")
	require.GreaterOrEqual(t, remoteIdx, 0, "expected --config=remote in args")
	require.GreaterOrEqual(t, ciIdx, 0, "expected --config=ci in args")
	assert.Greater(t, remoteIdx, verbIdx, "--config=remote must come after the bazel verb")
	assert.Greater(t, ciIdx, verbIdx, "--config=ci must come after the bazel verb")

	aspectsIdx := slices.IndexFunc(args, func(a string) bool { return strings.HasPrefix(a, "--aspects=") })
	require.GreaterOrEqual(t, aspectsIdx, 0)
	assert.Less(t, remoteIdx, aspectsIdx, "--config flags must come before gavel's own --aspects flag")
	assert.Less(t, ciIdx, aspectsIdx, "--config flags must come before gavel's own --aspects flag")

	outputGroupsIdx := slices.Index(args, "--output_groups=gavel_submissions")
	keepGoingIdx := slices.Index(args, "--keep_going")
	assert.Less(t, remoteIdx, outputGroupsIdx)
	assert.Less(t, remoteIdx, keepGoingIdx)

	markerIdx := slices.Index(args, "--")
	require.GreaterOrEqual(t, markerIdx, 0)
	assert.Less(t, aspectsIdx, markerIdx, "-- separator and targets must stay at the very end")
	assert.Greater(t, slices.Index(args, "//..."), markerIdx)
}

func TestBuildBazelArgs_RejectsInvalidBazelConfigName(t *testing.T) {
	tests := map[string]string{
		"empty":            "",
		"contains space":   "has space",
		"contains equals":  "has=equals",
		"starts with dash": "-startsdash",
	}

	for name, badValue := range tests {
		t.Run(name, func(t *testing.T) {
			config := AnalysisConfig{
				Targets:      []string{"//..."},
				BazelConfigs: []string{badValue},
			}

			_, err := buildBazelArgs(config)

			require.Error(t, err, "expected an error for bazel_config value %q", badValue)
			assert.Contains(t, err.Error(), "bazel_config")
		})
	}
}

func TestBuildBazelArgs_MultipleAspects(t *testing.T) {
	config := AnalysisConfig{
		Targets: []string{"//..."},
		Aspects: []catalog.Aspect{
			{Path: "@gavel//a:defs.bzl%lint_a"},
			{Path: "@gavel//a:defs.bzl%lint_b"},
		},
	}

	args, err := buildBazelArgs(config)
	require.NoError(t, err)

	found := false
	for _, a := range args {
		if a == "--aspects=@gavel//a:defs.bzl%lint_a,@gavel//a:defs.bzl%lint_b" {
			found = true
		}
	}
	assert.True(t, found, "expected combined --aspects flag")
}

func TestCollectAllSARIF_GroupsByAspect(t *testing.T) {
	dir := t.TempDir()

	createSARIFFile(t, dir, "pkg/foo", "foo.golangci.sarif", `{"runs":[]}`)
	createSARIFFile(t, dir, "pkg/bar", "bar.golangci.sarif", `{"runs":[]}`)
	createSARIFFile(t, dir, "pkg/foo", "foo.archtest.sarif", `{"runs":[]}`)

	aspects := []catalog.Aspect{
		{Name: "golangci", SARIFSuffix: ".golangci.sarif"},
		{Name: "archtest", SARIFSuffix: ".archtest.sarif"},
	}

	result, err := collectAllSARIF(dir, aspects)

	require.NoError(t, err)
	assert.Len(t, result["golangci"], 2)
	assert.Len(t, result["archtest"], 1)
}

func TestCollectAllSARIF_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	aspects := []catalog.Aspect{
		{Name: "golangci", SARIFSuffix: ".golangci.sarif"},
	}

	result, err := collectAllSARIF(dir, aspects)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestScopeBinDirScopsToTargetSubdir(t *testing.T) {
	binDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(binDir, "apps", "web"), 0o755))

	scoped := scopeBinDir(binDir, []string{"//apps/web/..."})

	assert.Equal(t, filepath.Join(binDir, "apps", "web"), scoped)
}

func TestScopeBinDirRootPatternReturnsFullDir(t *testing.T) {
	binDir := t.TempDir()

	scoped := scopeBinDir(binDir, []string{"//..."})

	assert.Equal(t, binDir, scoped)
}

func TestScopeBinDirEmptyTargetsReturnsFullDir(t *testing.T) {
	binDir := t.TempDir()

	scoped := scopeBinDir(binDir, nil)

	assert.Equal(t, binDir, scoped)
}

func TestScopeBinDirNonExistentSubdirReturnsFullDir(t *testing.T) {
	binDir := t.TempDir()

	scoped := scopeBinDir(binDir, []string{"//nonexistent/..."})

	assert.Equal(t, binDir, scoped)
}

func TestCollectAllSARIF_ScopedDoesNotLeakCrossProject(t *testing.T) {
	binDir := t.TempDir()

	createSARIFFile(t, binDir, "core/domain", "core.archtest.sarif", `{"runs":[{"results":[{"message":{"text":"leak"}}]}]}`)
	createSARIFFile(t, binDir, "apps/web", "web.archtest.sarif", `{"runs":[{"results":[]}]}`)

	scoped := scopeBinDir(binDir, []string{"//apps/web/..."})
	aspects := []catalog.Aspect{{Name: "archtest", SARIFSuffix: ".archtest.sarif"}}

	result, err := collectAllSARIF(scoped, aspects)

	require.NoError(t, err)
	assert.Len(t, result["archtest"], 1, "should only find web's SARIF, not core's")
}

func TestSARIFReportsFromResult_SkipsMissingAspects(t *testing.T) {
	result := &AnalysisResult{
		SARIFFiles: map[string][][]byte{
			"golangci": {[]byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`)},
		},
	}
	aspects := []catalog.Aspect{
		{Name: "golangci"},
		{Name: "nonexistent"},
	}

	reports := SARIFReportsFromResult(result, aspects)

	require.Len(t, reports, 1)
	assert.Equal(t, "golangci-lint", reports[0].Source)
}

func TestSARIFReportsFromResult_EmptySARIFFiles(t *testing.T) {
	result := &AnalysisResult{SARIFFiles: map[string][][]byte{}}

	reports := SARIFReportsFromResult(result, []catalog.Aspect{{Name: "asp"}})

	assert.Nil(t, reports)
}

func TestRunAnalysis_Success(t *testing.T) {
	binDir := t.TempDir()
	createSARIFFile(t, binDir, "core", "core.golangci.sarif", `{"runs":[]}`)

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("build ok\n")},
		{Stdout: []byte(binDir + "\n")},
	}}
	config := AnalysisConfig{
		Workspace: "/ws",
		Targets:   []string{"//core/..."},
		Aspects:   []catalog.Aspect{{Name: "golangci", SARIFSuffix: ".golangci.sarif"}},
	}

	result, err := runAnalysis(t.Context(), fake, config)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.SARIFFiles["golangci"], 1)
	assert.Nil(t, result.BuildWarning)
}

func TestRunAnalysis_BuildAndBinDirError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("build failed"), Err: fmt.Errorf("build error")},
		{Stderr: []byte("not ws"), Err: fmt.Errorf("bindir error")},
	}}
	config := AnalysisConfig{
		Workspace: "/ws",
		Targets:   []string{"//..."},
		Aspects:   []catalog.Aspect{{Name: "golangci", SARIFSuffix: ".golangci.sarif"}},
	}

	_, err := runAnalysis(t.Context(), fake, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "analysis")
	assert.Contains(t, err.Error(), "build error")
}

func TestRunAnalysis_BinDirErrorOnly(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("ok\n")},
		{Stderr: []byte("fail"), Err: fmt.Errorf("bindir error")},
	}}
	config := AnalysisConfig{
		Workspace: "/ws",
		Targets:   []string{"//..."},
		Aspects:   []catalog.Aspect{{Name: "golangci"}},
	}

	_, err := runAnalysis(t.Context(), fake, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel info bazel-bin")
}

func TestRunAnalysis_BuildErrorNoSARIF(t *testing.T) {
	binDir := t.TempDir()
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("build failed"), Err: fmt.Errorf("build error")},
		{Stdout: []byte(binDir + "\n")},
	}}
	config := AnalysisConfig{
		Workspace: "/ws",
		Targets:   []string{"//..."},
		Aspects:   []catalog.Aspect{{Name: "golangci", SARIFSuffix: ".golangci.sarif"}},
	}

	_, err := runAnalysis(t.Context(), fake, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "build error")
}

func TestRunAnalysis_BuildErrorWithSARIF(t *testing.T) {
	binDir := t.TempDir()
	createSARIFFile(t, binDir, "core", "core.golangci.sarif", `{"runs":[]}`)

	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("partial"), Err: fmt.Errorf("build error")},
		{Stdout: []byte(binDir + "\n")},
	}}
	config := AnalysisConfig{
		Workspace: "/ws",
		Targets:   []string{"//core/..."},
		Aspects:   []catalog.Aspect{{Name: "golangci", SARIFSuffix: ".golangci.sarif"}},
	}

	result, err := runAnalysis(t.Context(), fake, config)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.BuildWarning)
	assert.Len(t, result.SARIFFiles["golangci"], 1)
}

func TestRunAnalysis_WithCoverage(t *testing.T) {
	workspace := t.TempDir()
	binDir := t.TempDir()
	createCoverageFile(t, filepath.Join(workspace, "bazel-testlogs"), "pkg/main_test", "SF:main.go\nDA:1,1\nend_of_record\n")

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("coverage ok\n")},
		{Stdout: []byte(binDir + "\n")},
	}}
	config := AnalysisConfig{
		Workspace:       workspace,
		Targets:         []string{"//..."},
		Aspects:         []catalog.Aspect{{Name: "golangci", SARIFSuffix: ".golangci.sarif"}},
		IncludeCoverage: true,
	}

	result, err := runAnalysis(t.Context(), fake, config)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, string(result.CoverageData), "SF:main.go")
}

func createSARIFFile(t *testing.T, baseDir, subdir, name, content string) {
	t.Helper()
	dir := filepath.Join(baseDir, subdir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}
