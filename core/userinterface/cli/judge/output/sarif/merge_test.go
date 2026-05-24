package sarif

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeMultipleReports(t *testing.T) {
	doc1 := []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"tool-a"}},"results":[]}]}`)
	doc2 := []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"tool-b"}},"results":[]}]}`)

	merged, err := merge([][]byte{doc1, doc2})

	require.NoError(t, err)
	var got struct {
		Schema  string            `json:"$schema"`
		Version string            `json:"version"`
		Runs    []json.RawMessage `json:"runs"`
	}
	require.NoError(t, json.Unmarshal(merged, &got))
	assert.Equal(t, "2.1.0", got.Version)
	assert.NotEmpty(t, got.Schema)
	assert.Len(t, got.Runs, 2)
}

func TestMergeEmptyInput(t *testing.T) {
	merged, err := merge(nil)

	require.NoError(t, err)
	var got struct {
		Runs []json.RawMessage `json:"runs"`
	}
	require.NoError(t, json.Unmarshal(merged, &got))
	assert.Empty(t, got.Runs)
}

func TestMergeMultipleRunsInOneDoc(t *testing.T) {
	doc := []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"a"}},"results":[]},{"tool":{"driver":{"name":"b"}},"results":[]}]}`)

	merged, err := merge([][]byte{doc})

	require.NoError(t, err)
	var got struct {
		Runs []json.RawMessage `json:"runs"`
	}
	require.NoError(t, json.Unmarshal(merged, &got))
	assert.Len(t, got.Runs, 2)
}

func TestMergeInvalidJSONReturnsError(t *testing.T) {
	_, err := merge([][]byte{[]byte(`not json`)})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse SARIF document")
}
