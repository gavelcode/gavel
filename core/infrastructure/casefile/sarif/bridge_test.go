package sarif

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validBridgeSARIF = `{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"todo"}},"results":[
 {"ruleId":"todo:found","level":"warning","message":{"text":"TODO"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"pkg/a.sh"},"region":{"startLine":3}}}]}]}]}`

func bridgeResult(mutate string) []byte {
	return []byte(fmt.Sprintf(`{"runs":[{"tool":{"driver":{"name":"todo"}},"results":[%s]}]}`, mutate))
}

func TestValidateBridgeSubmissionAcceptsConformingDocument(t *testing.T) {
	assert.Empty(t, ValidateBridgeSubmission([]byte(validBridgeSARIF)))
}

func TestValidateBridgeSubmissionAcceptsEmptyResults(t *testing.T) {
	assert.Empty(t, ValidateBridgeSubmission(bridgeResult("")))
}

func TestValidateBridgeSubmissionProblems(t *testing.T) {
	loc := func(uri string, line int) string {
		return fmt.Sprintf(`"locations":[{"physicalLocation":{"artifactLocation":{"uri":%q},"region":{"startLine":%d}}}]`, uri, line)
	}
	tests := []struct {
		name    string
		data    []byte
		problem string
	}{
		{name: "shouldRejectInvalidJSON", data: []byte("{nope"), problem: "not valid SARIF JSON"},
		{name: "shouldRejectNoRuns", data: []byte(`{"runs":[]}`), problem: "at least one run"},
		{name: "shouldRejectEmptyDriverName", data: []byte(`{"runs":[{"tool":{"driver":{"name":" "}},"results":[]}]}`), problem: "tool.driver.name"},
		{name: "shouldRejectEmptyRuleID", data: bridgeResult(`{"ruleId":"","message":{"text":"m"},` + loc("a.sh", 1) + `}`), problem: "ruleId"},
		{name: "shouldRejectWhitespaceRuleID", data: bridgeResult(`{"ruleId":"  ","message":{"text":"m"},` + loc("a.sh", 1) + `}`), problem: "ruleId"},
		{name: "shouldRejectMissingRuleID", data: bridgeResult(`{"message":{"text":"m"},` + loc("a.sh", 1) + `}`), problem: "ruleId"},
		{name: "shouldRejectEmptyMessage", data: bridgeResult(`{"ruleId":"r","message":{"text":""},` + loc("a.sh", 1) + `}`), problem: "message.text"},
		{name: "shouldRejectMissingLocation", data: bridgeResult(`{"ruleId":"r","message":{"text":"m"}}`), problem: "locations[0]"},
		{name: "shouldRejectEmptyURI", data: bridgeResult(`{"ruleId":"r","message":{"text":"m"},` + loc("", 1) + `}`), problem: "uri must be non-empty"},
		{name: "shouldRejectSchemeURI", data: bridgeResult(`{"ruleId":"r","message":{"text":"m"},` + loc("file:///a.sh", 1) + `}`), problem: "workspace-relative"},
		{name: "shouldRejectAbsoluteURI", data: bridgeResult(`{"ruleId":"r","message":{"text":"m"},` + loc("/a.sh", 1) + `}`), problem: "workspace-relative"},
		{name: "shouldRejectParentURI", data: bridgeResult(`{"ruleId":"r","message":{"text":"m"},` + loc("a/../../b.sh", 1) + `}`), problem: "workspace-relative"},
		{name: "shouldRejectExecrootURI", data: bridgeResult(`{"ruleId":"r","message":{"text":"m"},` + loc("x/execroot/_main/a.sh", 1) + `}`), problem: "workspace-relative"},
		{name: "shouldRejectZeroStartLine", data: bridgeResult(`{"ruleId":"r","message":{"text":"m"},` + loc("a.sh", 0) + `}`), problem: "startLine"},
		{name: "shouldRejectUnknownLevel", data: bridgeResult(`{"ruleId":"r","level":"fatal","message":{"text":"m"},` + loc("a.sh", 1) + `}`), problem: "level"},
		{
			name:    "shouldRejectFailedInvocationWithoutExplanation",
			data:    []byte(`{"runs":[{"tool":{"driver":{"name":"t"}},"results":[],"invocations":[{"executionSuccessful":false}]}]}`),
			problem: "toolExecutionNotifications",
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			problems := ValidateBridgeSubmission(tcase.data)

			require.NotEmpty(t, problems)
			assert.Contains(t, problems[0], tcase.problem)
		})
	}
}

func TestValidateBridgeSubmissionAllowsExplainedFailedInvocation(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"t"}},"results":[],"invocations":[{"executionSuccessful":false,"toolExecutionNotifications":[{"message":{"text":"boom"}}]}]}]}`)

	assert.Empty(t, ValidateBridgeSubmission(data))
}

func TestTagRunsAddsPropertyBagAndKeepsContent(t *testing.T) {
	tagged, err := TagRuns([]byte(validBridgeSARIF), map[string]any{"provenance": "bridge", "bridge": "todo", "guaranteed": false})

	require.NoError(t, err)
	var doc struct {
		Version string `json:"version"`
		Runs    []struct {
			Properties struct {
				Gavel map[string]any `json:"gavel"`
			} `json:"properties"`
			Results []json.RawMessage `json:"results"`
		} `json:"runs"`
	}
	require.NoError(t, json.Unmarshal(tagged, &doc))
	assert.Equal(t, "2.1.0", doc.Version)
	require.Len(t, doc.Runs, 1)
	assert.Equal(t, "bridge", doc.Runs[0].Properties.Gavel["provenance"])
	assert.Equal(t, false, doc.Runs[0].Properties.Gavel["guaranteed"])
	assert.Len(t, doc.Runs[0].Results, 1)
}

func TestTagRunsPreservesExistingProperties(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"t"}},"properties":{"other":1}}]}`)

	tagged, err := TagRuns(data, map[string]any{"provenance": "bridge"})

	require.NoError(t, err)
	assert.Contains(t, string(tagged), `"other":1`)
	assert.Contains(t, string(tagged), `"provenance":"bridge"`)
}

func TestTagRunsRejectsInvalidJSON(t *testing.T) {
	_, err := TagRuns([]byte("{nope"), map[string]any{"a": 1})

	require.Error(t, err)
}
