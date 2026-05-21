package sourceblob_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sourceblob "github.com/usegavel/gavel/core/infrastructure/platform/sourceblob"
)

func TestSourceBlobRepoSaveAndFetch(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := sourceblob.NewStorage(db)
	ctx := context.Background()

	content := []byte("package main\n\nfunc main() {}\n")
	require.NoError(t, repo.Save(ctx, project.ID().String(), "abc123", "main.go", content, "text/plain; charset=utf-8"))

	got, ct, err := repo.Fetch(ctx, project.ID().String(), "abc123", "main.go")
	require.NoError(t, err)
	assert.Equal(t, content, got)
	assert.Equal(t, "text/plain; charset=utf-8", ct)
}

func TestSourceBlobRepoFetchMissingReturnsNotFound(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := sourceblob.NewStorage(db)

	_, _, err := repo.Fetch(context.Background(), project.ID().String(), "missing-sha", "missing.go")
	assert.True(t, errors.Is(err, sourceblob.ErrNotFound), "expected ErrSourceBlobNotFound, got %v", err)
}

func TestSourceBlobRepoCommitIsolation(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := sourceblob.NewStorage(db)
	ctx := context.Background()

	firstVersion := []byte("// v1")
	secondVersion := []byte("// v2 — different content")
	require.NoError(t, repo.Save(ctx, project.ID().String(), "sha-old", "main.go", firstVersion, "text/plain"))
	require.NoError(t, repo.Save(ctx, project.ID().String(), "sha-new", "main.go", secondVersion, "text/plain"))

	gotOld, _, err := repo.Fetch(ctx, project.ID().String(), "sha-old", "main.go")
	require.NoError(t, err)
	gotNew, _, err := repo.Fetch(ctx, project.ID().String(), "sha-new", "main.go")
	require.NoError(t, err)

	assert.Equal(t, firstVersion, gotOld)
	assert.Equal(t, secondVersion, gotNew)
}

func TestSourceBlobRepoSaveIsIdempotent(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := sourceblob.NewStorage(db)
	ctx := context.Background()

	content := []byte("// first write")
	require.NoError(t, repo.Save(ctx, project.ID().String(), "abc", "x.go", content, "text/plain"))
	require.NoError(t, repo.Save(ctx, project.ID().String(), "abc", "x.go", []byte("// second write"), "text/plain"))

	got, _, err := repo.Fetch(ctx, project.ID().String(), "abc", "x.go")
	require.NoError(t, err)
	assert.Equal(t, content, got, "first write wins; subsequent writes are idempotent no-ops")
}

func TestSourceBlobRepoSaveCancelledContext(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := sourceblob.NewStorage(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Save(ctx, project.ID().String(), "abc", "main.go", []byte("content"), "text/plain")
	assert.Error(t, err)
}

func TestSourceBlobRepoFetchCancelledContext(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := sourceblob.NewStorage(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := repo.Fetch(ctx, project.ID().String(), "abc", "main.go")
	assert.Error(t, err)
}
