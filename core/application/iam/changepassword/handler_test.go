package changepassword_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/changepassword"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

const validSession = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

type setup struct {
	users    *memiam.UserRepository
	sessions *memiam.SessionRepository
	hasher   *memiam.FakeHasher
	user     user.User
	handler  *changepassword.Handler
}

func newSetup(t *testing.T, mustChange bool) *setup {
	t.Helper()
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	hasher := memiam.NewFakeHasher()

	tenantID := tenant.NewTenantID(uuid.New())
	email, _ := user.NewEmail("alice@example.com")
	hash, _ := hasher.Hash("hunter22")
	foundUser, err := user.NewUser(tenantID, email, "Alice", user.RoleAdmin, hash, mustChange, testTime)
	require.NoError(t, err)
	foundUser.ClearEvents()
	require.NoError(t, users.Save(context.Background(), foundUser))

	return &setup{
		users:    users,
		sessions: sessions,
		hasher:   hasher,
		user:     foundUser,
		handler:  changepassword.NewHandler(users, sessions, hasher),
	}
}

func TestExecuteChangesPassword(t *testing.T) {
	setup := newSetup(t, true)

	tok, _ := session.NewToken(validSession)
	sess, _ := session.NewSession(tok, setup.user.ID(), "ua", "ip", testTime, testTime.Add(time.Hour))
	sess.ClearEvents()
	require.NoError(t, setup.sessions.Save(context.Background(), sess))

	cmd, _ := changepassword.NewCommand(setup.user.ID().String(), "hunter22", "newPassword!", testTime.Add(time.Minute))
	result, err := setup.handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, setup.user.ID().String(), result.UserID)
	require.Len(t, result.Events, 1)
	assert.Equal(t, user.EventNamePasswordChanged, result.Events[0].Name)

	got, _ := setup.users.ByID(context.Background(), setup.user.ID())
	assert.False(t, got.MustChangePassword(), "ChangePassword must clear must-change-password")

	ok, _ := setup.hasher.Verify("newPassword!", got.PasswordHash())
	assert.True(t, ok)

	_, err = setup.sessions.ByTokenHash(context.Background(), sess.TokenHash())
	require.Error(t, err, "sessions must be wiped on password change")
}

func TestExecuteRejectsWrongCurrentPassword(t *testing.T) {
	setup := newSetup(t, false)
	cmd, _ := changepassword.NewCommand(setup.user.ID().String(), "wrong", "newPassword!", testTime.Add(time.Minute))
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, changepassword.ErrCurrentPasswordWrong)
}

func TestExecuteRejectsMissingUser(t *testing.T) {
	setup := newSetup(t, false)
	cmd, _ := changepassword.NewCommand(uuid.NewString(), "hunter22", "newPassword!", testTime.Add(time.Minute))
	_, err := setup.handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	cases := []struct {
		name        string
		userID      string
		current     string
		newPassword string
		at          time.Time
	}{
		{name: "empty user", userID: "", current: "old", newPassword: "newPassword!", at: testTime},
		{name: "empty current", userID: "u-1", current: "", newPassword: "newPassword!", at: testTime},
		{name: "short new", userID: "u-1", current: "old", newPassword: "short", at: testTime},
		{name: "same as current", userID: "u-1", current: "samepass", newPassword: "samepass", at: testTime},
		{name: "zero time", userID: "u-1", current: "old", newPassword: "newPassword!", at: time.Time{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := changepassword.NewCommand(tc.userID, tc.current, tc.newPassword, tc.at)
			require.Error(t, err)
			assert.ErrorIs(t, err, changepassword.ErrInvalidCommand)
		})
	}
}
