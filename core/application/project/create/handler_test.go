package create_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/project/create"
)

func TestHandlerExecuteSuccessful(t *testing.T) {
	projects := newFakeProjectRepo()
	handler := create.NewHandler(projects)

	cmd, err := create.NewCommand(testTenant.String(), "backend", "Backend Service", "//backend/...")
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.ProjectID)
	assert.Equal(t, 1, projects.count())
}

func TestNewCommandRejectsEmptyTargetPattern(t *testing.T) {
	_, err := create.NewCommand(testTenant.String(), "backend", "Backend Service", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, create.ErrInvalidCommand)
}

func TestNewCommandRejectsWhitespaceTargetPattern(t *testing.T) {
	_, err := create.NewCommand(testTenant.String(), "backend", "Backend Service", "   ")
	require.Error(t, err)
	assert.ErrorIs(t, err, create.ErrInvalidCommand)
}

func TestHandlerExecuteSaveError(t *testing.T) {
	projects := newFakeProjectRepo()
	projects.saveErr = errors.New("disk full")
	handler := create.NewHandler(projects)

	cmd, err := create.NewCommand(testTenant.String(), "backend", "Backend Service", "//backend/...")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteNewProjectValidationError(t *testing.T) {
	projects := newFakeProjectRepo()
	handler := create.NewHandler(projects)

	cmd, err := create.NewCommand(testTenant.String(), "UPPER_CASE", "Backend", "//backend/...")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Equal(t, 0, projects.count())
}

func TestNewHandlerRejectsNilRepo(t *testing.T) {
	assert.Panics(t, func() { create.NewHandler(nil) })
}
