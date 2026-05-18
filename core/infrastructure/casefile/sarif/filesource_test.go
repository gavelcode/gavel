package sarif_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/infrastructure/casefile/sarif"
)

func TestFileSourceReaderReadsLine(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.go"), []byte("package main\n\n  func Foo() {\n}\n"), 0644))

	reader := sarif.NewFileSourceReader(dir)
	content, err := reader.ReadLine("file.go", 3)

	require.NoError(t, err)
	assert.Equal(t, "func Foo() {", content, "must strip leading/trailing whitespace")
}

func TestFileSourceReaderRejectsNonPositiveLine(t *testing.T) {
	reader := sarif.NewFileSourceReader(t.TempDir())

	_, err := reader.ReadLine("file.go", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line must be positive")

	_, err = reader.ReadLine("file.go", -1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line must be positive")
}

func TestFileSourceReaderReturnsErrorForMissingFile(t *testing.T) {
	reader := sarif.NewFileSourceReader(t.TempDir())

	_, err := reader.ReadLine("nonexistent.go", 1)

	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestFileSourceReaderOutOfRange(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.go"), []byte("line1\nline2\n"), 0644))

	reader := sarif.NewFileSourceReader(dir)
	_, err := reader.ReadLine("file.go", 99)

	require.Error(t, err)
}
