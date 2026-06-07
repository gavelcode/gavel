package suspend_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/tenant/suspend"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

func seedTenant(t *testing.T, tenants *memiam.TenantRepository, slugRaw string) tenant.Tenant {
	t.Helper()
	slug, err := tenant.NewSlug(slugRaw)
	require.NoError(t, err)
	foundTenant, err := tenant.NewTenant(slug, "Display "+slugRaw, testTime)
	require.NoError(t, err)
	foundTenant.ClearEvents()
	require.NoError(t, tenants.Save(context.Background(), foundTenant))
	return foundTenant
}

func TestExecuteSuspendsTenant(t *testing.T) {
	tenants := memiam.NewTenantRepository()
	foundTenant := seedTenant(t, tenants, "acme")

	cmd, err := suspend.NewCommand(foundTenant.ID().String(), testTime.Add(time.Hour))
	require.NoError(t, err)

	result, err := suspend.NewHandler(tenants).Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, foundTenant.ID().String(), result.TenantID)
	require.Len(t, result.Events, 1)
	assert.Equal(t, tenant.EventNameTenantSuspended, result.Events[0].Name)

	got, _ := tenants.ByID(context.Background(), foundTenant.ID())
	assert.False(t, got.Status().IsActive())
}

func TestExecuteRejectsMissingTenant(t *testing.T) {
	tenants := memiam.NewTenantRepository()
	cmd, _ := suspend.NewCommand(uuid.NewString(), testTime.Add(time.Hour))
	_, err := suspend.NewHandler(tenants).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
}

func TestExecuteRejectsDoubleSuspend(t *testing.T) {
	tenants := memiam.NewTenantRepository()
	foundTenant := seedTenant(t, tenants, "acme")

	cmd, _ := suspend.NewCommand(foundTenant.ID().String(), testTime.Add(time.Hour))
	_, err := suspend.NewHandler(tenants).Execute(context.Background(), cmd)
	require.NoError(t, err)

	_, err = suspend.NewHandler(tenants).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	_, err := suspend.NewCommand("", testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, suspend.ErrInvalidCommand)

	_, err = suspend.NewCommand("t-1", time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, suspend.ErrInvalidCommand)
}
