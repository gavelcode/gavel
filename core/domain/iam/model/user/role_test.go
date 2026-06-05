package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

func TestNewRole(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    user.Role
		wantErr bool
	}{
		{name: "admin", value: "admin", want: user.RoleAdmin},
		{name: "maintainer", value: "maintainer", want: user.RoleMaintainer},
		{name: "viewer", value: "viewer", want: user.RoleViewer},
		{name: "uppercase normalised", value: "ADMIN", want: user.RoleAdmin},
		{name: "whitespace trimmed", value: " viewer ", want: user.RoleViewer},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "unknown rejected", value: "owner", wantErr: true},
		{name: "case-mixed unknown rejected", value: "ROOT", wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			role, err := user.NewRole(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, user.ErrInvalidUser)
				return
			}
			require.NoError(t, err)
			assert.True(t, tcase.want.Equal(role), "expected %q got %q", tcase.want, role)
		})
	}
}

func TestRoleString(t *testing.T) {
	assert.Equal(t, "admin", user.RoleAdmin.String())
	assert.Equal(t, "maintainer", user.RoleMaintainer.String())
	assert.Equal(t, "viewer", user.RoleViewer.String())
}

func TestRoleIsAdmin(t *testing.T) {
	assert.True(t, user.RoleAdmin.IsAdmin())
	assert.False(t, user.RoleMaintainer.IsAdmin())
	assert.False(t, user.RoleViewer.IsAdmin())
}
