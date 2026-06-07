package create

import (
	"context"
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

func TestExecuteReturnsErrorOnNewTenantDomainFailure(t *testing.T) {
	handler := NewHandler(&stubTenantRepo{})
	cmd := Command{slug: "valid", displayName: "", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new tenant")
}

type stubTenantRepo struct{}

func (r *stubTenantRepo) Save(_ context.Context, _ tenant.Tenant) error { return nil }
func (r *stubTenantRepo) ByID(_ context.Context, _ tenant.TenantID) (tenant.Tenant, error) {
	return tenant.Tenant{}, nil
}
func (r *stubTenantRepo) BySlug(_ context.Context, _ tenant.Slug) (tenant.Tenant, error) {
	return tenant.Tenant{}, nil
}
