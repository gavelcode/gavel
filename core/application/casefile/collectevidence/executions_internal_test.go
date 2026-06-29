package collectevidence

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

type fakeToolExecParser struct {
	failures []evidencedto.ToolFailure
}

func (f fakeToolExecParser) ParseToolExecutions(_ []byte) ([]evidencedto.ToolFailure, error) {
	return f.failures, nil
}

func TestAppendToolExecutionEvidence_BuildsEvidenceOnFailure(t *testing.T) {
	handler := &Handler{toolExecution: fakeToolExecParser{
		failures: []evidencedto.ToolFailure{{Tool: "golangci-lint", Reason: "compile errors"}},
	}}

	out := handler.appendToolExecutionEvidence(nil, []RawFile{{Data: []byte("{}")}})

	require.Len(t, out, 1)
	require.NotNil(t, out[0].ToolExecution)
	assert.Equal(t, "tool_execution", out[0].Subtype)
	require.Len(t, out[0].ToolExecution.Failures, 1)
	assert.Equal(t, "golangci-lint", out[0].ToolExecution.Failures[0].Tool)
}

func TestAppendToolExecutionEvidence_NoEvidenceWhenClean(t *testing.T) {
	handler := &Handler{toolExecution: fakeToolExecParser{failures: nil}}

	out := handler.appendToolExecutionEvidence([]evidencedto.Evidence{}, []RawFile{{Data: []byte("{}")}})

	assert.Empty(t, out)
}

func TestAppendToolExecutionEvidence_NoParserConfigured(t *testing.T) {
	handler := &Handler{}

	out := handler.appendToolExecutionEvidence([]evidencedto.Evidence{}, []RawFile{{Data: []byte("{}")}})

	assert.Empty(t, out)
}
