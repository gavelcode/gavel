package sarif_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/sarif"
)

type sarifDoc struct {
	Schema  string            `json:"$schema"`
	Version string            `json:"version"`
	Runs    []json.RawMessage `json:"runs"`
}

func TestWrite_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.sarif")

	docs := [][]byte{
		[]byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[{"ruleId":"unused","message":{"text":"unused var"}}]}]}`),
	}

	err := sarif.Write(path, docs)

	require.NoError(t, err)
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got sarifDoc
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, "2.1.0", got.Version)
	assert.Len(t, got.Runs, 1)
}

func TestWrite_NoDocsWritesEmptySARIF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.sarif")

	err := sarif.Write(path, nil)

	require.NoError(t, err)
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got sarifDoc
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, "2.1.0", got.Version)
	assert.Empty(t, got.Runs)
}

func TestWrite_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "report.sarif")

	err := sarif.Write(path, nil)

	require.NoError(t, err)
	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestWrite_MergesMultipleDocs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "merged.sarif")

	docs := [][]byte{
		[]byte(`{"runs":[{"tool":{"driver":{"name":"tool-a"}},"results":[]}]}`),
		[]byte(`{"runs":[{"tool":{"driver":{"name":"tool-b"}},"results":[]}]}`),
	}

	err := sarif.Write(path, docs)

	require.NoError(t, err)
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got sarifDoc
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Len(t, got.Runs, 2)
}

func TestWrite_MergeErrorReturnsWrapped(t *testing.T) {
	docs := [][]byte{
		[]byte(`not valid json`),
	}

	err := sarif.Write("/tmp/should-not-exist.sarif", docs)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "merge SARIF")
}

func TestWrite_MkdirAllErrorReturnsWrapped(t *testing.T) {
	err := sarif.Write("/dev/null/impossible/report.sarif", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create directory")
}
