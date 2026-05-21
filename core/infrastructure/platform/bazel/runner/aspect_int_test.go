package runner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

func TestRunAspect_Success(t *testing.T) {
	binDir := t.TempDir()
	createSARIFFile(t, binDir, "pkg", "pkg.golangci.sarif", `{"runs":[]}`)

	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("build ok\n")},
		{Stdout: []byte(binDir + "\n")},
	}}
	asp := catalog.Aspect{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"}

	results, err := runAspect(t.Context(), fake, "/ws", []string{"//pkg:lib"}, asp)

	require.NoError(t, err)
	require.Len(t, results, 1)
}

func TestRunAspect_BuildAndBinDirError(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("build failed"), Err: fmt.Errorf("build error")},
		{Stderr: []byte("not workspace"), Err: fmt.Errorf("bindir error")},
	}}
	asp := catalog.Aspect{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"}

	_, err := runAspect(t.Context(), fake, "/ws", []string{"//pkg:lib"}, asp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "aspect golangci")
	assert.Contains(t, err.Error(), "build error")
}

func TestRunAspect_BinDirErrorOnly(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("build ok\n")},
		{Stderr: []byte("not workspace"), Err: fmt.Errorf("bindir error")},
	}}
	asp := catalog.Aspect{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"}

	_, err := runAspect(t.Context(), fake, "/ws", []string{"//pkg:lib"}, asp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel info bazel-bin")
}

func TestRunAspect_BuildErrorNoSARIF(t *testing.T) {
	binDir := t.TempDir()
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("build failed"), Err: fmt.Errorf("build error")},
		{Stdout: []byte(binDir + "\n")},
	}}
	asp := catalog.Aspect{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"}

	_, err := runAspect(t.Context(), fake, "/ws", []string{"//pkg:lib"}, asp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "aspect golangci")
}

func TestRunAspect_BuildErrorWithSARIF(t *testing.T) {
	binDir := t.TempDir()
	createSARIFFile(t, binDir, "pkg", "pkg.golangci.sarif", `{"runs":[]}`)

	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("partial failure"), Err: fmt.Errorf("build error")},
		{Stdout: []byte(binDir + "\n")},
	}}
	asp := catalog.Aspect{Name: "golangci", Path: "@gavel//a:defs.bzl%lint", SARIFSuffix: ".golangci.sarif"}

	results, err := runAspect(t.Context(), fake, "/ws", []string{"//pkg:lib"}, asp)

	require.NoError(t, err)
	require.Len(t, results, 1)
}
