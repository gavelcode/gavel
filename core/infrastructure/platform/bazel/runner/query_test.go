package runner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQueryLabels_EmptyOutput(t *testing.T) {
	assert.Empty(t, parseQueryLabels(nil))
	assert.Empty(t, parseQueryLabels([]byte("")))
	assert.Empty(t, parseQueryLabels([]byte("\n\n")))
}

func TestParseQueryLabels_SingleLine(t *testing.T) {
	got := parseQueryLabels([]byte("//backend:server\n"))
	assert.Equal(t, []string{"//backend:server"}, got)
}

func TestParseQueryLabels_MultipleLines(t *testing.T) {
	out := []byte("//backend:server\n//backend:server_test\n//backend/api:lib\n")
	got := parseQueryLabels(out)
	assert.Equal(t, []string{"//backend:server", "//backend:server_test", "//backend/api:lib"}, got)
}

func TestParseQueryLabels_TrimsAndSkipsBlanks(t *testing.T) {
	out := []byte("  //a:b  \n\n  //c:d\n  \n")
	got := parseQueryLabels(out)
	assert.Equal(t, []string{"//a:b", "//c:d"}, got)
}

func TestQueryTargetsOfKind_Success(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("//backend:server\n//backend:lib\n")},
	}}

	targets, err := queryTargetsOfKind(t.Context(), fake, "/ws", "//backend/...", []string{"go_library", "go_binary"})

	require.NoError(t, err)
	assert.Equal(t, []string{"//backend:server", "//backend:lib"}, targets)
	require.Len(t, fake.calls, 1)
	assert.Contains(t, fake.calls[0].Args[1], "kind")
}

func TestQueryTargetsOfKind_EmptyKinds(t *testing.T) {
	fake := &fakeRunner{}

	targets, err := queryTargetsOfKind(t.Context(), fake, "/ws", "//...", nil)

	require.NoError(t, err)
	assert.Nil(t, targets)
	assert.Empty(t, fake.calls)
}

func TestQueryTargetsOfKind_EmptyPattern(t *testing.T) {
	fake := &fakeRunner{}

	targets, err := queryTargetsOfKind(t.Context(), fake, "/ws", "  ", []string{"go_library"})

	require.NoError(t, err)
	assert.Nil(t, targets)
	assert.Empty(t, fake.calls)
}

func TestQueryTargetsOfKind_Error(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stderr: []byte("query error"), Err: fmt.Errorf("bazel failed")},
	}}

	_, err := queryTargetsOfKind(t.Context(), fake, "/ws", "//...", []string{"go_library"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel query")
}

func TestBazelTargetQueryDelegatesToInternal(t *testing.T) {
	fake := &fakeRunner{results: []fakeResult{
		{Stdout: []byte("//x:y\n")},
	}}
	q := NewBazelTargetQuery(fake)

	targets, err := q.QueryTargetsOfKind(t.Context(), "/ws", "//x/...", []string{"go_library"})

	require.NoError(t, err)
	assert.Equal(t, []string{"//x:y"}, targets)
}
