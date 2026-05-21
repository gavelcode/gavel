package sourceblob_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"
)

func setupDB(t *testing.T) *database.DB { return testkit.TestDB(t) }

func insertTestProject(t *testing.T, db *database.DB) projectmodel.Project {
	t.Helper()
	project, err := projectmodel.NewProject("test-project", "Test Project", "//test/...")
	require.NoError(t, err)
	repo := projectpostgres.NewRepository(db)
	require.NoError(t, repo.Save(context.Background(), project))
	return project
}
