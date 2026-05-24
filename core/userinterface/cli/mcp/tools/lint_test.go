package tools

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLintFiltering(t *testing.T) {
	input := []byte(`{"projects":[{"name":"core","findings":[
		{"tool":"golangci-lint","rule_id":"unused","severity":"warning","file_path":"main.go","line":10,"message":"unused var","fingerprint":"fp1"},
		{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"other.go","line":5,"message":"unchecked error","fingerprint":"fp2"},
		{"tool":"golangci-lint","rule_id":"govet","severity":"warning","file_path":"main.go","line":20,"message":"shadow","fingerprint":"fp3","status":"new"}
	]}]}`)

	var resp lintResponse
	require.NoError(t, json.Unmarshal(input, &resp))

	var matched []lintFinding
	for _, p := range resp.Projects {
		for _, f := range p.Findings {
			if f.FilePath == "main.go" {
				matched = append(matched, f)
			}
		}
	}

	require.Len(t, matched, 2)
	assert.Equal(t, 10, matched[0].Line)
	assert.Equal(t, 20, matched[1].Line)
	assert.Equal(t, "new", matched[1].Status)
}

func TestLintFiltering_NoMatches(t *testing.T) {
	input := []byte(`{"projects":[{"name":"core","findings":[
		{"tool":"golangci-lint","rule_id":"unused","severity":"warning","file_path":"other.go","line":10,"message":"x","fingerprint":"fp1"}
	]}]}`)

	var resp lintResponse
	require.NoError(t, json.Unmarshal(input, &resp))

	var matched []lintFinding
	for _, p := range resp.Projects {
		for _, f := range p.Findings {
			if f.FilePath == "main.go" {
				matched = append(matched, f)
			}
		}
	}

	assert.Empty(t, matched)
}
