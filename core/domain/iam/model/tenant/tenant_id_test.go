package tenant_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

func TestNewTenantIDWrapsUUID(t *testing.T) {
	u := uuid.New()
	id := tenant.NewTenantID(u)
	assert.Equal(t, u, id.UUID())
	assert.Equal(t, u.String(), id.String())
}

func TestTenantIDEqualByUnderlyingUUID(t *testing.T) {
	u := uuid.New()
	a := tenant.NewTenantID(u)
	b := tenant.NewTenantID(u)
	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(tenant.NewTenantID(uuid.New())))
}

func TestParseTenantIDAcceptsCanonicalUUID(t *testing.T) {
	u := uuid.New()
	id, err := tenant.ParseTenantID(u.String())
	require.NoError(t, err)
	assert.Equal(t, u, id.UUID())
}

func TestParseTenantIDRejectsNonUUID(t *testing.T) {
	for _, raw := range []string{"", "   ", "t-1", "garbage", uuid.Nil.String()} {
		_, err := tenant.ParseTenantID(raw)
		require.Error(t, err)
		assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
	}
}

func TestTenantIDIsZero(t *testing.T) {
	assert.True(t, tenant.TenantID{}.IsZero(), "zero-value TenantID is zero")
	assert.True(t, tenant.NewTenantID(uuid.Nil).IsZero(), "nil UUID is zero")
	assert.False(t, tenant.NewTenantID(uuid.New()).IsZero(), "a real UUID is not zero")
	assert.False(t, tenant.LocalTenantID.IsZero(), "the local sentinel is not zero")
}
