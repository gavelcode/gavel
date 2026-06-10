package ops_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/ops"
)

func TestGetHealthReturnsOK(t *testing.T) {
	handler := ops.New()

	resp, err := handler.GetHealth(context.Background(), gen.GetHealthRequestObject{})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetHealth200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	assert.Equal(t, gen.Ok, jsonResp.Status)
}
