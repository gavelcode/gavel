package apitoken_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
)

func TestNewScope(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    apitoken.Scope
		wantErr bool
	}{
		{name: "ingest", value: "ingest", want: apitoken.ScopeIngest},
		{name: "read", value: "read", want: apitoken.ScopeRead},
		{name: "admin", value: "admin", want: apitoken.ScopeAdmin},
		{name: "project_sync", value: "project_sync", want: apitoken.ScopeProjectSync},
		{name: "uppercase normalised", value: "ADMIN", want: apitoken.ScopeAdmin},
		{name: "whitespace trimmed", value: " read ", want: apitoken.ScopeRead},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "unknown rejected", value: "write", wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			scope, err := apitoken.NewScope(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, apitoken.ErrInvalid)
				return
			}
			require.NoError(t, err)
			assert.True(t, tcase.want.Equal(scope))
		})
	}
}

func TestScopeString(t *testing.T) {
	assert.Equal(t, "ingest", apitoken.ScopeIngest.String())
	assert.Equal(t, "read", apitoken.ScopeRead.String())
	assert.Equal(t, "admin", apitoken.ScopeAdmin.String())
	assert.Equal(t, "project_sync", apitoken.ScopeProjectSync.String())
}

func TestScopesContain(t *testing.T) {
	scopes := []apitoken.Scope{apitoken.ScopeRead, apitoken.ScopeIngest}
	assert.True(t, apitoken.Scopes(scopes).Contains(apitoken.ScopeRead))
	assert.True(t, apitoken.Scopes(scopes).Contains(apitoken.ScopeIngest))
	assert.False(t, apitoken.Scopes(scopes).Contains(apitoken.ScopeAdmin))
	assert.False(t, apitoken.Scopes(nil).Contains(apitoken.ScopeRead))
}
