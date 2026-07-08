package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/pleading/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	pleadingpostgres "github.com/usegavel/gavel/core/infrastructure/pleading/postgres"
)

func newPleading(t *testing.T, projectID projectmodel.ProjectID, number int, commitSHA string) model.Pleading {
	t.Helper()
	p, err := model.FilePleading(testTenantID, projectID, number, "title", "alice", "feature", "main", commitSHA)
	require.NoError(t, err)
	return p
}

func TestPleadingRepoSaveAndFindByID(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := pleadingpostgres.NewRepository(db)
	ctx := context.Background()

	pleading := newPleading(t, project.ID(), 42, "abc123")
	require.NoError(t, repo.Save(ctx, pleading))

	found, err := repo.FindByID(ctx, testTenantID, pleading.ID())
	require.NoError(t, err)

	assert.True(t, pleading.ID().Equal(found.ID()))
	assert.True(t, project.ID().Equal(found.ProjectID()))
	assert.Equal(t, 42, found.Number())
	assert.Equal(t, "title", found.Title())
	assert.Equal(t, "alice", found.Petitioner())
	assert.Equal(t, "feature", found.SourceBranch())
	assert.Equal(t, "main", found.TargetBranch())
	assert.Equal(t, "abc123", found.CommitSHA())
	assert.True(t, model.StatusOpen.Equal(found.Status()))
}

func TestPleadingRepoFindByIDNotFound(t *testing.T) {
	db := setupDB(t)
	repo := pleadingpostgres.NewRepository(db)

	id := model.NewPleadingID(uuid.New())

	_, err := repo.FindByID(context.Background(), testTenantID, id)
	require.Error(t, err)
	assert.Equal(t, failure.NotFound, failure.Of(err))
}

func TestPleadingRepoSaveUpdatesByID(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := pleadingpostgres.NewRepository(db)
	ctx := context.Background()

	pleading := newPleading(t, project.ID(), 1, "sha1")
	require.NoError(t, repo.Save(ctx, pleading))

	require.NoError(t, pleading.MarkMerged(time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, pleading))

	found, err := repo.FindByID(ctx, testTenantID, pleading.ID())
	require.NoError(t, err)
	assert.True(t, model.StatusMerged.Equal(found.Status()), "resolve mutation persisted via UPDATE-by-id")
}

func TestPleadingRepoUpsertsByNaturalKey(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := pleadingpostgres.NewRepository(db)
	ctx := context.Background()

	first := newPleading(t, project.ID(), 99, "sha-old")
	require.NoError(t, repo.Save(ctx, first))

	second, err := model.FilePleading(testTenantID, project.ID(), 99, "updated title", "bob", "feature-v2", "main", "sha-new")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, second))

	stored, err := repo.FindByID(ctx, testTenantID, first.ID())
	require.NoError(t, err)
	assert.Equal(t, "updated title", stored.Title())
	assert.Equal(t, "bob", stored.Petitioner())
	assert.Equal(t, "feature-v2", stored.SourceBranch())
	assert.Equal(t, "sha-new", stored.CommitSHA())
}

func TestPleadingRepoNaturalKeyConflictPreservesStatus(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := pleadingpostgres.NewRepository(db)
	ctx := context.Background()

	first := newPleading(t, project.ID(), 7, "sha1")
	require.NoError(t, repo.Save(ctx, first))
	require.NoError(t, first.MarkMerged(time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, first))

	refile, err := model.FilePleading(testTenantID, project.ID(), 7, "still merged?", "alice", "feature", "main", "sha2")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, refile))

	stored, err := repo.FindByID(ctx, testTenantID, first.ID())
	require.NoError(t, err)
	assert.True(t, model.StatusMerged.Equal(stored.Status()),
		"refile via natural-key upsert must not reset a terminal status")
	assert.Equal(t, "still merged?", stored.Title(), "non-status fields are refreshed by refile")
}

func TestPleadingRepoSaveUpdateReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := pleadingpostgres.NewRepository(testDB)
	ctx := context.Background()

	pleading := newPleading(t, project.ID(), 1, "sha1")
	require.NoError(t, repo.Save(ctx, pleading))

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Save(cancelledCtx, pleading)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update pleading")
}

func TestPleadingRepoSaveInsertReturnsErrorOnFKViolation(t *testing.T) {
	testDB := setupDB(t)
	repo := pleadingpostgres.NewRepository(testDB)

	fakeProjectID := projectmodel.NewProjectID(uuid.New())
	pleading := newPleading(t, fakeProjectID, 1, "sha1")

	err := repo.Save(context.Background(), pleading)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert pleading")
}

func TestPleadingRepoFindByIDReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := pleadingpostgres.NewRepository(testDB)
	ctx := context.Background()

	pleading := newPleading(t, project.ID(), 1, "sha1")
	require.NoError(t, repo.Save(ctx, pleading))

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindByID(cancelledCtx, testTenantID, pleading.ID())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scan pleading")
}

func TestPleadingRepoFindByIDReturnsErrorOnCorruptedStatus(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := pleadingpostgres.NewRepository(testDB)
	ctx := context.Background()

	pleading := newPleading(t, project.ID(), 1, "sha1")
	require.NoError(t, repo.Save(ctx, pleading))

	_, err := testDB.ExecContext(ctx,
		"UPDATE pleadings SET status = 'invalid_status' WHERE id = ?",
		pleading.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, testTenantID, pleading.ID())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reconstitute pleading status")
}
