package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/infrastructure/gavelspace/memory"
)

func TestGavelspaceRepositorySaveAndFindByName(t *testing.T) {
	repo := memory.NewGavelspaceRepository()
	ctx := context.Background()

	gavelspace, err := gsmodel.NewGavelspace(testTenantID, "my-monorepo")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, gavelspace))

	found, err := repo.FindByName(ctx, testTenantID, gavelspace.ID())
	require.NoError(t, err)
	assert.Equal(t, gavelspace.ID(), found.ID())
}

func TestGavelspaceRepositoryFindByNameNotFound(t *testing.T) {
	repo := memory.NewGavelspaceRepository()
	ctx := context.Background()

	name, err := gsmodel.NewGavelspaceID("nonexistent")
	require.NoError(t, err)

	_, err = repo.FindByName(ctx, testTenantID, name)
	assert.ErrorIs(t, err, memory.ErrGavelspaceNotFound)
}

func TestGavelspaceRepositorySaveOverwrites(t *testing.T) {
	repo := memory.NewGavelspaceRepository()
	ctx := context.Background()

	gavelspace, err := gsmodel.NewGavelspace(testTenantID, "my-monorepo")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, gavelspace))
	require.NoError(t, repo.Save(ctx, gavelspace))

	found, err := repo.FindByName(ctx, testTenantID, gavelspace.ID())
	require.NoError(t, err)
	assert.Equal(t, gavelspace.ID(), found.ID())
}

var testTenantID = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

func TestGavelspaceRepositoryIsolatesByTenant(t *testing.T) {
	repo := memory.NewGavelspaceRepository()
	ctx := context.Background()

	gavelspace, err := gsmodel.NewGavelspace(testTenantID, "my-monorepo")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, gavelspace))

	otherTenant := tenant.NewTenantID(uuid.MustParse("33333333-3333-3333-3333-333333333333"))
	_, err = repo.FindByName(ctx, otherTenant, gavelspace.ID())
	assert.ErrorIs(t, err, memory.ErrGavelspaceNotFound, "a tenant must not read another tenant's gavelspace")
}
