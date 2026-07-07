package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tenantmodel "github.com/usegavel/gavel/core/domain/iam/model/tenant"
	usermodel "github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/infrastructure/iam/postgres"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
)

var provisionerTestTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func buildTenant(t *testing.T, slugRaw string) tenantmodel.Tenant {
	t.Helper()
	slug, err := tenantmodel.NewSlug(slugRaw)
	require.NoError(t, err)
	newTenant, err := tenantmodel.NewTenant(slug, "Display "+slugRaw, provisionerTestTime)
	require.NoError(t, err)
	return newTenant
}

func buildAdmin(t *testing.T, tenantID tenantmodel.TenantID, emailRaw string) usermodel.User {
	t.Helper()
	email, err := usermodel.NewEmail(emailRaw)
	require.NoError(t, err)
	hash, err := usermodel.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$YWJjZGVmZ2hpamts$c29tZWtleXZhbHVlaGVy")
	require.NoError(t, err)
	admin, err := usermodel.NewUser(tenantID, email, "Administrator", usermodel.RoleAdmin, hash, true, provisionerTestTime)
	require.NoError(t, err)
	return admin
}

func TestProvisionPersistsTenantAndAdmin(t *testing.T) {
	testDB := testkit.TestDB(t)
	ctx := context.Background()
	provisioner := postgres.NewTenantProvisioner(testDB)

	newTenant := buildTenant(t, "acme")
	admin := buildAdmin(t, newTenant.ID(), "admin@acme.com")
	require.NoError(t, provisioner.Provision(ctx, newTenant, admin))

	slug, _ := tenantmodel.NewSlug("acme")
	savedTenant, err := postgres.NewTenantRepo(testDB).BySlug(ctx, slug)
	require.NoError(t, err)
	assert.Equal(t, newTenant.ID().String(), savedTenant.ID().String())

	email, _ := usermodel.NewEmail("admin@acme.com")
	savedAdmin, err := postgres.NewUserRepo(testDB).ByEmail(ctx, newTenant.ID(), email)
	require.NoError(t, err)
	assert.Equal(t, usermodel.RoleAdmin, savedAdmin.Role())
}

func TestProvisionRollsBackTenantWhenAdminSaveFails(t *testing.T) {
	testDB := testkit.TestDB(t)
	ctx := context.Background()
	provisioner := postgres.NewTenantProvisioner(testDB)

	newTenant := buildTenant(t, "rollback")
	// The admin references a different, non-existent tenant, so its insert
	// violates the tenant_id foreign key after the tenant row is written.
	orphanTenantID := tenantmodel.NewTenantID(uuid.New())
	admin := buildAdmin(t, orphanTenantID, "admin@rollback.com")

	err := provisioner.Provision(ctx, newTenant, admin)
	require.Error(t, err, "a failing admin save must fail the whole provision")

	slug, _ := tenantmodel.NewSlug("rollback")
	_, err = postgres.NewTenantRepo(testDB).BySlug(ctx, slug)
	require.ErrorIs(t, err, tenantmodel.ErrTenantNotFound, "the tenant must be rolled back, not left orphaned")
}
