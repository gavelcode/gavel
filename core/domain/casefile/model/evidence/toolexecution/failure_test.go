package toolexecution_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
)

func TestNewFailure(t *testing.T) {
	failed, err := toolexecution.NewFailure("golangci-lint", "exit 2: invalid config")

	require.NoError(t, err)
	assert.Equal(t, "golangci-lint", failed.Tool())
	assert.Equal(t, "exit 2: invalid config", failed.Reason())
}

func TestNewFailureValidation(t *testing.T) {
	tests := []struct {
		name   string
		tool   string
		reason string
	}{
		{name: "shouldRejectEmptyTool", tool: "   ", reason: "boom"},
		{name: "shouldRejectEmptyReason", tool: "ruff", reason: ""},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := toolexecution.NewFailure(tcase.tool, tcase.reason)

			require.Error(t, err)
			assert.ErrorIs(t, err, toolexecution.ErrInvalidFailure)
		})
	}
}
