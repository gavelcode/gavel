package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze_Success(t *testing.T) {
	binDir := t.TempDir()
	sarifData := `{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"golangci-lint","rules":[{"id":"errcheck","shortDescription":{"text":"errcheck"}}]}},"results":[{"ruleId":"errcheck","level":"error","message":{"text":"unchecked error"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"main.go"},"region":{"startLine":10}}}],"fingerprints":{"gavel/v1":"abc123"}}]}]}`
	createSARIFFile(t, binDir, "pkg", "pkg.golangci.sarif", sarifData)

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("build ok\n")},
		{Stdout: []byte(binDir + "\n")},
	}}
	analyzer := NewBazelTargetAnalyzer(fake)

	findings, err := analyzer.Analyze(t.Context(), "/ws", "//pkg:lib", []string{"go"})

	require.NoError(t, err)
	require.NotEmpty(t, findings)
	assert.Equal(t, "errcheck", findings[0].RuleID)
	assert.Equal(t, "main.go", findings[0].FilePath)
}

func TestAnalyze_BuildError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("compilation failed"), Err: fmt.Errorf("exit 1")},
	}}
	analyzer := NewBazelTargetAnalyzer(fake)

	_, err := analyzer.Analyze(t.Context(), "/ws", "//pkg:lib", []string{"go"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel")
}

func TestAnalyze_BinDirError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("build ok\n")},
		{Stderr: []byte("not a workspace"), Err: fmt.Errorf("exit 1")},
	}}
	analyzer := NewBazelTargetAnalyzer(fake)

	_, err := analyzer.Analyze(t.Context(), "/ws", "//pkg:lib", []string{"go"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel info bazel-bin")
}

func TestAnalyze_NoSARIFFiles(t *testing.T) {
	binDir := t.TempDir()
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("build ok\n")},
		{Stdout: []byte(binDir + "\n")},
	}}
	analyzer := NewBazelTargetAnalyzer(fake)

	findings, err := analyzer.Analyze(t.Context(), "/ws", "//pkg:lib", []string{"go"})

	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestAnalyze_InvalidSARIF(t *testing.T) {
	binDir := t.TempDir()
	dir := filepath.Join(binDir, "pkg")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg.go_golangci_lint_submission_aspect.sarif"), []byte("not json"), 0o644))

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("build ok\n")},
		{Stdout: []byte(binDir + "\n")},
	}}
	analyzer := NewBazelTargetAnalyzer(fake)

	findings, err := analyzer.Analyze(t.Context(), "/ws", "//pkg:lib", []string{"go"})

	require.NoError(t, err)
	assert.Empty(t, findings)
}
