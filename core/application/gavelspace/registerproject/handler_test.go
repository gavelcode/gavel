package registerproject_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/usegavel/gavel/core/application/gavelspace/registerproject"
	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func TestHandlerRegistersExistingProject(t *testing.T) {
	repo := newFakeGavelspaceRepo()
	gs, err := gsmodel.NewGavelspace(testTenantID, "alpha")
	require.NoError(t, err)
	gs.ClearEvents()
	repo.seed(gs)

	handler := registerproject.NewHandler(repo)
	id := projectmodel.NewProjectID(uuid.New())

	cmd, err := registerproject.NewCommand(testTenant, "alpha", id.String(), "//svc/...")
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Events, "ProjectAdded event drained to caller")

	name, err := gsmodel.NewGavelspaceID("alpha")
	require.NoError(t, err)
	loaded, err := repo.FindByName(context.Background(), testTenantID, name)
	require.NoError(t, err)
	require.Len(t, loaded.Projects(), 1, "project added to gavelspace")
}

func TestHandlerGavelspaceNotFound(t *testing.T) {
	repo := newFakeGavelspaceRepo()
	handler := registerproject.NewHandler(repo)
	id := projectmodel.NewProjectID(uuid.New())
	cmd, err := registerproject.NewCommand(testTenant, uuid.NewString(), id.String(), "//svc/...")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestNewHandlerPanicsOnNilRepo(t *testing.T) {
	assert.Panics(t, func() { registerproject.NewHandler(nil) })
}

func TestHandlerExecuteInvalidProjectID(t *testing.T) {
	repo := newFakeGavelspaceRepo()
	gs, err := gsmodel.NewGavelspace(testTenantID, "alpha")
	require.NoError(t, err)
	repo.seed(gs)

	handler := registerproject.NewHandler(repo)
	cmd, err := registerproject.NewCommand(testTenant, "alpha", "not-a-uuid", "//svc/...")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project id")
}

func TestHandlerExecuteSaveError(t *testing.T) {
	repo := newFakeGavelspaceRepo()
	gs, err := gsmodel.NewGavelspace(testTenantID, "alpha")
	require.NoError(t, err)
	gs.ClearEvents()
	repo.seed(gs)
	repo.saveErr = errors.New("disk full")

	handler := registerproject.NewHandler(repo)
	id := projectmodel.NewProjectID(uuid.New())
	cmd, err := registerproject.NewCommand(testTenant, "alpha", id.String(), "//svc/...")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save gavelspace")
}

func TestHandlerExecuteInvalidTenant(t *testing.T) {
	repo := newFakeGavelspaceRepo()
	handler := registerproject.NewHandler(repo)
	id := projectmodel.NewProjectID(uuid.New())
	cmd, err := registerproject.NewCommand("not-a-uuid", "alpha", id.String(), "//svc/...")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant id")
}
