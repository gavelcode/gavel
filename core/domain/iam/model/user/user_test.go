package user_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

var testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func mustTenantID(t *testing.T) tenant.TenantID {
	t.Helper()
	return tenant.NewTenantID(uuid.New())
}

func mustEmail(t *testing.T, raw string) user.Email {
	t.Helper()
	e, err := user.NewEmail(raw)
	require.NoError(t, err)
	return e
}

func mustHash(t *testing.T) user.PasswordHash {
	t.Helper()
	h, err := user.NewPasswordHash(validArgon2Hash)
	require.NoError(t, err)
	return h
}

type newUserInput struct {
	tenantID           tenant.TenantID
	email              user.Email
	displayName        string
	role               user.Role
	hash               user.PasswordHash
	mustChangePassword bool
	createdAt          time.Time
}

func validUserInput(t *testing.T) newUserInput {
	t.Helper()
	return newUserInput{
		tenantID:           mustTenantID(t),
		email:              mustEmail(t, "alice@example.com"),
		displayName:        "Alice",
		role:               user.RoleAdmin,
		hash:               mustHash(t),
		mustChangePassword: false,
		createdAt:          testTime,
	}
}

func TestNewUser(t *testing.T) {
	input := validUserInput(t)
	usr, err := user.NewUser(input.tenantID, input.email, input.displayName, input.role, input.hash, input.mustChangePassword, input.createdAt)
	require.NoError(t, err)

	assert.True(t, input.tenantID.Equal(usr.TenantID()))
	assert.True(t, input.email.Equal(usr.Email()))
	assert.Equal(t, input.displayName, usr.DisplayName())
	assert.True(t, input.role.Equal(usr.Role()))
	assert.True(t, input.hash.Equal(usr.PasswordHash()))
	assert.False(t, usr.MustChangePassword())
	assert.True(t, usr.IsActive())
	assert.Equal(t, input.createdAt, usr.CreatedAt())
	assert.True(t, usr.LastLoginAt().IsZero())
}

func TestNewUserRejectsInvalidInputs(t *testing.T) {
	input := validUserInput(t)

	cases := []struct {
		name  string
		patch func(*newUserInput)
	}{
		{name: "empty display name", patch: func(p *newUserInput) { p.displayName = "  " }},
		{name: "zero createdAt", patch: func(p *newUserInput) { p.createdAt = time.Time{} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bad := input
			tc.patch(&bad)
			_, err := user.NewUser(bad.tenantID, bad.email, bad.displayName, bad.role, bad.hash, bad.mustChangePassword, bad.createdAt)
			require.Error(t, err)
			assert.ErrorIs(t, err, user.ErrInvalidUser)
		})
	}
}

func TestNewUserRecordsUserCreatedEvent(t *testing.T) {
	input := validUserInput(t)
	usr, err := user.NewUser(input.tenantID, input.email, input.displayName, input.role, input.hash, true, input.createdAt)
	require.NoError(t, err)

	events := usr.Events()
	require.Len(t, events, 1)
	created, ok := events[0].(user.UserCreated)
	require.True(t, ok)
	assert.True(t, created.UserID().Equal(usr.ID()))
	assert.True(t, created.TenantID().Equal(input.tenantID))
	assert.True(t, created.Email().Equal(input.email))
	assert.True(t, created.Role().Equal(input.role))
	assert.True(t, created.MustChangePassword())
	assert.Equal(t, input.createdAt, created.OccurredAt())
}

func TestReconstituteUser(t *testing.T) {
	id := user.NewUserID(uuid.New())
	input := validUserInput(t)

	usr, err := user.ReconstituteUser(id, input.tenantID, input.email, input.displayName, input.role, input.hash, true, true, input.createdAt, testTime.Add(time.Hour))
	require.NoError(t, err)
	assert.True(t, id.Equal(usr.ID()))
	assert.True(t, usr.MustChangePassword())
	assert.True(t, usr.IsActive())
	assert.Equal(t, testTime.Add(time.Hour), usr.LastLoginAt())
	assert.Empty(t, usr.Events())
}

func TestReconstituteUserRejectsInvalidInputs(t *testing.T) {
	id := user.NewUserID(uuid.New())
	input := validUserInput(t)

	_, err := user.ReconstituteUser(id, input.tenantID, input.email, "  ", input.role, input.hash, false, true, input.createdAt, time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestUserDeactivateRejectsZeroTimestamp(t *testing.T) {
	input := validUserInput(t)
	usr, _ := user.NewUser(input.tenantID, input.email, input.displayName, input.role, input.hash, false, input.createdAt)

	err := usr.Deactivate(time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestUserChangePassword(t *testing.T) {
	input := validUserInput(t)
	usr, _ := user.NewUser(input.tenantID, input.email, input.displayName, input.role, input.hash, true, input.createdAt)
	usr.ClearEvents()

	newHash, err := user.NewPasswordHash(validArgon2Hash)
	require.NoError(t, err)

	eventTime := testTime.Add(time.Hour)
	err = usr.ChangePassword(newHash, eventTime)
	require.NoError(t, err)
	assert.False(t, usr.MustChangePassword(), "ChangePassword must clear the must-change flag")
	assert.True(t, newHash.Equal(usr.PasswordHash()))

	events := usr.Events()
	require.Len(t, events, 1)
	changed, ok := events[0].(user.PasswordChanged)
	require.True(t, ok)
	assert.True(t, changed.UserID().Equal(usr.ID()))
	assert.Equal(t, eventTime, changed.OccurredAt())

	err = usr.ChangePassword(newHash, time.Time{})
	require.Error(t, err, "zero occurredAt must be rejected")
}

func TestUserDeactivate(t *testing.T) {
	input := validUserInput(t)
	usr, _ := user.NewUser(input.tenantID, input.email, input.displayName, input.role, input.hash, false, input.createdAt)
	usr.ClearEvents()

	eventTime := testTime.Add(time.Hour)
	err := usr.Deactivate(eventTime)
	require.NoError(t, err)
	assert.False(t, usr.IsActive())

	events := usr.Events()
	require.Len(t, events, 1)
	deactivated, ok := events[0].(user.UserDeactivated)
	require.True(t, ok)
	assert.True(t, deactivated.UserID().Equal(usr.ID()))
	assert.Equal(t, eventTime, deactivated.OccurredAt())

	err = usr.Deactivate(eventTime.Add(time.Hour))
	require.Error(t, err, "Deactivate on already-inactive user must be rejected")
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestUserTouchLogin(t *testing.T) {
	input := validUserInput(t)
	usr, _ := user.NewUser(input.tenantID, input.email, input.displayName, input.role, input.hash, false, input.createdAt)
	usr.ClearEvents()

	eventTime := testTime.Add(time.Hour)
	require.NoError(t, usr.TouchLogin(eventTime))
	assert.Equal(t, eventTime, usr.LastLoginAt())
	assert.Empty(t, usr.Events(), "TouchLogin must not record events")

	require.Error(t, usr.TouchLogin(time.Time{}), "zero occurredAt must be rejected")
}
