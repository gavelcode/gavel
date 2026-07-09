package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

func TestBuildBazelArgs_TargetsAfterOptionsMarker(t *testing.T) {
	config := AnalysisConfig{
		Targets: []string{"//core/...", "-//core/gen/..."},
		Aspects: []catalog.Aspect{{Path: "p"}},
	}

	args := buildBazelArgs(config)

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

	args := buildBazelArgs(config)

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

	args := buildBazelArgs(config)

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

	args := buildBazelArgs(config)

	assert.Equal(t, "coverage", args[0])
	assert.Contains(t, args, "--combined_report=lcov")
	assert.Contains(t, args, "--test_size_filters=small,medium")
	for _, a := range args {
		assert.NotContains(t, a, "instrumentation_filter")
	}
}

func TestBuildBazelArgs_CoverageBoundsMemory(t *testing.T) {
	cov := buildBazelArgs(AnalysisConfig{Targets: []string{"//..."}, IncludeCoverage: true})
	assert.Contains(t, cov, "--local_resources=memory=HOST_RAM*0.67")

	build := buildBazelArgs(AnalysisConfig{Targets: []string{"//..."}, IncludeCoverage: false})
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

	args := buildBazelArgs(config)

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

	args := buildBazelArgs(config)

	assert.Contains(t, args, "--test_size_filters=small,medium")
}

func TestBuildBazelArgs_CustomTestSizeFilters(t *testing.T) {
	config := AnalysisConfig{
		Targets:         []string{"//..."},
		Aspects:         []catalog.Aspect{{Path: "p"}},
		IncludeCoverage: true,
		TestSizeFilters: "small",
	}

	args := buildBazelArgs(config)

	assert.Contains(t, args, "--test_size_filters=small")
	assert.NotContains(t, args, "--test_size_filters=small,medium")
}

func TestBuildBazelArgs_TestTagFilters(t *testing.T) {
	config := AnalysisConfig{
		Targets:        []string{"//..."},
		Aspects:        []catalog.Aspect{{Path: "p"}},
		TestTagFilters: "-integration,-manual",
	}

	args := buildBazelArgs(config)

	assert.Contains(t, args, "--test_tag_filters=-integration,-manual")
}

func TestBuildBazelArgs_MultipleAspects(t *testing.T) {
	config := AnalysisConfig{
		Targets: []string{"//..."},
		Aspects: []catalog.Aspect{
			{Path: "@gavel//a:defs.bzl%lint_a"},
			{Path: "@gavel//a:defs.bzl%lint_b"},
		},
	}

	args := buildBazelArgs(config)

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
	binDir := t.TempDir()
	outputPath := t.TempDir()
	coverageDir := filepath.Join(outputPath, "_coverage")
	require.NoError(t, os.MkdirAll(coverageDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(coverageDir, "_coverage_report.dat"), []byte("SF:main.go\nend_of_record\n"), 0o644))

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("coverage ok\n")},
		{Stdout: []byte(binDir + "\n")},
		{Stdout: []byte(outputPath + "\n")},
	}}
	config := AnalysisConfig{
		Workspace:       "/ws",
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
