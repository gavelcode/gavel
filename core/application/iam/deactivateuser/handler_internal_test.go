package deactivateuser

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

var internalTestTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func TestNewHandlerPanicsOnNilUsers(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(nil, &stubSessionRepo{})
	})
}

func TestNewHandlerPanicsOnNilSessions(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubUserRepo{}, nil)
	})
}

func TestExecuteReturnsErrorOnInvalidUserID(t *testing.T) {
	handler := NewHandler(&stubUserRepo{}, &stubSessionRepo{})
	cmd := Command{userID: "not-a-uuid", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestExecuteReturnsErrorOnUserSaveFailure(t *testing.T) {
	u := seedInternalUser(t)
	users := &stubUserRepo{user: u, saveErr: errors.New("save broken")}
	handler := NewHandler(users, &stubSessionRepo{})

	cmd := Command{userID: u.ID().String(), occurredAt: internalTestTime.Add(time.Hour)}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save user")
}

func TestExecuteReturnsErrorOnSessionDeleteFailure(t *testing.T) {
	u := seedInternalUser(t)
	users := &stubUserRepo{user: u}
	sessions := &stubSessionRepo{deleteErr: errors.New("delete broken")}
	handler := NewHandler(users, sessions)

	cmd := Command{userID: u.ID().String(), occurredAt: internalTestTime.Add(time.Hour)}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete sessions")
}

func seedInternalUser(t *testing.T) user.User {
	t.Helper()
	tenantID := tenant.NewTenantID(uuid.New())
	email, _ := user.NewEmail("alice@example.com")
	hash, _ := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	u, err := user.NewUser(tenantID, email, "Alice", user.RoleAdmin, hash, false, internalTestTime)
	require.NoError(t, err)
	u.ClearEvents()
	return u
}

type stubUserRepo struct {
	user    user.User
	saveErr error
}

func (r *stubUserRepo) Save(_ context.Context, _ user.User) error { return r.saveErr }
func (r *stubUserRepo) ByID(_ context.Context, _ user.UserID) (user.User, error) {
	return r.user, nil
}
func (r *stubUserRepo) ByEmail(_ context.Context, _ tenant.TenantID, _ user.Email) (user.User, error) {
	return r.user, nil
}
func (r *stubUserRepo) CountByTenant(_ context.Context, _ tenant.TenantID) (int, error) {
	return 0, nil
}

type stubSessionRepo struct {
	deleteErr error
}

func (r *stubSessionRepo) Save(_ context.Context, _ session.Session) error { return nil }
func (r *stubSessionRepo) ByTokenHash(_ context.Context, _ session.TokenHash) (session.Session, error) {
	return session.Session{}, nil
}
func (r *stubSessionRepo) DeleteAllForUser(_ context.Context, _ user.UserID) error {
	return r.deleteErr
}
func (r *stubSessionRepo) DeleteExpired(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}
