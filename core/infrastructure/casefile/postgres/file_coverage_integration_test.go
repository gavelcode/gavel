package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
)

func TestFileCoverageStore_SaveAndFetch(t *testing.T) {
	database := setupDB(t)
	project := insertTestProject(t, database)
	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	repo := postgres.NewRepository(database)
	require.NoError(t, repo.Save(context.Background(), caseFile))

	store := postgres.NewFileCoverageStore(database)
	ctx := context.Background()

	entries := []evidencedto.FileCoverage{
		{FilePath: "src/main.go", Covered: []int{1, 3, 5}, Uncovered: []int{2, 4}},
		{FilePath: "src/util.go", Covered: []int{10, 12, 15}, Uncovered: []int{11, 13}},
	}

	err := store.Save(ctx, caseFile.ID().String(), entries)
	require.NoError(t, err)

	got, err := store.Fetch(ctx, caseFile.ID().String(), "src/main.go")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "src/main.go", got.FilePath)
	assert.Equal(t, []int{1, 3, 5}, got.Covered)
	assert.Equal(t, []int{2, 4}, got.Uncovered)

	got2, err := store.Fetch(ctx, caseFile.ID().String(), "src/util.go")
	require.NoError(t, err)
	require.NotNil(t, got2)
	assert.Equal(t, []int{10, 12, 15}, got2.Covered)
	assert.Equal(t, []int{11, 13}, got2.Uncovered)
}

func TestFileCoverageStore_FetchNonExistent(t *testing.T) {
	database := setupDB(t)
	project := insertTestProject(t, database)
	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	repo := postgres.NewRepository(database)
	require.NoError(t, repo.Save(context.Background(), caseFile))

	store := postgres.NewFileCoverageStore(database)
	ctx := context.Background()

	got, err := store.Fetch(ctx, caseFile.ID().String(), "nonexistent.go")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestFileCoverageStore_SaveIdempotent(t *testing.T) {
	database := setupDB(t)
	project := insertTestProject(t, database)
	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	repo := postgres.NewRepository(database)
	require.NoError(t, repo.Save(context.Background(), caseFile))

	store := postgres.NewFileCoverageStore(database)
	ctx := context.Background()

	initial := []evidencedto.FileCoverage{
		{FilePath: "src/main.go", Covered: []int{1, 3}, Uncovered: []int{2}},
	}
	require.NoError(t, store.Save(ctx, caseFile.ID().String(), initial))

	updated := []evidencedto.FileCoverage{
		{FilePath: "src/main.go", Covered: []int{1, 2, 3}, Uncovered: nil},
	}
	require.NoError(t, store.Save(ctx, caseFile.ID().String(), updated))

	got, err := store.Fetch(ctx, caseFile.ID().String(), "src/main.go")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []int{1, 2, 3}, got.Covered)
	assert.Empty(t, got.Uncovered)
}

func TestFileCoverageStore_SaveEmptySlice(t *testing.T) {
	db := setupDB(t)
	store := postgres.NewFileCoverageStore(db)

	err := store.Save(context.Background(), "00000000-0000-0000-0000-000000000000", nil)
	require.NoError(t, err)
}
