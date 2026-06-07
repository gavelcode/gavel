package deactivateuser_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/deactivateuser"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

const validSession = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

func seedUser(t *testing.T, users *memiam.UserRepository) user.User {
	t.Helper()
	tenantID := tenant.NewTenantID(uuid.New())
	email, _ := user.NewEmail("alice@example.com")
	hash, _ := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	foundUser, err := user.NewUser(tenantID, email, "Alice", user.RoleAdmin, hash, false, testTime)
	require.NoError(t, err)
	foundUser.ClearEvents()
	require.NoError(t, users.Save(context.Background(), foundUser))
	return foundUser
}

func TestExecuteDeactivatesUserAndDropsSessions(t *testing.T) {
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	foundUser := seedUser(t, users)

	tok, _ := session.NewToken(validSession)
	sess, _ := session.NewSession(tok, foundUser.ID(), "ua", "ip", testTime, testTime.Add(time.Hour))
	sess.ClearEvents()
	require.NoError(t, sessions.Save(context.Background(), sess))

	cmd, _ := deactivateuser.NewCommand(foundUser.ID().String(), testTime.Add(time.Hour))
	result, err := deactivateuser.NewHandler(users, sessions).Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, foundUser.ID().String(), result.UserID)
	require.Len(t, result.Events, 1)
	assert.Equal(t, user.EventNameUserDeactivated, result.Events[0].Name)

	got, _ := users.ByID(context.Background(), foundUser.ID())
	assert.False(t, got.IsActive())

	_, err = sessions.ByTokenHash(context.Background(), sess.TokenHash())
	require.Error(t, err, "session must be revoked along with the user")
}

func TestExecuteRejectsMissingUser(t *testing.T) {
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	cmd, _ := deactivateuser.NewCommand(uuid.NewString(), testTime.Add(time.Hour))
	_, err := deactivateuser.NewHandler(users, sessions).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestExecuteRejectsAlreadyInactive(t *testing.T) {
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	foundUser := seedUser(t, users)

	cmd, _ := deactivateuser.NewCommand(foundUser.ID().String(), testTime.Add(time.Hour))
	_, err := deactivateuser.NewHandler(users, sessions).Execute(context.Background(), cmd)
	require.NoError(t, err)

	_, err = deactivateuser.NewHandler(users, sessions).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	_, err := deactivateuser.NewCommand("", testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, deactivateuser.ErrInvalidCommand)

	_, err = deactivateuser.NewCommand("foundUser-1", time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, deactivateuser.ErrInvalidCommand)
}
