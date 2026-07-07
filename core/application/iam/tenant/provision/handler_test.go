package provision_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/tenant/provision"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func newHandler() (*provision.Handler, *memiam.TenantRepository, *memiam.UserRepository) {
	tenants := memiam.NewTenantRepository()
	users := memiam.NewUserRepository()
	provisioner := memiam.NewProvisioner(tenants, users)
	return provision.NewHandler(provisioner, memiam.NewFakeHasher()), tenants, users
}

func TestExecuteProvisionsTenantWithAdmin(t *testing.T) {
	handler, tenants, users := newHandler()

	cmd, err := provision.NewCommand("acme", "Acme Corp", "admin@acme.com", "Administrator", "s3cret-pass", testTime)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.NotEmpty(t, result.TenantID)
	assert.NotEmpty(t, result.AdminUserID)
	require.Len(t, result.Events, 2, "provision must record TenantCreated + UserCreated")
	assert.Equal(t, tenant.EventNameTenantCreated, result.Events[0].Name)

	slug, _ := tenant.NewSlug("acme")
	savedTenant, err := tenants.BySlug(context.Background(), slug)
	require.NoError(t, err)
	assert.True(t, savedTenant.Status().IsActive())

	email, _ := user.NewEmail("admin@acme.com")
	admin, err := users.ByEmail(context.Background(), savedTenant.ID(), email)
	require.NoError(t, err)
	assert.Equal(t, user.RoleAdmin, admin.Role(), "the seeded admin must have the admin role")
	assert.True(t, admin.MustChangePassword(), "the provisioned admin must rotate its password on first login")
	assert.Equal(t, savedTenant.ID().String(), admin.TenantID().String())
}

func TestExecuteRejectsDuplicateSlug(t *testing.T) {
	handler, _, _ := newHandler()

	first, _ := provision.NewCommand("acme", "Acme", "admin@acme.com", "Administrator", "s3cret-pass", testTime)
	_, err := handler.Execute(context.Background(), first)
	require.NoError(t, err)

	second, _ := provision.NewCommand("acme", "Acme Two", "other@acme.com", "Administrator", "s3cret-pass", testTime)
	_, err = handler.Execute(context.Background(), second)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrSlugTaken)
}

func TestExecuteRejectsInvalidSlug(t *testing.T) {
	handler, _, _ := newHandler()

	cmd, _ := provision.NewCommand("ACME corp!", "Acme", "admin@acme.com", "Administrator", "s3cret-pass", testTime)
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err, "domain rejects a slug with whitespace and uppercase")
	assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
}

func TestExecuteRejectsInvalidAdminEmail(t *testing.T) {
	handler, _, _ := newHandler()

	cmd, _ := provision.NewCommand("acme", "Acme", "not-an-email", "Administrator", "s3cret-pass", testTime)
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err, "domain rejects a malformed admin email")
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name                                               string
		slug, displayName, email, adminName, adminPassword string
		at                                                 time.Time
	}{
		{name: "empty slug", displayName: "Acme", email: "a@b.com", adminName: "Admin", adminPassword: "pw", at: testTime},
		{name: "empty display name", slug: "acme", email: "a@b.com", adminName: "Admin", adminPassword: "pw", at: testTime},
		{name: "empty admin email", slug: "acme", displayName: "Acme", adminName: "Admin", adminPassword: "pw", at: testTime},
		{name: "empty admin name", slug: "acme", displayName: "Acme", email: "a@b.com", adminPassword: "pw", at: testTime},
		{name: "empty admin password", slug: "acme", displayName: "Acme", email: "a@b.com", adminName: "Admin", at: testTime},
		{name: "zero time", slug: "acme", displayName: "Acme", email: "a@b.com", adminName: "Admin", adminPassword: "pw", at: time.Time{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := provision.NewCommand(tc.slug, tc.displayName, tc.email, tc.adminName, tc.adminPassword, tc.at)
			require.Error(t, err)
			assert.ErrorIs(t, err, provision.ErrInvalidCommand)
		})
	}
}
