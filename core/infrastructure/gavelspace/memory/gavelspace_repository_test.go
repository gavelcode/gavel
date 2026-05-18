package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/infrastructure/gavelspace/memory"
)

func TestGavelspaceRepositorySaveAndFindByName(t *testing.T) {
	repo := memory.NewGavelspaceRepository()
	ctx := context.Background()

	gavelspace, err := gsmodel.NewGavelspace("my-monorepo")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, gavelspace))

	found, err := repo.FindByName(ctx, gavelspace.ID())
	require.NoError(t, err)
	assert.Equal(t, gavelspace.ID(), found.ID())
}

func TestGavelspaceRepositoryFindByNameNotFound(t *testing.T) {
	repo := memory.NewGavelspaceRepository()
	ctx := context.Background()

	name, err := gsmodel.NewGavelspaceID("nonexistent")
	require.NoError(t, err)

	_, err = repo.FindByName(ctx, name)
	assert.ErrorIs(t, err, memory.ErrGavelspaceNotFound)
}

func TestGavelspaceRepositorySaveOverwrites(t *testing.T) {
	repo := memory.NewGavelspaceRepository()
	ctx := context.Background()

	gavelspace, err := gsmodel.NewGavelspace("my-monorepo")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, gavelspace))
	require.NoError(t, repo.Save(ctx, gavelspace))

	found, err := repo.FindByName(ctx, gavelspace.ID())
	require.NoError(t, err)
	assert.Equal(t, gavelspace.ID(), found.ID())
}
