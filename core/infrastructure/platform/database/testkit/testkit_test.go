package testkit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartPostgresContainerReturnsErrorOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := startPostgresContainer(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "container runtime unavailable")
}
