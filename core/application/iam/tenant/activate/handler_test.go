package activate_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/tenant/activate"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func seedSuspendedTenant(t *testing.T, tenants *memiam.TenantRepository, slugRaw string) tenant.Tenant {
	t.Helper()
	slug, err := tenant.NewSlug(slugRaw)
	require.NoError(t, err)
	foundTenant, err := tenant.NewTenant(slug, "Display "+slugRaw, testTime)
	require.NoError(t, err)
	require.NoError(t, foundTenant.Suspend(testTime.Add(time.Hour)))
	foundTenant.ClearEvents()
	require.NoError(t, tenants.Save(context.Background(), foundTenant))
	return foundTenant
}

func TestExecuteActivatesTenant(t *testing.T) {
	tenants := memiam.NewTenantRepository()
	foundTenant := seedSuspendedTenant(t, tenants, "acme")

	cmd, err := activate.NewCommand(foundTenant.ID().String(), testTime.Add(2*time.Hour))
	require.NoError(t, err)

	result, err := activate.NewHandler(tenants).Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, foundTenant.ID().String(), result.TenantID)
	require.Len(t, result.Events, 1)
	assert.Equal(t, tenant.EventNameTenantActivated, result.Events[0].Name)

	got, _ := tenants.ByID(context.Background(), foundTenant.ID())
	assert.True(t, got.Status().IsActive())
}

func TestExecuteRejectsMissingTenant(t *testing.T) {
	tenants := memiam.NewTenantRepository()
	cmd, _ := activate.NewCommand(uuid.NewString(), testTime.Add(time.Hour))
	_, err := activate.NewHandler(tenants).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
}

func TestExecuteRejectsActivatingActiveTenant(t *testing.T) {
	tenants := memiam.NewTenantRepository()
	slug, err := tenant.NewSlug("acme")
	require.NoError(t, err)
	active, err := tenant.NewTenant(slug, "Acme", testTime)
	require.NoError(t, err)
	active.ClearEvents()
	require.NoError(t, tenants.Save(context.Background(), active))

	cmd, _ := activate.NewCommand(active.ID().String(), testTime.Add(time.Hour))
	_, err = activate.NewHandler(tenants).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	_, err := activate.NewCommand("", testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, activate.ErrInvalidCommand)

	_, err = activate.NewCommand("t-1", time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, activate.ErrInvalidCommand)
}
