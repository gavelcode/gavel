package activate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

var internalTestTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func TestNewHandlerPanicsOnNilTenants(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(nil)
	})
}

func TestExecuteReturnsErrorOnInvalidTenantID(t *testing.T) {
	handler := NewHandler(&stubTenantRepo{})
	cmd := Command{tenantID: "not-a-uuid", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
}

func TestExecuteReturnsErrorOnTenantSaveFailure(t *testing.T) {
	tn := seedInternalSuspendedTenant(t)
	repo := &stubTenantRepo{tenant: tn, saveErr: errors.New("save broken")}
	handler := NewHandler(repo)

	cmd := Command{tenantID: tn.ID().String(), occurredAt: internalTestTime.Add(2 * time.Hour)}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save tenant")
}

func seedInternalSuspendedTenant(t *testing.T) tenant.Tenant {
	t.Helper()
	slug, err := tenant.NewSlug("acme")
	require.NoError(t, err)
	tn, err := tenant.NewTenant(slug, "Acme", internalTestTime)
	require.NoError(t, err)
	require.NoError(t, tn.Suspend(internalTestTime.Add(time.Hour)))
	tn.ClearEvents()
	return tn
}

type stubTenantRepo struct {
	tenant  tenant.Tenant
	saveErr error
}

func (r *stubTenantRepo) Save(_ context.Context, _ tenant.Tenant) error { return r.saveErr }
func (r *stubTenantRepo) ByID(_ context.Context, _ tenant.TenantID) (tenant.Tenant, error) {
	return r.tenant, nil
}
func (r *stubTenantRepo) BySlug(_ context.Context, _ tenant.Slug) (tenant.Tenant, error) {
	return r.tenant, nil
}
