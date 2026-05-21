package runner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBazelBinDir_Success(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("/private/var/tmp/_bazel/bazel-bin\n")},
	}}

	binDir, err := bazelBinDir(t.Context(), fake, "/ws")

	require.NoError(t, err)
	assert.Equal(t, "/private/var/tmp/_bazel/bazel-bin", binDir)
	require.Len(t, fake.calls, 1)
	assert.Equal(t, "bazel", fake.calls[0].Name)
	assert.Equal(t, []string{"info", "bazel-bin"}, fake.calls[0].Args)
}

func TestBazelBinDir_Error(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("not a workspace"), Err: fmt.Errorf("exit 1")},
	}}

	_, err := bazelBinDir(t.Context(), fake, "/ws")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel info bazel-bin")
	assert.Contains(t, err.Error(), "not a workspace")
}

func TestBazelOutputPathWith_Success(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("/private/var/tmp/_bazel/output\n")},
	}}

	path, err := bazelOutputPathWith(t.Context(), fake, "/ws")

	require.NoError(t, err)
	assert.Equal(t, "/private/var/tmp/_bazel/output", path)
}

func TestBazelOutputPathWith_Error(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("fail"), Err: fmt.Errorf("exit 1")},
	}}

	_, err := bazelOutputPathWith(t.Context(), fake, "/ws")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel info output_path")
}
