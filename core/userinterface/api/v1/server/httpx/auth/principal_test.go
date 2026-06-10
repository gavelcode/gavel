package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

func TestWithPrincipalAndFromContext(t *testing.T) {
	principal := &auth.Principal{
		UserID:      "user-1",
		TenantID:    "tenant-1",
		Email:       "user@example.com",
		DisplayName: "Test User",
		Role:        "admin",
	}
	ctx := auth.WithPrincipal(context.Background(), principal)

	got, ok := auth.PrincipalFromContext(ctx)

	require.True(t, ok)
	assert.Equal(t, "user-1", got.UserID)
	assert.Equal(t, "tenant-1", got.TenantID)
	assert.Equal(t, "user@example.com", got.Email)
	assert.Equal(t, "admin", got.Role)
}

func TestPrincipalFromContextMissing(t *testing.T) {
	got, ok := auth.PrincipalFromContext(context.Background())

	assert.False(t, ok)
	assert.Nil(t, got)
}
