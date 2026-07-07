package removeproject_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/gavelspace/removeproject"
	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func TestHandlerExecuteSuccessful(t *testing.T) {
	gavelspaces := newFakeGavelspaceRepo()
	projectID := projectmodel.NewProjectID(uuid.New())
	gavelspace := seededGavelspace(t, "monorepo", projectID, "//svc/...")
	gavelspaces.seed(gavelspace)

	handler := removeproject.NewHandler(gavelspaces)
	cmd := mustCommand(t, "monorepo", projectID.String())

	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	persisted, err := gavelspaces.FindByName(context.Background(), testTenantID, gavelspace.ID())
	require.NoError(t, err)
	assert.Empty(t, persisted.Projects(), "project ref removed from gavelspace")
}

func TestHandlerExecuteGavelspaceNotFound(t *testing.T) {
	gavelspaces := newFakeGavelspaceRepo()

	handler := removeproject.NewHandler(gavelspaces)
	cmd := mustCommand(t, "missing", "proj-1")

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteProjectNotInGavelspace(t *testing.T) {
	gavelspaces := newFakeGavelspaceRepo()
	gavelspace, err := gsmodel.NewGavelspace(testTenantID, "monorepo")
	require.NoError(t, err)
	gavelspaces.seed(gavelspace)

	handler := removeproject.NewHandler(gavelspaces)
	cmd := mustCommand(t, "monorepo", uuid.NewString())

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, gsmodel.ErrProjectNotFound)
}

func TestHandlerExecuteSaveErrorPropagated(t *testing.T) {
	gavelspaces := newFakeGavelspaceRepo()
	projectID := projectmodel.NewProjectID(uuid.New())
	gavelspace := seededGavelspace(t, "monorepo", projectID, "//svc/...")
	gavelspaces.seed(gavelspace)
	gavelspaces.saveErr = errors.New("disk full")

	handler := removeproject.NewHandler(gavelspaces)
	cmd := mustCommand(t, "monorepo", projectID.String())

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteInvalidProjectIDRejected(t *testing.T) {
	gavelspaces := newFakeGavelspaceRepo()
	gavelspace, err := gsmodel.NewGavelspace(testTenantID, "monorepo")
	require.NoError(t, err)
	gavelspaces.seed(gavelspace)

	handler := removeproject.NewHandler(gavelspaces)

	cmd, err := removeproject.NewCommand(testTenant, "monorepo", "valid-id-format")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err, "non-existent project rejected by domain")
}

func TestHandlerExecuteDrainsRemovedEvent(t *testing.T) {
	gavelspaces := newFakeGavelspaceRepo()
	projectID := projectmodel.NewProjectID(uuid.New())
	gavelspace := seededGavelspace(t, "monorepo", projectID, "//svc/...")
	gavelspace.ClearEvents()
	gavelspaces.seed(gavelspace)

	handler := removeproject.NewHandler(gavelspaces)
	result, err := handler.Execute(context.Background(), mustCommand(t, "monorepo", projectID.String()))
	require.NoError(t, err)

	require.Len(t, result.Events, 1, "ProjectRemoved event drained to caller")
	assert.Equal(t, gsmodel.EventNameProjectRemoved, result.Events[0].Name)

	persisted, err := gavelspaces.FindByName(context.Background(), testTenantID, gavelspace.ID())
	require.NoError(t, err)
	assert.Empty(t, persisted.Events(), "events drained before persistence; not retained")
}

func TestNewHandlerRejectsNilRepo(t *testing.T) {
	assert.Panics(t, func() { removeproject.NewHandler(nil) })
}

func mustCommand(t *testing.T, gavelspaceName, projectID string) removeproject.Command {
	t.Helper()
	cmd, err := removeproject.NewCommand(testTenant, gavelspaceName, projectID)
	require.NoError(t, err)
	return cmd
}

func seededGavelspace(t *testing.T, name string, projectID projectmodel.ProjectID, targetPattern string) gsmodel.Gavelspace {
	t.Helper()
	gavelspace, err := gsmodel.NewGavelspace(testTenantID, name)
	require.NoError(t, err)
	ref, err := gsmodel.NewProjectRef(projectID, targetPattern)
	require.NoError(t, err)
	require.NoError(t, gavelspace.AddProject(ref, time.Now().UTC()))
	return gavelspace
}

func TestHandlerExecuteInvalidTenant(t *testing.T) {
	gavelspaces := newFakeGavelspaceRepo()
	handler := removeproject.NewHandler(gavelspaces)
	cmd, err := removeproject.NewCommand("not-a-uuid", "monorepo", uuid.NewString())
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant id")
}
