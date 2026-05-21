package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDiffOutputSingleHunk(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -10,0 +11,3 @@
+line1
+line2
+line3
`
	got, err := parseDiffOutput(diff)
	require.NoError(t, err)

	assert.Equal(t, []int{11, 12, 13}, got["main.go"])
}

func TestParseDiffOutputMultipleHunks(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -5,0 +6,2 @@
+a
+b
@@ -20,0 +22,1 @@
+c
`
	got, err := parseDiffOutput(diff)
	require.NoError(t, err)

	assert.Equal(t, []int{6, 7, 22}, got["main.go"])
}

func TestParseDiffOutputMultipleFiles(t *testing.T) {
	diff := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1,0 +2,1 @@
+new
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -5,0 +6,2 @@
+x
+y
`
	got, err := parseDiffOutput(diff)
	require.NoError(t, err)

	assert.Equal(t, []int{2}, got["a.go"])
	assert.Equal(t, []int{6, 7}, got["b.go"])
}

func TestParseDiffOutputSingleLineAddition(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -10,0 +11 @@
+single
`
	got, err := parseDiffOutput(diff)
	require.NoError(t, err)

	assert.Equal(t, []int{11}, got["main.go"])
}

func TestParseDiffOutputEmptyDiff(t *testing.T) {
	got, err := parseDiffOutput("")
	require.NoError(t, err)

	assert.Empty(t, got)
}

func TestParseDiffOutputNoHunks(t *testing.T) {
	diff := `diff --git a/binary.bin b/binary.bin
Binary files differ
`
	got, err := parseDiffOutput(diff)
	require.NoError(t, err)

	assert.Empty(t, got)
}

func TestParseDiffOutputHunkBeforeFileIsSkipped(t *testing.T) {
	diff := "@@ -1,0 +5,2 @@\n+a\n+b\n"

	got, err := parseDiffOutput(diff)
	require.NoError(t, err)

	assert.Empty(t, got)
}

func TestParseDiffOutputZeroCountHunk(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -5,3 +5,0 @@
`
	got, err := parseDiffOutput(diff)
	require.NoError(t, err)

	assert.Empty(t, got["main.go"])
}

func TestParseDiffOutputReturnsErrorOnOverflowStart(t *testing.T) {
	diff := "+++ b/main.go\n@@ -1 +99999999999999999999999 @@\n"

	_, err := parseDiffOutput(diff)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse hunk start")
}

func TestParseDiffOutputReturnsErrorOnOverflowCount(t *testing.T) {
	diff := "+++ b/main.go\n@@ -1 +1,99999999999999999999999 @@\n"

	_, err := parseDiffOutput(diff)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse hunk count")
}
