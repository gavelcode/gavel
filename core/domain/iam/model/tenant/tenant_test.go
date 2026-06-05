package tenant_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

var iamTestTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func mustTenantSlug(t *testing.T, raw string) tenant.Slug {
	t.Helper()
	slug, err := tenant.NewSlug(raw)
	require.NoError(t, err)
	return slug
}

func TestNewTenant(t *testing.T) {
	slug := mustTenantSlug(t, "acme")
	tests := []struct {
		name        string
		displayName string
		createdAt   time.Time
		wantErr     bool
	}{
		{name: "valid", displayName: "Acme Corp", createdAt: iamTestTime},
		{name: "trimmed display name", displayName: "  Acme Corp  ", createdAt: iamTestTime},
		{name: "empty display name rejected", displayName: "", createdAt: iamTestTime, wantErr: true},
		{name: "whitespace display name rejected", displayName: "   ", createdAt: iamTestTime, wantErr: true},
		{name: "zero createdAt rejected", displayName: "Acme Corp", createdAt: time.Time{}, wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			tnt, err := tenant.NewTenant(slug, tcase.displayName, tcase.createdAt)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
				return
			}
			require.NoError(t, err)
			assert.True(t, slug.Equal(tnt.Slug()))
			assert.Equal(t, "Acme Corp", tnt.DisplayName())
			assert.True(t, tnt.Status().IsActive())
			assert.Equal(t, tcase.createdAt, tnt.CreatedAt())
		})
	}
}

func TestNewTenantRecordsTenantCreatedEvent(t *testing.T) {
	slug := mustTenantSlug(t, "acme")
	tnt, err := tenant.NewTenant(slug, "Acme Corp", iamTestTime)
	require.NoError(t, err)

	events := tnt.Events()
	require.Len(t, events, 1)
	created, ok := events[0].(tenant.TenantCreated)
	require.True(t, ok, "first event must be TenantCreated, got %T", events[0])
	assert.True(t, created.TenantID().Equal(tnt.ID()))
	assert.True(t, created.Slug().Equal(slug))
	assert.Equal(t, "Acme Corp", created.DisplayName())
	assert.Equal(t, iamTestTime, created.OccurredAt())
	assert.Equal(t, tenant.EventNameTenantCreated, created.EventName())
}

func TestReconstituteTenant(t *testing.T) {
	tenantID := tenant.NewTenantID(uuid.New())
	slug := mustTenantSlug(t, "acme")

	tnt, err := tenant.ReconstituteTenant(tenantID, slug, "Acme Corp", tenant.StatusActive, iamTestTime)
	require.NoError(t, err)
	assert.True(t, tenantID.Equal(tnt.ID()))
	assert.Empty(t, tnt.Events(), "Reconstitute must not record events")

	_, err = tenant.ReconstituteTenant(tenantID, slug, "", tenant.StatusActive, iamTestTime)
	require.Error(t, err)

	_, err = tenant.ReconstituteTenant(tenantID, slug, "Acme Corp", tenant.StatusActive, time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
}

func TestTenantSuspend(t *testing.T) {
	slug := mustTenantSlug(t, "acme")
	tnt, _ := tenant.NewTenant(slug, "Acme Corp", iamTestTime)
	tnt.ClearEvents()

	suspendedAt := iamTestTime.Add(time.Hour)
	err := tnt.Suspend(suspendedAt)
	require.NoError(t, err)
	assert.False(t, tnt.Status().IsActive())

	events := tnt.Events()
	require.Len(t, events, 1)
	suspended, ok := events[0].(tenant.TenantSuspended)
	require.True(t, ok)
	assert.True(t, suspended.TenantID().Equal(tnt.ID()))
	assert.Equal(t, suspendedAt, suspended.OccurredAt())
	assert.Equal(t, tenant.EventNameTenantSuspended, suspended.EventName())

	err = tnt.Suspend(suspendedAt.Add(time.Hour))
	require.Error(t, err, "Suspend on an already-suspended tenant.Tenant must be rejected")
	assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
}

func TestTenantSuspendRejectsZeroTimestamp(t *testing.T) {
	slug := mustTenantSlug(t, "acme")
	tnt, _ := tenant.NewTenant(slug, "Acme Corp", iamTestTime)

	err := tnt.Suspend(time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
}
