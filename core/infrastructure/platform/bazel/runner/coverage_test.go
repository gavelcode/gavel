package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectIndividualCoverageFiles_FindsFiles(t *testing.T) {
	dir := t.TempDir()
	testlogsDir := filepath.Join(dir, "bazel-testlogs")

	createCoverageFile(t, testlogsDir, "pkg/foo_test", "SF:pkg/foo.go\nDA:1,1\nend_of_record\n")
	createCoverageFile(t, testlogsDir, "pkg/bar_test", "SF:pkg/bar.go\nDA:1,0\nend_of_record\n")

	data, count, err := collectIndividualCoverageFiles(t.Context(), dir)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Contains(t, string(data), "SF:pkg/foo.go")
	assert.Contains(t, string(data), "SF:pkg/bar.go")
}

func TestCollectIndividualCoverageFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bazel-testlogs"), 0o755))

	data, count, err := collectIndividualCoverageFiles(t.Context(), dir)

	require.NoError(t, err)
	assert.Nil(t, data)
	assert.Equal(t, 0, count)
}

func TestCollectIndividualCoverageFiles_SkipsEmptyFiles(t *testing.T) {
	dir := t.TempDir()
	testlogsDir := filepath.Join(dir, "bazel-testlogs")

	createCoverageFile(t, testlogsDir, "pkg/foo_test", "SF:pkg/foo.go\nDA:1,1\nend_of_record\n")
	createCoverageFile(t, testlogsDir, "pkg/empty_test", "")

	data, count, err := collectIndividualCoverageFiles(t.Context(), dir)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Contains(t, string(data), "SF:pkg/foo.go")
}

func TestCollectIndividualCoverageFiles_ConcatenatesValidLCOV(t *testing.T) {
	dir := t.TempDir()
	testlogsDir := filepath.Join(dir, "bazel-testlogs")

	lcov1 := "SF:a.go\nDA:1,1\nDA:2,0\nLF:2\nLH:1\nend_of_record\n"
	lcov2 := "SF:b.go\nDA:1,1\nLF:1\nLH:1\nend_of_record\n"
	createCoverageFile(t, testlogsDir, "a_test", lcov1)
	createCoverageFile(t, testlogsDir, "b_test", lcov2)

	data, count, err := collectIndividualCoverageFiles(t.Context(), dir)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Contains(t, string(data), "end_of_record")
	assert.Contains(t, string(data), "SF:a.go")
	assert.Contains(t, string(data), "SF:b.go")
}

func TestCollectIndividualCoverageFiles_FollowsTestlogsSymlink(t *testing.T) {
	realTestlogs := t.TempDir()
	createCoverageFile(t, realTestlogs, "pkg/foo_test", "SF:pkg/foo.go\nDA:1,1\nend_of_record\n")

	workspace := t.TempDir()
	require.NoError(t, os.Symlink(realTestlogs, filepath.Join(workspace, "bazel-testlogs")))

	data, count, err := collectIndividualCoverageFiles(t.Context(), workspace)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Contains(t, string(data), "SF:pkg/foo.go")
}

func TestResolveTestlogsDir_PrefersSymlink(t *testing.T) {
	dir := t.TempDir()
	testlogsDir := filepath.Join(dir, "bazel-testlogs")
	require.NoError(t, os.MkdirAll(testlogsDir, 0o755))

	got, err := resolveTestlogsDir(t.Context(), dir)

	require.NoError(t, err)
	assert.Equal(t, testlogsDir, got)
}

func TestRunCoverage_Success(t *testing.T) {
	outputPath := t.TempDir()
	coverageDir := filepath.Join(outputPath, "_coverage")
	require.NoError(t, os.MkdirAll(coverageDir, 0o755))
	lcov := "SF:main.go\nDA:1,1\nend_of_record\n"
	require.NoError(t, os.WriteFile(filepath.Join(coverageDir, "_coverage_report.dat"), []byte(lcov), 0o644))

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("coverage ok\n")},
		{Stdout: []byte(outputPath + "\n")},
	}}

	result, err := runCoverage(t.Context(), fake, "/ws", []string{"//..."})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, lcov, string(result.Data))
	assert.Nil(t, result.Warning)
}

func TestRunCoverage_NoReport(t *testing.T) {
	outputPath := t.TempDir()
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("ok\n")},
		{Stdout: []byte(outputPath + "\n")},
	}}

	result, err := runCoverage(t.Context(), fake, "/ws", []string{"//..."})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Data)
}

func TestRunCoverage_RunErrorNoReport(t *testing.T) {
	outputPath := t.TempDir()
	fake := &fakeRunner{results: []fakeResult{
		{Err: fmt.Errorf("coverage failed")},
		{Stdout: []byte(outputPath + "\n")},
	}}

	result, err := runCoverage(t.Context(), fake, "/ws", []string{"//..."})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Data)
	require.NotNil(t, result.Warning)
	assert.Contains(t, result.Warning.Error(), "coverage failed")
}

func TestRunCoverage_RunErrorWithReport(t *testing.T) {
	outputPath := t.TempDir()
	coverageDir := filepath.Join(outputPath, "_coverage")
	require.NoError(t, os.MkdirAll(coverageDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(coverageDir, "_coverage_report.dat"), []byte("SF:x\nend_of_record\n"), 0o644))

	fake := &fakeRunner{results: []fakeResult{
		{Err: fmt.Errorf("partial failure")},
		{Stdout: []byte(outputPath + "\n")},
	}}

	result, err := runCoverage(t.Context(), fake, "/ws", []string{"//..."})

	require.NoError(t, err)
	assert.NotNil(t, result.Data)
	assert.Contains(t, result.Warning.Error(), "partial failures")
}

func TestRunCoverage_OutputPathError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("ok\n")},
		{Stderr: []byte("fail"), Err: fmt.Errorf("info error")},
	}}

	_, err := runCoverage(t.Context(), fake, "/ws", []string{"//..."})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel info output_path")
}

func TestFindCombinedReportWith_BuildsPath(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("/output\n")},
	}}

	path, err := findCombinedReportWith(t.Context(), fake, "/ws")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/output", "_coverage", "_coverage_report.dat"), path)
}

func TestResolveTestlogsDirWith_FallbackToBazelInfo(t *testing.T) {
	workspace := t.TempDir()
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("/some/output_base/execroot/project/bazel-out\n")},
	}}

	got, err := resolveTestlogsDirWith(t.Context(), fake, workspace)

	require.NoError(t, err)
	assert.Equal(t, "/some/output_base/execroot/project/testlogs", got)
}

func TestResolveTestlogsDirWith_BazelInfoError(t *testing.T) {
	workspace := t.TempDir()
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("fail"), Err: fmt.Errorf("info error")},
	}}

	_, err := resolveTestlogsDirWith(t.Context(), fake, workspace)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel info output_path")
}

func TestCollectCoverageDataWith_CollectsIndividual(t *testing.T) {
	workspace := t.TempDir()
	outputPath := t.TempDir()
	testlogsDir := filepath.Join(workspace, "bazel-testlogs")
	createCoverageFile(t, testlogsDir, "pkg/foo_test", "SF:foo.go\nDA:1,1\nend_of_record\n")

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte(outputPath + "\n")},
	}}

	data, warning := collectCoverageDataWith(t.Context(), fake, workspace, nil)

	require.NotNil(t, data)
	assert.Contains(t, string(data), "SF:foo.go")
	assert.Nil(t, warning)
}

func TestCollectCoverageDataWith_CollectsIndividualWithRunErr(t *testing.T) {
	workspace := t.TempDir()
	outputPath := t.TempDir()
	testlogsDir := filepath.Join(workspace, "bazel-testlogs")
	createCoverageFile(t, testlogsDir, "pkg/foo_test", "SF:foo.go\nDA:1,1\nend_of_record\n")

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte(outputPath + "\n")},
	}}

	data, warning := collectCoverageDataWith(t.Context(), fake, workspace, fmt.Errorf("coverage failed"))

	require.NotNil(t, data)
	require.NotNil(t, warning)
	assert.Contains(t, warning.Error(), "partial failures")
}

func TestCollectCoverageDataWith_NoDataNoRunErr(t *testing.T) {
	workspace := t.TempDir()
	outputPath := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, "bazel-testlogs"), 0o755))

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte(outputPath + "\n")},
	}}

	data, warning := collectCoverageDataWith(t.Context(), fake, workspace, nil)

	assert.Nil(t, data)
	assert.Nil(t, warning)
}

func TestCollectCoverageDataWith_NoDataWithRunErr(t *testing.T) {
	workspace := t.TempDir()
	outputPath := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, "bazel-testlogs"), 0o755))

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte(outputPath + "\n")},
	}}
	runErr := fmt.Errorf("coverage failed")

	data, warning := collectCoverageDataWith(t.Context(), fake, workspace, runErr)

	assert.Nil(t, data)
	require.NotNil(t, warning)
	assert.Contains(t, warning.Error(), "coverage failed")
}

func TestCollectCoverageDataWith_ResolveTestlogsError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("fail"), Err: fmt.Errorf("info error")},
	}}

	data, warning := collectCoverageDataWith(t.Context(), fake, "/ws", nil)

	assert.Nil(t, data)
	require.NotNil(t, warning)
	assert.Contains(t, warning.Error(), "collect coverage")
}

func TestCollectCoverageDataWith_FallbackErrorWithRunErr(t *testing.T) {
	workspace := t.TempDir()
	outputPath := t.TempDir()

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte(outputPath + "\n")},
	}}
	runErr := fmt.Errorf("coverage failed")

	data, warning := collectCoverageDataWith(t.Context(), fake, workspace, runErr)

	assert.Nil(t, data)
	require.NotNil(t, warning)
	assert.Contains(t, warning.Error(), "coverage failed")
}

func createCoverageFile(t *testing.T, testlogsDir, testPath, content string) {
	t.Helper()
	dir := filepath.Join(testlogsDir, testPath)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	if content != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "coverage.dat"), []byte(content), 0o644))
	} else {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "coverage.dat"), nil, 0o644))
	}
}
