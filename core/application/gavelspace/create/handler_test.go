package create_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/gavelspace/create"
)

func TestHandlerExecuteSuccessful(t *testing.T) {
	repo := newFakeGavelspaceRepo()
	handler := create.NewHandler(repo)
	cmd := mustCommand(t, "monorepo")

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, "monorepo", result.Name)
	assert.Equal(t, 1, repo.count())
}

func TestHandlerExecuteSaveErrorPropagated(t *testing.T) {
	repo := newFakeGavelspaceRepo()
	repo.saveErr = errors.New("disk full")
	handler := create.NewHandler(repo)
	cmd := mustCommand(t, "monorepo")

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Equal(t, 0, repo.count())
}

func TestHandlerExecuteInvalidName(t *testing.T) {
	repo := newFakeGavelspaceRepo()
	handler := create.NewHandler(repo)

	_, err := handler.Execute(context.Background(), create.Command{})
	require.Error(t, err)
	assert.Equal(t, 0, repo.count())
}

func TestNewHandlerRejectsNilDependency(t *testing.T) {
	assert.Panics(t, func() { create.NewHandler(nil) })
}

func mustCommand(t *testing.T, name string) create.Command {
	t.Helper()
	cmd, err := create.NewCommand(name)
	require.NoError(t, err)
	return cmd
}
