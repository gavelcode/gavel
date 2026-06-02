package memory_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/pleading/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/infrastructure/pleading/memory"
)

func TestPleadingRepository_SaveAndFindByID(t *testing.T) {
	repo := memory.NewPleadingRepository()
	ctx := context.Background()

	projectID := projectmodel.NewProjectID(uuid.New())

	pleading, err := model.FilePleading(projectID, 1, "Fix bug", "alice", "feature", "main", "abc123")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, pleading))

	found, err := repo.FindByID(ctx, pleading.ID())
	require.NoError(t, err)
	assert.Equal(t, pleading.ID(), found.ID())
	assert.Equal(t, "Fix bug", found.Title())
}

func TestPleadingRepository_FindByID_NotFound(t *testing.T) {
	repo := memory.NewPleadingRepository()
	ctx := context.Background()

	id := model.NewPleadingID(uuid.New())
	_, err := repo.FindByID(ctx, id)
	require.Error(t, err)
}
