package tenant_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

func TestNewTenantStatus(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    tenant.Status
		wantErr bool
	}{
		{name: "active", value: "active", want: tenant.StatusActive},
		{name: "suspended", value: "suspended", want: tenant.StatusSuspended},
		{name: "uppercase rejected (no normalisation, persisted lowercase by contract)", value: "ACTIVE", wantErr: true},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "unknown rejected", value: "deleted", wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			status, err := tenant.NewStatus(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
				return
			}
			require.NoError(t, err)
			assert.True(t, tcase.want.Equal(status))
		})
	}
}

func TestTenantStatusString(t *testing.T) {
	assert.Equal(t, "active", tenant.StatusActive.String())
	assert.Equal(t, "suspended", tenant.StatusSuspended.String())
}

func TestTenantStatusIsActive(t *testing.T) {
	assert.True(t, tenant.StatusActive.IsActive())
	assert.False(t, tenant.StatusSuspended.IsActive())
}
