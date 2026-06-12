package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/apps/server/internal/platform/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("GAVEL_DATA_DIR", t.TempDir())
	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, ":8080", cfg.Addr)
	assert.Equal(t, "postgres://localhost:5432/gavel?sslmode=disable", cfg.DatabaseURL)
	assert.False(t, cfg.SecureCookies)
	assert.Equal(t, "gavel_session", cfg.SessionCookie)
	assert.Equal(t, "gav_", cfg.APITokenPrefix)
}

func TestLoadCustomEnvVars(t *testing.T) {
	t.Setenv("GAVEL_DATA_DIR", t.TempDir())
	t.Setenv("GAVEL_ADDR", ":9090")
	t.Setenv("GAVEL_DATABASE_URL", "postgres://custom:5432/db")
	t.Setenv("GAVEL_SESSION_TTL_HOURS", "24")
	t.Setenv("GAVEL_SECURE_COOKIES", "true")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, ":9090", cfg.Addr)
	assert.Equal(t, "postgres://custom:5432/db", cfg.DatabaseURL)
	assert.True(t, cfg.SecureCookies)
}

func TestLoadFailsWhenDataDirNotCreatable(t *testing.T) {
	t.Setenv("GAVEL_DATA_DIR", "/dev/null/invalid")

	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create data dir")
}

func TestLoadFailsWhenSarifDirNotCreatable(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("GAVEL_DATA_DIR", tmp)

	require.NoError(t, os.WriteFile(tmp+"/analyses", []byte("block"), 0o644))

	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create sarif dir")
}
