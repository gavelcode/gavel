package tenant_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

func TestNewTenantSlug(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "lowercase alnum", value: "acme", want: "acme"},
		{name: "with hyphen", value: "acme-corp", want: "acme-corp"},
		{name: "with digit", value: "team42", want: "team42"},
		{name: "starts with digit", value: "1team", want: "1team"},
		{name: "uppercase normalised", value: "ACME-Corp", want: "acme-corp"},
		{name: "surrounding whitespace trimmed", value: "  acme  ", want: "acme"},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "whitespace only rejected", value: "   ", wantErr: true},
		{name: "internal whitespace rejected", value: "acme corp", wantErr: true},
		{name: "leading hyphen rejected", value: "-acme", wantErr: true},
		{name: "trailing hyphen rejected", value: "acme-", wantErr: true},
		{name: "underscore rejected", value: "acme_corp", wantErr: true},
		{name: "dot rejected", value: "acme.corp", wantErr: true},
		{name: "slash rejected", value: "acme/corp", wantErr: true},
		{name: "non-ascii rejected", value: "acmé", wantErr: true},
		{name: "too long rejected", value: strings.Repeat("a", 64), wantErr: true},
		{name: "at limit accepted", value: strings.Repeat("a", 63), want: strings.Repeat("a", 63)},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			slug, err := tenant.NewSlug(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.want, slug.String())
		})
	}
}

func TestTenantSlugEqualAndZero(t *testing.T) {
	a, _ := tenant.NewSlug("acme")
	b, _ := tenant.NewSlug("ACME")
	c, _ := tenant.NewSlug("acme-corp")
	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(c))
}
