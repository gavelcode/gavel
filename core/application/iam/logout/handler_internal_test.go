package logout

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

var internalTestTime = time.Date(2026, time.June, 18, 12, 0, 0, 0, time.UTC)

func TestNewHandlerPanicsOnNilSessions(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(nil)
	})
}

func TestExecuteReturnsErrorOnInvalidSessionToken(t *testing.T) {
	handler := NewHandler(&stubSessionRepo{})
	cmd := Command{sessionToken: "too-short", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse session token")
}

func TestExecuteReturnsErrorOnSessionSaveFailure(t *testing.T) {
	userID := user.NewUserID(uuid.New())
	tok, _ := session.NewToken("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	sess, err := session.NewSession(tok, userID, "ua", "ip", internalTestTime, internalTestTime.Add(time.Hour))
	require.NoError(t, err)
	sess.ClearEvents()

	handler := NewHandler(&stubSessionRepo{session: sess, saveErr: errors.New("save broken")})
	cmd := Command{sessionToken: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", occurredAt: internalTestTime.Add(time.Minute)}
	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save session")
}

type stubSessionRepo struct {
	session session.Session
	saveErr error
}

func (r *stubSessionRepo) Save(_ context.Context, _ session.Session) error { return r.saveErr }
func (r *stubSessionRepo) ByTokenHash(_ context.Context, _ session.TokenHash) (session.Session, error) {
	return r.session, nil
}
func (r *stubSessionRepo) DeleteAllForUser(_ context.Context, _ user.UserID) error { return nil }
func (r *stubSessionRepo) DeleteExpired(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}
