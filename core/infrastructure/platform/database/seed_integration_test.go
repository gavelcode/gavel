package database_test

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/infrastructure/iam/argon2"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
)

func TestSeed(t *testing.T) {
	database := testkit.TestDB(t)
	ctx := context.Background()

	t.Run("defaultTenantExists", func(t *testing.T) {
		var slug, displayName, status string
		err := database.DB.QueryRowContext(ctx,
			"SELECT slug, display_name, status FROM iam_tenants WHERE slug = $1",
			"default",
		).Scan(&slug, &displayName, &status)
		require.NoError(t, err)
		require.Equal(t, "default", slug)
		require.Equal(t, "Default", displayName)
		require.Equal(t, "active", status)
	})

	t.Run("adminUserExistsWithChangemePassword", func(t *testing.T) {
		var hashStr, role string
		var mustChange bool
		err := database.DB.QueryRowContext(ctx, `
			SELECT u.password_hash, u.role, u.must_change_password
			FROM iam_users u
			JOIN iam_tenants t ON t.id = u.tenant_id
			WHERE t.slug = $1 AND u.email = $2
		`, "default", "admin@gavel.local").Scan(&hashStr, &role, &mustChange)
		require.NoError(t, err)
		require.Equal(t, "admin", role)
		require.True(t, mustChange, "first login must be forced to change the seeded password")

		hash, err := user.NewPasswordHash(hashStr)
		require.NoError(t, err)
		ok, err := argon2.New(rand.Reader).Verify("changeme", hash)
		require.NoError(t, err)
		require.True(t, ok, "seeded hash must validate the literal \"changeme\"")
	})
}
