package firstadmin_test

import (
	"crypto/rand"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/apps/server/internal/platform/config"
	"github.com/usegavel/gavel/apps/server/internal/platform/firstadmin"
)

func TestResolvePasswordUsesConfiguredValue(t *testing.T) {
	cfg := &config.Config{AdminPassword: "s3cret-from-env"}

	password, generated, err := firstadmin.ResolvePassword(cfg, rand.Reader)

	require.NoError(t, err)
	assert.Equal(t, "s3cret-from-env", password)
	assert.False(t, generated, "a configured password is not generated")
}

func TestResolvePasswordGeneratesWhenUnset(t *testing.T) {
	cfg := &config.Config{AdminPassword: ""}

	password, generated, err := firstadmin.ResolvePassword(cfg, rand.Reader)

	require.NoError(t, err)
	assert.NotEmpty(t, password)
	assert.True(t, generated, "an unset password must be flagged generated so the caller logs it after commit")
}

func TestResolvePasswordErrorsWhenGenerationFails(t *testing.T) {
	cfg := &config.Config{AdminPassword: ""}

	_, _, err := firstadmin.ResolvePassword(cfg, iotest.ErrReader(errors.New("rng broken")))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "generate initial admin password")
}
