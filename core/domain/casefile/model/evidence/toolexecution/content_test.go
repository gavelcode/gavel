package toolexecution_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
)

func validFailure(t *testing.T, tool, reason string) toolexecution.Failure {
	t.Helper()
	failed, err := toolexecution.NewFailure(tool, reason)
	require.NoError(t, err)
	return failed
}

func TestNewToolExecutionContent(t *testing.T) {
	tests := []struct {
		name     string
		failures []toolexecution.Failure
	}{
		{name: "shouldCreateWithFailures", failures: []toolexecution.Failure{validFailure(t, "ruff", "boom")}},
		{name: "shouldCreateWithEmpty", failures: []toolexecution.Failure{}},
		{name: "shouldCreateWithNil", failures: nil},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			content, err := toolexecution.NewContent(tcase.failures)

			require.NoError(t, err)
			assert.Equal(t, len(tcase.failures), len(content.Failures()))
		})
	}
}

func TestToolExecutionContentTypeAndSubtype(t *testing.T) {
	content, err := toolexecution.NewContent(nil)
	require.NoError(t, err)

	assert.Equal(t, evidence.TypeAnalysis, content.Type())
	assert.Equal(t, evidence.SubtypeToolExecution, content.Subtype())
}

func TestToolExecutionContentMerge(t *testing.T) {
	c1, err := toolexecution.NewContent([]toolexecution.Failure{validFailure(t, "ruff", "a")})
	require.NoError(t, err)
	c2, err := toolexecution.NewContent([]toolexecution.Failure{validFailure(t, "pmd", "b")})
	require.NoError(t, err)

	merged, err := c1.Merge(c2)
	require.NoError(t, err)

	mergedContent, ok := merged.(toolexecution.Content)
	require.True(t, ok)
	assert.Equal(t, 2, len(mergedContent.Failures()))
}

func TestToolExecutionContentDefensiveCopy(t *testing.T) {
	original := []toolexecution.Failure{validFailure(t, "ruff", "a")}
	other := validFailure(t, "pmd", "b")

	content, err := toolexecution.NewContent(original)
	require.NoError(t, err)

	original[0] = other
	assert.NotEqual(t, other, content.Failures()[0])

	returned := content.Failures()
	returned[0] = other
	assert.NotEqual(t, other, content.Failures()[0])
}
