package login_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/login"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

type setup struct {
	tenants  *memiam.TenantRepository
	users    *memiam.UserRepository
	sessions *memiam.SessionRepository
	hasher   *memiam.FakeHasher
	secrets  *memiam.FakeSecretGenerator
	tenant   tenant.Tenant
	user     user.User
	handler  *login.Handler
}

func newSetup(t *testing.T) *setup {
	t.Helper()
	tenants := memiam.NewTenantRepository()
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	hasher := memiam.NewFakeHasher()
	secrets := memiam.NewFakeSecretGenerator()

	slug, _ := tenant.NewSlug("acme")
	foundTenant, _ := tenant.NewTenant(slug, "Acme", testTime)
	foundTenant.ClearEvents()
	require.NoError(t, tenants.Save(context.Background(), foundTenant))

	email, _ := user.NewEmail("alice@example.com")
	hash, _ := hasher.Hash("hunter22")
	foundUser, err := user.NewUser(foundTenant.ID(), email, "Alice", user.RoleAdmin, hash, false, testTime)
	require.NoError(t, err)
	foundUser.ClearEvents()
	require.NoError(t, users.Save(context.Background(), foundUser))

	return &setup{
		tenants:  tenants,
		users:    users,
		sessions: sessions,
		hasher:   hasher,
		secrets:  secrets,
		tenant:   foundTenant,
		user:     foundUser,
		handler:  login.NewHandler(tenants, users, sessions, hasher, secrets),
	}
}

func TestExecuteLoginSuccess(t *testing.T) {
	setup := newSetup(t)
	cmd, err := login.NewCommand("acme", "alice@example.com", "hunter22", "Mozilla/5.0", "203.0.113.42", testTime, 24*time.Hour)
	require.NoError(t, err)

	result, err := setup.handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, setup.user.ID().String(), result.UserID)
	assert.Equal(t, setup.tenant.ID().String(), result.TenantID)
	assert.NotEmpty(t, result.SessionToken)
	assert.Equal(t, testTime.Add(24*time.Hour), result.SessionExpiresAt)
	require.Len(t, result.Events, 1)
	assert.Equal(t, session.EventNameCreated, result.Events[0].Name)

	tok, _ := session.NewToken(result.SessionToken)
	hash := session.HashToken(tok)
	session, err := setup.sessions.ByTokenHash(context.Background(), hash)
	require.NoError(t, err)
	assert.True(t, setup.user.ID().Equal(session.UserID()))

	got, _ := setup.users.ByID(context.Background(), setup.user.ID())
	assert.Equal(t, testTime, got.LastLoginAt())
}

func TestExecuteRejectsWrongPassword(t *testing.T) {
	setup := newSetup(t)
	cmd, _ := login.NewCommand("acme", "alice@example.com", "wrong", "ua", "ip", testTime, time.Hour)
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestExecuteRejectsUnknownTenant(t *testing.T) {
	setup := newSetup(t)
	cmd, _ := login.NewCommand("nope", "alice@example.com", "hunter22", "ua", "ip", testTime, time.Hour)
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestExecuteRejectsSuspendedTenant(t *testing.T) {
	setup := newSetup(t)
	require.NoError(t, setup.tenant.Suspend(testTime.Add(time.Minute)))
	setup.tenant.ClearEvents()
	require.NoError(t, setup.tenants.Save(context.Background(), setup.tenant))

	cmd, _ := login.NewCommand("acme", "alice@example.com", "hunter22", "ua", "ip", testTime, time.Hour)
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestExecuteRejectsInactiveUser(t *testing.T) {
	setup := newSetup(t)
	require.NoError(t, setup.user.Deactivate(testTime.Add(time.Minute)))
	setup.user.ClearEvents()
	require.NoError(t, setup.users.Save(context.Background(), setup.user))

	cmd, _ := login.NewCommand("acme", "alice@example.com", "hunter22", "ua", "ip", testTime, time.Hour)
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestExecuteRejectsUnknownEmail(t *testing.T) {
	setup := newSetup(t)
	cmd, _ := login.NewCommand("acme", "missing@example.com", "hunter22", "ua", "ip", testTime, time.Hour)
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	cases := []struct {
		name string
		fn   func() (login.Command, error)
	}{
		{name: "empty tenant slug", fn: func() (login.Command, error) {
			return login.NewCommand("", "a@b.com", "pass", "ua", "ip", testTime, time.Hour)
		}},
		{name: "empty email", fn: func() (login.Command, error) {
			return login.NewCommand("acme", "", "pass", "ua", "ip", testTime, time.Hour)
		}},
		{name: "empty password", fn: func() (login.Command, error) {
			return login.NewCommand("acme", "a@b.com", "", "ua", "ip", testTime, time.Hour)
		}},
		{name: "zero time", fn: func() (login.Command, error) {
			return login.NewCommand("acme", "a@b.com", "pass", "ua", "ip", time.Time{}, time.Hour)
		}},
		{name: "non-positive ttl", fn: func() (login.Command, error) {
			return login.NewCommand("acme", "a@b.com", "pass", "ua", "ip", testTime, 0)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.fn()
			require.Error(t, err)
			assert.ErrorIs(t, err, login.ErrInvalidCommand)
		})
	}
}
