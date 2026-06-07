package logout_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/logout"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 18, 12, 0, 0, 0, time.UTC)
	testUser = user.NewUserID(uuid.New())
)

func seedSession(t *testing.T, sessions *memiam.SessionRepository) (session.Session, session.Token) {
	t.Helper()
	secrets := memiam.NewFakeSecretGenerator()
	token, err := secrets.NewSessionToken()
	require.NoError(t, err)

	sess, err := session.NewSession(token, testUser, "test-agent", "127.0.0.1", testTime, testTime.Add(24*time.Hour))
	require.NoError(t, err)
	sess.ClearEvents()
	require.NoError(t, sessions.Save(context.Background(), sess))
	return sess, token
}

func TestShouldRevokeSessionByTokenHash(t *testing.T) {
	sessions := memiam.NewSessionRepository()
	sess, token := seedSession(t, sessions)

	cmd, err := logout.NewCommand(token.String(), testTime.Add(time.Minute))
	require.NoError(t, err)

	result, err := logout.NewHandler(sessions).Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, sess.ID().String(), result.SessionID)
	require.Len(t, result.Events, 1)
	assert.Equal(t, session.EventNameRevoked, result.Events[0].Name)

	got, err := sessions.ByTokenHash(context.Background(), session.HashToken(token))
	require.NoError(t, err)
	assert.True(t, got.IsRevoked())
}

func TestShouldReturnErrorWhenSessionNotFound(t *testing.T) {
	sessions := memiam.NewSessionRepository()
	secrets := memiam.NewFakeSecretGenerator()
	token, err := secrets.NewSessionToken()
	require.NoError(t, err)

	cmd, err := logout.NewCommand(token.String(), testTime)
	require.NoError(t, err)

	_, err = logout.NewHandler(sessions).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrNotFound)
}

func TestShouldReturnErrorWhenAlreadyRevoked(t *testing.T) {
	sessions := memiam.NewSessionRepository()
	_, token := seedSession(t, sessions)

	cmd, err := logout.NewCommand(token.String(), testTime.Add(time.Minute))
	require.NoError(t, err)

	h := logout.NewHandler(sessions)
	_, err = h.Execute(context.Background(), cmd)
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrInvalid)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	_, err := logout.NewCommand("", testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, logout.ErrInvalidCommand)

	_, err = logout.NewCommand("some-token", time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, logout.ErrInvalidCommand)
}
