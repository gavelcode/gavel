package create_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/tenant/create"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

func TestExecuteCreatesTenant(t *testing.T) {
	tenants := memiam.NewTenantRepository()
	handler := create.NewHandler(tenants)

	cmd, err := create.NewCommand("acme", "Acme Corp", testTime)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.NotEmpty(t, result.TenantID)
	assert.Equal(t, "acme", result.Slug)
	assert.Equal(t, "Acme Corp", result.DisplayName)
	require.Len(t, result.Events, 1)
	assert.Equal(t, tenant.EventNameTenantCreated, result.Events[0].Name)

	tenantID, _ := tenant.ParseTenantID(result.TenantID)
	saved, err := tenants.ByID(context.Background(), tenantID)
	require.NoError(t, err)
	assert.True(t, saved.Status().IsActive())
}

func TestExecuteRejectsDuplicateSlug(t *testing.T) {
	tenants := memiam.NewTenantRepository()
	handler := create.NewHandler(tenants)

	cmd, _ := create.NewCommand("acme", "Acme Corp", testTime)
	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrSlugTaken)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name        string
		slug        string
		displayName string
		at          time.Time
	}{
		{name: "empty slug", slug: "", displayName: "n", at: testTime},
		{name: "blank slug", slug: "   ", displayName: "n", at: testTime},
		{name: "empty display name", slug: "acme", displayName: "", at: testTime},
		{name: "zero time", slug: "acme", displayName: "n", at: time.Time{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := create.NewCommand(tc.slug, tc.displayName, tc.at)
			require.Error(t, err)
			assert.ErrorIs(t, err, create.ErrInvalidCommand)
		})
	}
}

func TestExecutePropagatesDomainValidationError(t *testing.T) {
	tenants := memiam.NewTenantRepository()
	handler := create.NewHandler(tenants)

	cmd, _ := create.NewCommand("ACME corp!", "Acme", testTime)
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err, "domain rejects slug with whitespace and uppercase")
	assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
}
