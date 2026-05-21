package runner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRdepsQuery_DefaultScope(t *testing.T) {
	got := buildRdepsQuery("file1.go file2.go", "")

	assert.Equal(t, `rdeps(//..., set(file1.go file2.go))`, got)
}

func TestBuildRdepsQuery_WithProjectScope(t *testing.T) {
	got := buildRdepsQuery("file1.go", "//core/...")

	assert.Equal(t, `rdeps(//core/..., set(file1.go))`, got)
}

func TestBuildAffectedFileSet_StripsWorkspacePrefix(t *testing.T) {
	got := buildAffectedFileSet("/workspace", []string{"/workspace/core/main.go", "/workspace/apps/server.go"})

	assert.Equal(t, "core/main.go apps/server.go", got)
}

func TestBuildAffectedFileSet_RelativePaths(t *testing.T) {
	got := buildAffectedFileSet("/workspace", []string{"core/main.go"})

	assert.Equal(t, "core/main.go", got)
}

func TestParseAffectedLabels_FiltersNonLabels(t *testing.T) {
	input := "Loading: 5 packages\n//core:main\nERROR: something\n//apps:server\n\n"

	got := parseAffectedLabels(input)

	assert.Equal(t, []string{"//core:main", "//apps:server"}, got)
}

func TestParseAffectedLabels_EmptyOutput(t *testing.T) {
	got := parseAffectedLabels("")

	assert.Empty(t, got)
}

func TestBuildOwnerTarget(t *testing.T) {
	got := buildOwnerTarget("pkg/data/bi-storage-uploader/internal/metrics")

	assert.Equal(t, "//pkg/data/bi-storage-uploader/internal/metrics:all", got)
}

func TestBuildOwnerTarget_RootPackage(t *testing.T) {
	got := buildOwnerTarget("")

	assert.Equal(t, "//:all", got)
}

func TestExtractPackagePath_WithTargetName(t *testing.T) {
	assert.Equal(t, "internal/domain/order", extractPackagePath("//internal/domain/order:all"))
}

func TestExtractPackagePath_WildcardPattern(t *testing.T) {
	assert.Equal(t, "core", extractPackagePath("//core/..."))
}

func TestExtractPackagePath_BareLabel(t *testing.T) {
	assert.Equal(t, "pkg/data", extractPackagePath("//pkg/data:go_default_library"))
}

func TestExtractPackagePath_RootTarget(t *testing.T) {
	assert.Equal(t, "", extractPackagePath("//:main"))
}

func TestFindAffectedTargets_Success(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("//core:main\n//apps:server\n")},
	}}
	resolver := NewBazelTargetResolver(fake)

	targets, err := resolver.FindAffectedTargets(t.Context(), "/ws", []string{"/ws/core/main.go"}, "")

	require.NoError(t, err)
	assert.Equal(t, []string{"//core:main", "//apps:server"}, targets)
	require.Len(t, fake.calls, 1)
	assert.Equal(t, "bazel", fake.calls[0].Name)
	assert.Equal(t, "/ws", fake.calls[0].Dir)
}

func TestFindAffectedTargets_EmptyChangedFiles(t *testing.T) {
	fake := &fakeRunner{}
	resolver := NewBazelTargetResolver(fake)

	targets, err := resolver.FindAffectedTargets(t.Context(), "/ws", nil, "")

	require.NoError(t, err)
	assert.Nil(t, targets)
	assert.Empty(t, fake.calls)
}

func TestFindAffectedTargets_ErrorWithPartialResults(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("//core:main\n"), Err: fmt.Errorf("partial failure")},
	}}
	resolver := NewBazelTargetResolver(fake)

	targets, err := resolver.FindAffectedTargets(t.Context(), "/ws", []string{"/ws/a.go"}, "//core/...")

	require.NoError(t, err)
	assert.Equal(t, []string{"//core:main"}, targets)
}

func TestFindAffectedTargets_ErrorNoResults(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("no targets found"), Err: fmt.Errorf("query failed")},
	}}
	resolver := NewBazelTargetResolver(fake)

	_, err := resolver.FindAffectedTargets(t.Context(), "/ws", []string{"/ws/a.go"}, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel query")
	assert.Contains(t, err.Error(), "no targets found")
}

func TestFindOwnerTarget_Success(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("pkg/data/service\n")},
	}}
	resolver := NewBazelTargetResolver(fake)

	target, err := resolver.FindOwnerTarget(t.Context(), "/ws", "pkg/data/service/main.go")

	require.NoError(t, err)
	assert.Equal(t, "//pkg/data/service:all", target)
}

func TestFindOwnerTarget_EmptyPackage(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("\n")},
	}}
	resolver := NewBazelTargetResolver(fake)

	_, err := resolver.FindOwnerTarget(t.Context(), "/ws", "orphan.go")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no package found")
}

func TestFindOwnerTarget_RunnerError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("not found"), Err: fmt.Errorf("query failed")},
	}}
	resolver := NewBazelTargetResolver(fake)

	_, err := resolver.FindOwnerTarget(t.Context(), "/ws", "missing.go")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel query owner")
}
