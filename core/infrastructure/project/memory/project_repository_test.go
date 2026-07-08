package memory_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/infrastructure/project/memory"
)

func TestProjectRepositorySaveAndFindByID(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "backend", "backend", "//backend/...")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, testTenantID, project.ID())
	require.NoError(t, err)
	assert.Equal(t, project.ID(), found.ID())
	assert.Equal(t, "backend", found.Name())
}

func TestProjectRepositorySaveAndFindByName(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "backend", "backend", "//backend/...")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByName(ctx, testTenantID, "backend")
	require.NoError(t, err)
	assert.Equal(t, project.ID(), found.ID())
}

func TestProjectRepositoryFindByIDNotFound(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()
	id := projectmodel.NewProjectID(uuid.New())
	_, err := repo.FindByID(ctx, testTenantID, id)
	assert.ErrorIs(t, err, memory.ErrProjectNotFound)
}

func TestProjectRepositorySaveAndFindByKey(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "backend", "backend", "//backend/...")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByKey(ctx, testTenantID, "backend")
	require.NoError(t, err)
	assert.Equal(t, project.ID(), found.ID())
}

func TestProjectRepositoryFindByKeyNotFound(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()

	_, err := repo.FindByKey(ctx, testTenantID, "nonexistent")
	assert.ErrorIs(t, err, memory.ErrProjectNotFound)
}

func TestProjectRepositoryFindByNameNotFound(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()

	_, err := repo.FindByName(ctx, testTenantID, "nonexistent")
	assert.ErrorIs(t, err, memory.ErrProjectNotFound)
}

func TestProjectRepositoryBaselinePersistsToDisk(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)
	repo := memory.NewProjectRepositoryWithBaseline(store)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "core", "core", "//core/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1", "fp-2"}, []string{"arch-1"}, nil, nil)

	require.NoError(t, repo.Save(ctx, project))

	repo2 := memory.NewProjectRepositoryWithBaseline(store)
	project2, err := projectmodel.NewProject(testTenantID, "core", "core", "//core/...")
	require.NoError(t, err)
	require.NoError(t, repo2.Save(ctx, project2))

	found, err := repo2.FindByName(ctx, testTenantID, "core")
	require.NoError(t, err)
	baseline := found.Baseline("main")
	assert.True(t, baseline.HasPrevious(), "baseline should persist across repo instances")
	assert.Equal(t, []string{"fp-1", "fp-2"}, baseline.Fingerprints())
	assert.Equal(t, []string{"arch-1"}, baseline.ArchIDs())
}

func TestProjectRepositoryReconstitutesFileCoverageFromDisk(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)
	repo := memory.NewProjectRepositoryWithBaseline(store)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "core", "core", "//core/...")
	require.NoError(t, err)
	pct := 75.0
	entry, err := projectmodel.NewFileCoverageEntry("core/a.go", []int{1, 2, 3}, []int{4})
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1"}, nil, &pct, []projectmodel.FileCoverageEntry{entry})
	require.NoError(t, repo.Save(ctx, project))

	repo2 := memory.NewProjectRepositoryWithBaseline(store)
	project2, err := projectmodel.NewProject(testTenantID, "core", "core", "//core/...")
	require.NoError(t, err)
	require.NoError(t, repo2.Save(ctx, project2))

	found, err := repo2.FindByName(ctx, testTenantID, "core")
	require.NoError(t, err)
	baseline := found.Baseline("main")
	require.Len(t, baseline.FileCoverage(), 1, "per-file coverage must survive the disk round-trip")
	assert.Equal(t, "core/a.go", baseline.FileCoverage()[0].FilePath())
	assert.Equal(t, []int{1, 2, 3}, baseline.FileCoverage()[0].Covered())
	assert.Equal(t, []int{4}, baseline.FileCoverage()[0].Uncovered())
}

func TestProjectRepositoryWithoutBaselineStoreDoesNotPersist(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "core", "core", "//core/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1"}, nil, nil, nil)

	require.NoError(t, repo.Save(ctx, project))

	repo2 := memory.NewProjectRepository()
	project2, err := projectmodel.NewProject(testTenantID, "core", "core", "//core/...")
	require.NoError(t, err)
	require.NoError(t, repo2.Save(ctx, project2))

	found, err := repo2.FindByName(ctx, testTenantID, "core")
	require.NoError(t, err)
	baseline := found.Baseline("main")
	assert.False(t, baseline.HasPrevious(), "without BaselineStore, baseline is lost between instances")
}

func TestProjectRepositorySaveOverwrites(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "backend", "backend", "//backend/...")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, testTenantID, project.ID())
	require.NoError(t, err)
	assert.Equal(t, "//backend/...", found.TargetPattern())

	require.NoError(t, repo.Save(ctx, project))

	found2, err := repo.FindByID(ctx, testTenantID, project.ID())
	require.NoError(t, err)
	assert.Equal(t, project.ID(), found2.ID())
}

func TestProjectRepositorySetBaselineStore(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)
	repo := memory.NewProjectRepository()
	ctx := context.Background()

	repo.SetBaselineStore(store)

	project, err := projectmodel.NewProject(testTenantID, "core", "core", "//core/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1"}, nil, nil, nil)

	require.NoError(t, repo.Save(ctx, project))

	repo2 := memory.NewProjectRepositoryWithBaseline(store)
	project2, err := projectmodel.NewProject(testTenantID, "core", "core", "//core/...")
	require.NoError(t, err)
	require.NoError(t, repo2.Save(ctx, project2))

	found, err := repo2.FindByName(ctx, testTenantID, "core")
	require.NoError(t, err)
	baseline := found.Baseline("main")
	assert.True(t, baseline.HasPrevious())
	assert.Equal(t, []string{"fp-1"}, baseline.Fingerprints())
}

func TestProjectRepositorySaveReturnsErrorWhenBaselineStoreFails(t *testing.T) {
	workspace := t.TempDir()
	store := memory.NewBaselineStore(workspace)
	repo := memory.NewProjectRepositoryWithBaseline(store)
	ctx := context.Background()

	dir := filepath.Join(workspace, ".gavel", "baseline", "core")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.Chmod(dir, 0o444))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	project, err := projectmodel.NewProject(testTenantID, "core", "core", "//core/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1"}, nil, nil, nil)

	err = repo.Save(ctx, project)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save baselines")
}
