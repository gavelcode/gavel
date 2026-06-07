package createuser_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/createuser"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

type setup struct {
	tenants *memiam.TenantRepository
	users   *memiam.UserRepository
	hasher  *memiam.FakeHasher
	tenant  tenant.Tenant
	handler *createuser.Handler
}

func newSetup(t *testing.T) *setup {
	t.Helper()
	tenants := memiam.NewTenantRepository()
	users := memiam.NewUserRepository()
	hasher := memiam.NewFakeHasher()

	slug, err := tenant.NewSlug("acme")
	require.NoError(t, err)
	foundTenant, err := tenant.NewTenant(slug, "Acme", testTime)
	require.NoError(t, err)
	foundTenant.ClearEvents()
	require.NoError(t, tenants.Save(context.Background(), foundTenant))

	return &setup{
		tenants: tenants,
		users:   users,
		hasher:  hasher,
		tenant:  foundTenant,
		handler: createuser.NewHandler(tenants, users, hasher),
	}
}

func TestExecuteCreatesUser(t *testing.T) {
	setup := newSetup(t)
	cmd, err := createuser.NewCommand(setup.tenant.ID().String(), "alice@example.com", "Alice", "admin", "hunter22", false, testTime)
	require.NoError(t, err)

	result, err := setup.handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.NotEmpty(t, result.UserID)
	assert.Equal(t, "alice@example.com", result.Email)
	assert.Equal(t, "admin", result.Role)
	require.Len(t, result.Events, 1)
	assert.Equal(t, user.EventNameUserCreated, result.Events[0].Name)

	uid, _ := user.ParseUserID(result.UserID)
	saved, err := setup.users.ByID(context.Background(), uid)
	require.NoError(t, err)
	assert.True(t, saved.IsActive())
}

func TestExecuteRejectsMissingTenant(t *testing.T) {
	setup := newSetup(t)
	cmd, _ := createuser.NewCommand(uuid.NewString(), "alice@example.com", "Alice", "admin", "hunter22", false, testTime)
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
}

func TestExecuteRejectsSuspendedTenant(t *testing.T) {
	setup := newSetup(t)
	require.NoError(t, setup.tenant.Suspend(testTime.Add(time.Hour)))
	setup.tenant.ClearEvents()
	require.NoError(t, setup.tenants.Save(context.Background(), setup.tenant))

	cmd, _ := createuser.NewCommand(setup.tenant.ID().String(), "alice@example.com", "Alice", "admin", "hunter22", false, testTime)
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestExecuteRejectsDuplicateEmailWithinTenant(t *testing.T) {
	setup := newSetup(t)
	cmd, _ := createuser.NewCommand(setup.tenant.ID().String(), "alice@example.com", "Alice", "admin", "hunter22", false, testTime)
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	cmd2, _ := createuser.NewCommand(setup.tenant.ID().String(), "alice@example.com", "Alice", "viewer", "hunter22", false, testTime)
	_, err = setup.handler.Execute(context.Background(), cmd2)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrEmailAlreadyInUse)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name string
		fn   func() (createuser.Command, error)
	}{
		{name: "empty tenant", fn: func() (createuser.Command, error) {
			return createuser.NewCommand("", "a@b.com", "A", "admin", "hunter22", false, testTime)
		}},
		{name: "empty email", fn: func() (createuser.Command, error) {
			return createuser.NewCommand("t-1", "", "A", "admin", "hunter22", false, testTime)
		}},
		{name: "empty name", fn: func() (createuser.Command, error) {
			return createuser.NewCommand("t-1", "a@b.com", "", "admin", "hunter22", false, testTime)
		}},
		{name: "empty role", fn: func() (createuser.Command, error) {
			return createuser.NewCommand("t-1", "a@b.com", "A", "", "hunter22", false, testTime)
		}},
		{name: "short password", fn: func() (createuser.Command, error) {
			return createuser.NewCommand("t-1", "a@b.com", "A", "admin", "short", false, testTime)
		}},
		{name: "zero time", fn: func() (createuser.Command, error) {
			return createuser.NewCommand("t-1", "a@b.com", "A", "admin", "hunter22", false, time.Time{})
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.fn()
			require.Error(t, err)
			assert.ErrorIs(t, err, createuser.ErrInvalidCommand)
		})
	}
}
