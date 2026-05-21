package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

func TestAspectCollectorCollect_Success(t *testing.T) {
	binDir := t.TempDir()
	sarifData := `{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`
	createSARIFFile(t, binDir, "pkg", "pkg.golangci.sarif", sarifData)

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("build ok\n")},
		{Stdout: []byte(binDir + "\n")},
	}}
	asp := catalog.Aspect{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"}
	collector := NewAspectCollector([]catalog.Aspect{asp}, fake)

	reports, err := collector.Collect(t.Context(), "/ws", []string{"//pkg:lib"})

	require.NoError(t, err)
	require.Len(t, reports, 1)
	assert.Equal(t, "golangci-lint", reports[0].Source)
}

func TestAspectCollectorCollect_MultipleAspects(t *testing.T) {
	binDir := t.TempDir()
	createSARIFFile(t, binDir, "pkg", "pkg.golangci.sarif", `{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`)
	createSARIFFile(t, binDir, "pkg", "pkg.archtest.sarif", `{"runs":[{"tool":{"driver":{"name":"archtest"}},"results":[]}]}`)

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("build ok\n")},
		{Stdout: []byte(binDir + "\n")},
		{Stdout: []byte("build ok\n")},
		{Stdout: []byte(binDir + "\n")},
	}}
	aspects := []catalog.Aspect{
		{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"},
		{Name: "archtest", Path: "@gavel//a:defs.bzl%arch", SARIFSuffix: ".archtest.sarif"},
	}
	collector := NewAspectCollector(aspects, fake)

	reports, err := collector.Collect(t.Context(), "/ws", []string{"//pkg:lib"})

	require.NoError(t, err)
	assert.Len(t, reports, 2)
}

func TestAspectCollectorCollect_AspectError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("build failed"), Err: fmt.Errorf("build error")},
		{Stderr: []byte("not ws"), Err: fmt.Errorf("bindir error")},
	}}
	asp := catalog.Aspect{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"}
	collector := NewAspectCollector([]catalog.Aspect{asp}, fake)

	_, err := collector.Collect(t.Context(), "/ws", []string{"//pkg:lib"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "aspect golangci")
}

func TestReportCollectorCollect_Success(t *testing.T) {
	binDir := t.TempDir()
	sarifData := []byte(`{"runs":[{"tool":{"driver":{"name":"PMD"}},"results":[]}]}`)
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "pkg.AspectRulesLint.report"), sarifData, 0o644))

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte(binDir + "\n")},
	}}
	collector := NewReportCollector(fake)

	reports, err := collector.Collect(t.Context(), "/ws", nil)

	require.NoError(t, err)
	require.Len(t, reports, 1)
	assert.Equal(t, "PMD", reports[0].Source)
}

func TestReportCollectorCollect_BinDirError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("fail"), Err: fmt.Errorf("bindir error")},
	}}
	collector := NewReportCollector(fake)

	_, err := collector.Collect(t.Context(), "/ws", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel info bazel-bin")
}

func TestHybridCollectorCollect_Success(t *testing.T) {
	binDir := t.TempDir()
	reportData := []byte(`{"runs":[{"tool":{"driver":{"name":"PMD"}},"results":[]}]}`)
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "pkg.AspectRulesLint.report"), reportData, 0o644))
	createSARIFFile(t, binDir, "pkg", "pkg.golangci.sarif", `{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[]}]}`)

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte(binDir + "\n")},
		{Stdout: []byte("build ok\n")},
		{Stdout: []byte(binDir + "\n")},
	}}
	exclusiveAspects := []catalog.Aspect{
		{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"},
	}
	collector := NewHybridCollector(exclusiveAspects, fake)

	reports, err := collector.Collect(t.Context(), "/ws", []string{"//pkg:lib"})

	require.NoError(t, err)
	assert.Len(t, reports, 2)
}

func TestHybridCollectorCollect_PrimaryError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("fail"), Err: fmt.Errorf("bindir error")},
	}}
	collector := NewHybridCollector(nil, fake)

	_, err := collector.Collect(t.Context(), "/ws", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel info bazel-bin")
}

func TestHybridCollectorCollect_SupplementError(t *testing.T) {
	binDir := t.TempDir()
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte(binDir + "\n")},
		{Stderr: []byte("build failed"), Err: fmt.Errorf("build error")},
		{Stderr: []byte("not ws"), Err: fmt.Errorf("bindir error")},
	}}
	exclusiveAspects := []catalog.Aspect{
		{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"},
	}
	collector := NewHybridCollector(exclusiveAspects, fake)

	_, err := collector.Collect(t.Context(), "/ws", []string{"//pkg:lib"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "aspect golangci")
}

func TestHasRulesLintReportsWithRunner_Found(t *testing.T) {
	binDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "pkg.AspectRulesLint.report"), []byte(`{}`), 0o644))

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte(binDir + "\n")},
	}}

	got := hasRulesLintReportsWithRunner(t.Context(), fake, "/ws")

	assert.True(t, got)
}

func TestHasRulesLintReportsWithRunner_BinDirError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("fail"), Err: fmt.Errorf("bindir error")},
	}}

	got := hasRulesLintReportsWithRunner(t.Context(), fake, "/ws")

	assert.False(t, got)
}
