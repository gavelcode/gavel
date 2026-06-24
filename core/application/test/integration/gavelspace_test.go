package appintegration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gavelspacecreate "github.com/usegavel/gavel/core/application/gavelspace/create"
	"github.com/usegavel/gavel/core/application/gavelspace/registerproject"
	"github.com/usegavel/gavel/core/application/gavelspace/removeproject"
	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	memgavelspace "github.com/usegavel/gavel/core/infrastructure/gavelspace/memory"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
)

func mustCreateGavelspace(t *testing.T, repo *memgavelspace.GavelspaceRepository, name string) string {
	t.Helper()
	ctx := context.Background()

	handler := gavelspacecreate.NewHandler(repo)
	cmd, err := gavelspacecreate.NewCommand(name)
	require.NoError(t, err)

	result, err := handler.Execute(ctx, cmd)
	require.NoError(t, err)
	require.NotEmpty(t, result.Name)

	return result.Name
}

func TestGavelspaceLifecycle_CreateAndRetrieve(t *testing.T) {
	ctx := context.Background()
	repo := memgavelspace.NewGavelspaceRepository()

	handler := gavelspacecreate.NewHandler(repo)
	cmd, err := gavelspacecreate.NewCommand("my-monorepo")
	require.NoError(t, err)

	result, err := handler.Execute(ctx, cmd)
	require.NoError(t, err)
	require.NotEmpty(t, result.Name)
	assert.Equal(t, "my-monorepo", result.Name)

	gsID, err := gsmodel.NewGavelspaceID("my-monorepo")
	require.NoError(t, err)

	found, err := repo.FindByName(ctx, gsID)
	require.NoError(t, err)

	assert.Equal(t, "my-monorepo", found.ID().String())
	assert.Empty(t, found.Projects())
}

func TestGavelspaceLifecycle_RegisterProject(t *testing.T) {
	ctx := context.Background()
	gsRepo := memgavelspace.NewGavelspaceRepository()
	projRepo := memproject.NewProjectRepository()

	gsName := mustCreateGavelspace(t, gsRepo, "register-gs")
	projectID := mustCreateProject(t, projRepo, "register-proj")

	regHandler := registerproject.NewHandler(gsRepo)
	regCmd, err := registerproject.NewCommand(gsName, projectID, "//register-proj/...")
	require.NoError(t, err)

	_, err = regHandler.Execute(ctx, regCmd)
	require.NoError(t, err)

	gsID, err := gsmodel.NewGavelspaceID(gsName)
	require.NoError(t, err)

	found, err := gsRepo.FindByName(ctx, gsID)
	require.NoError(t, err)

	projects := found.Projects()
	require.Len(t, projects, 1)
	assert.Equal(t, projectID, projects[0].ID().String())
	assert.Equal(t, "//register-proj/...", projects[0].TargetPattern())
}

func TestGavelspaceLifecycle_RegisterAndRemoveProject(t *testing.T) {
	ctx := context.Background()
	gsRepo := memgavelspace.NewGavelspaceRepository()
	projRepo := memproject.NewProjectRepository()

	gsName := mustCreateGavelspace(t, gsRepo, "remove-gs")
	projectID := mustCreateProject(t, projRepo, "remove-proj")

	regHandler := registerproject.NewHandler(gsRepo)
	regCmd, err := registerproject.NewCommand(gsName, projectID, "//remove-proj/...")
	require.NoError(t, err)
	_, err = regHandler.Execute(ctx, regCmd)
	require.NoError(t, err)

	rmHandler := removeproject.NewHandler(gsRepo)
	rmCmd, err := removeproject.NewCommand(gsName, projectID)
	require.NoError(t, err)
	_, err = rmHandler.Execute(ctx, rmCmd)
	require.NoError(t, err)

	gsID, err := gsmodel.NewGavelspaceID(gsName)
	require.NoError(t, err)

	found, err := gsRepo.FindByName(ctx, gsID)
	require.NoError(t, err)

	assert.Empty(t, found.Projects())
}

func TestGavelspaceLifecycle_RejectDuplicateRegistration(t *testing.T) {
	ctx := context.Background()
	gsRepo := memgavelspace.NewGavelspaceRepository()
	projRepo := memproject.NewProjectRepository()

	gsName := mustCreateGavelspace(t, gsRepo, "dup-gs")
	projectID := mustCreateProject(t, projRepo, "dup-proj")

	regHandler := registerproject.NewHandler(gsRepo)

	regCmd, err := registerproject.NewCommand(gsName, projectID, "//dup-proj/...")
	require.NoError(t, err)
	_, err = regHandler.Execute(ctx, regCmd)
	require.NoError(t, err)

	dupCmd, err := registerproject.NewCommand(gsName, projectID, "//dup-proj/...")
	require.NoError(t, err)
	_, err = regHandler.Execute(ctx, dupCmd)
	require.Error(t, err, "registering the same target pattern twice must fail")
}
