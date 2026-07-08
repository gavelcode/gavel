package sourceblob_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"
)

var testTenantID = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

func setupDB(t *testing.T) *database.DB {
	testDB := testkit.TestDB(t)
	seedTenant(t, testDB)
	return testDB
}

func seedTenant(t *testing.T, testDB *database.DB) {
	t.Helper()
	_, err := testDB.ExecContext(context.Background(),
		`INSERT INTO iam_tenants (id, slug, display_name, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		testTenantID.UUID(), "test-tenant", "Test Tenant", "active", database.Now())
	require.NoError(t, err)
}

func insertTestProject(t *testing.T, db *database.DB) projectmodel.Project {
	t.Helper()
	project, err := projectmodel.NewProject(testTenantID, "test-project", "Test Project", "//test/...")
	require.NoError(t, err)
	repo := projectpostgres.NewRepository(db)
	require.NoError(t, repo.Save(context.Background(), project))
	return project
}
