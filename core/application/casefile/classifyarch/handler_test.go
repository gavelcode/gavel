package classifyarch_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classifyarch"
)

func TestExecute_AllNew(t *testing.T) {
	handler := classifyarch.NewHandler()
	cmd := classifyarch.NewCommand([]string{"rule:a:b", "rule:c:d"}, nil)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 2, result.NewCount)
	assert.Equal(t, 0, result.FixedCount)
	assert.Equal(t, 0, result.ExistingCount)
	assert.True(t, result.NewIDs["rule:a:b"])
	assert.True(t, result.NewIDs["rule:c:d"])
}

func TestExecute_AllFixed(t *testing.T) {
	handler := classifyarch.NewHandler()
	cmd := classifyarch.NewCommand(nil, []string{"rule:a:b", "rule:c:d"})

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 0, result.NewCount)
	assert.Equal(t, 2, result.FixedCount)
	assert.Equal(t, 0, result.ExistingCount)
	assert.Empty(t, result.NewIDs)
}

func TestExecute_Mixed(t *testing.T) {
	handler := classifyarch.NewHandler()
	cmd := classifyarch.NewCommand(
		[]string{"rule:a:b", "rule:g:h"},
		[]string{"rule:a:b", "rule:c:d", "rule:e:f"},
	)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 1, result.NewCount)
	assert.Equal(t, 2, result.FixedCount)
	assert.Equal(t, 1, result.ExistingCount)
	assert.True(t, result.NewIDs["rule:g:h"])
	assert.False(t, result.NewIDs["rule:a:b"])
}

func TestExecute_NoChange(t *testing.T) {
	handler := classifyarch.NewHandler()
	cmd := classifyarch.NewCommand(
		[]string{"rule:a:b", "rule:c:d"},
		[]string{"rule:a:b", "rule:c:d"},
	)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 0, result.NewCount)
	assert.Equal(t, 0, result.FixedCount)
	assert.Equal(t, 2, result.ExistingCount)
}

func TestExecute_BothEmpty(t *testing.T) {
	handler := classifyarch.NewHandler()
	cmd := classifyarch.NewCommand(nil, nil)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 0, result.NewCount)
	assert.Equal(t, 0, result.FixedCount)
	assert.Equal(t, 0, result.ExistingCount)
}
