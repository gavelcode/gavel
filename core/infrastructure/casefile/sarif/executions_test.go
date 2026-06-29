package sarif

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseToolExecutionsFailure(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"golangci-lint"}},"results":[],"invocations":[{"executionSuccessful":false,"toolExecutionNotifications":[{"message":{"text":"56 compile errors prevented analysis"}}]}]}]}`)

	failures, err := NewParser().ParseToolExecutions(data)

	require.NoError(t, err)
	require.Len(t, failures, 1)
	assert.Equal(t, "golangci-lint", failures[0].Tool)
	assert.Contains(t, failures[0].Reason, "56 compile errors")
}

func TestParseToolExecutionsSuccess(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"ruff"}},"results":[],"invocations":[{"executionSuccessful":true}]}]}`)

	failures, err := NewParser().ParseToolExecutions(data)

	require.NoError(t, err)
	assert.Empty(t, failures)
}

func TestParseToolExecutionsNoInvocations(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"pmd"}},"results":[]}]}`)

	failures, err := NewParser().ParseToolExecutions(data)

	require.NoError(t, err)
	assert.Empty(t, failures)
}

func TestParseToolExecutionsFailureWithoutNotifications(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"eslint"}},"results":[],"invocations":[{"executionSuccessful":false}]}]}`)

	failures, err := NewParser().ParseToolExecutions(data)

	require.NoError(t, err)
	require.Len(t, failures, 1)
	assert.NotEmpty(t, failures[0].Reason)
}

func TestParseToolExecutionsEmptyData(t *testing.T) {
	failures, err := NewParser().ParseToolExecutions(nil)

	require.NoError(t, err)
	assert.Empty(t, failures)
}

func TestParseToolExecutionsInvalidJSON(t *testing.T) {
	failures, err := NewParser().ParseToolExecutions([]byte(`{not valid sarif`))

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDecodeSARIF)
	assert.Nil(t, failures)
}

func TestParseToolExecutionsMissingToolNameFallsBackToUnknown(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"  "}},"results":[],"invocations":[{"executionSuccessful":false,"toolExecutionNotifications":[{"message":{"text":"boom"}}]}]}]}`)

	failures, err := NewParser().ParseToolExecutions(data)

	require.NoError(t, err)
	require.Len(t, failures, 1)
	assert.Equal(t, unknownTool, failures[0].Tool)
}

func TestParseToolExecutionsBlankNotificationTextFallsBackToDefault(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"pmd"}},"results":[],"invocations":[{"executionSuccessful":false,"toolExecutionNotifications":[{"message":{"text":"   "}}]}]}]}`)

	failures, err := NewParser().ParseToolExecutions(data)

	require.NoError(t, err)
	require.Len(t, failures, 1)
	assert.Equal(t, "analyzer reported an unsuccessful run", failures[0].Reason)
}
