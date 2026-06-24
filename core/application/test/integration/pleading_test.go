package appintegration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pleadingfile "github.com/usegavel/gavel/core/application/pleading/file"
	pleadingresolve "github.com/usegavel/gavel/core/application/pleading/resolve"
	projectcreate "github.com/usegavel/gavel/core/application/project/create"
	pleadingmodel "github.com/usegavel/gavel/core/domain/pleading/model"
	mempleading "github.com/usegavel/gavel/core/infrastructure/pleading/memory"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
)

func mustCreateProject(t *testing.T, projectRepo *memproject.ProjectRepository, key string) string {
	t.Helper()
	ctx := context.Background()

	handler := projectcreate.NewHandler(projectRepo)
	cmd, err := projectcreate.NewCommand(key, key+"-name", "//"+key+"/...")
	require.NoError(t, err)

	result, err := handler.Execute(ctx, cmd)
	require.NoError(t, err)
	require.NotEmpty(t, result.ProjectID)

	return result.ProjectID
}

func TestPleadingLifecycle_FileAndRetrieve(t *testing.T) {
	ctx := context.Background()
	projectRepo := memproject.NewProjectRepository()
	pleadingRepo := mempleading.NewPleadingRepository()

	projectID := mustCreateProject(t, projectRepo, "retrieve-project")

	fileHandler := pleadingfile.NewHandler(pleadingRepo)
	fileCmd, err := pleadingfile.NewCommand(
		projectID, 1, "Add login feature", "alice",
		"feature/login", "main", "abc123def456",
	)
	require.NoError(t, err)

	fileResult, err := fileHandler.Execute(ctx, fileCmd)
	require.NoError(t, err)
	require.NotEmpty(t, fileResult.PleadingID)
	assert.Equal(t, "open", fileResult.Status)

	pleadingID, err := pleadingmodel.ParsePleadingID(fileResult.PleadingID)
	require.NoError(t, err)

	stored, err := pleadingRepo.FindByID(ctx, pleadingID)
	require.NoError(t, err)

	assert.Equal(t, 1, stored.Number())
	assert.Equal(t, "Add login feature", stored.Title())
	assert.Equal(t, "alice", stored.Petitioner())
	assert.Equal(t, "feature/login", stored.SourceBranch())
	assert.Equal(t, "main", stored.TargetBranch())
	assert.Equal(t, "abc123def456", stored.CommitSHA())
	assert.Equal(t, pleadingmodel.StatusOpen, stored.Status())
}

func TestPleadingLifecycle_FileAndResolve_Merged(t *testing.T) {
	ctx := context.Background()
	projectRepo := memproject.NewProjectRepository()
	pleadingRepo := mempleading.NewPleadingRepository()

	projectID := mustCreateProject(t, projectRepo, "merge-project")

	fileHandler := pleadingfile.NewHandler(pleadingRepo)
	fileCmd, err := pleadingfile.NewCommand(
		projectID, 10, "Refactor auth", "bob",
		"refactor/auth", "main", "deadbeef",
	)
	require.NoError(t, err)

	fileResult, err := fileHandler.Execute(ctx, fileCmd)
	require.NoError(t, err)

	resolveHandler := pleadingresolve.NewHandler(pleadingRepo)
	resolveCmd, err := pleadingresolve.NewCommand(fileResult.PleadingID, "merged")
	require.NoError(t, err)

	resolveResult, err := resolveHandler.Execute(ctx, resolveCmd)
	require.NoError(t, err)
	assert.True(t, resolveResult.Changed)
	assert.Equal(t, "merged", resolveResult.Status)

	pleadingID, err := pleadingmodel.ParsePleadingID(fileResult.PleadingID)
	require.NoError(t, err)

	stored, err := pleadingRepo.FindByID(ctx, pleadingID)
	require.NoError(t, err)
	assert.Equal(t, pleadingmodel.StatusMerged, stored.Status())
}

func TestPleadingLifecycle_FileAndResolve_Closed(t *testing.T) {
	ctx := context.Background()
	projectRepo := memproject.NewProjectRepository()
	pleadingRepo := mempleading.NewPleadingRepository()

	projectID := mustCreateProject(t, projectRepo, "close-project")

	fileHandler := pleadingfile.NewHandler(pleadingRepo)
	fileCmd, err := pleadingfile.NewCommand(
		projectID, 5, "Experimental change", "carol",
		"experiment/new-ui", "develop", "cafe0123",
	)
	require.NoError(t, err)

	fileResult, err := fileHandler.Execute(ctx, fileCmd)
	require.NoError(t, err)

	resolveHandler := pleadingresolve.NewHandler(pleadingRepo)
	resolveCmd, err := pleadingresolve.NewCommand(fileResult.PleadingID, "closed")
	require.NoError(t, err)

	resolveResult, err := resolveHandler.Execute(ctx, resolveCmd)
	require.NoError(t, err)
	assert.True(t, resolveResult.Changed)
	assert.Equal(t, "closed", resolveResult.Status)

	pleadingID, err := pleadingmodel.ParsePleadingID(fileResult.PleadingID)
	require.NoError(t, err)

	stored, err := pleadingRepo.FindByID(ctx, pleadingID)
	require.NoError(t, err)
	assert.Equal(t, pleadingmodel.StatusClosed, stored.Status())
}

func TestPleadingLifecycle_ResolveNonExistent(t *testing.T) {
	ctx := context.Background()
	pleadingRepo := mempleading.NewPleadingRepository()

	resolveHandler := pleadingresolve.NewHandler(pleadingRepo)
	resolveCmd, err := pleadingresolve.NewCommand(uuid.NewString(), "merged")
	require.NoError(t, err)

	_, err = resolveHandler.Execute(ctx, resolveCmd)
	require.Error(t, err)
}
