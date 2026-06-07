package changepassword

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
		NewHandler(nil, &stubSessionRepo{}, &stubHasher{})
	})
}

func TestNewHandlerPanicsOnNilSessions(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubUserRepo{}, nil, &stubHasher{})
	})
}

func TestNewHandlerPanicsOnNilHasher(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubUserRepo{}, &stubSessionRepo{}, nil)
	})
}

func TestExecuteReturnsErrorOnInvalidUserID(t *testing.T) {
	handler := NewHandler(&stubUserRepo{}, &stubSessionRepo{}, &stubHasher{})
	cmd := Command{userID: "not-a-uuid", currentPassword: "old12345", newPassword: "new12345!", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestExecuteReturnsErrorOnHasherVerifyFailure(t *testing.T) {
	foundUser := seedInternalUser(t)
	users := &stubUserRepo{user: foundUser}
	hasher := &stubHasher{verifyErr: errors.New("hasher broken")}
	handler := NewHandler(users, &stubSessionRepo{}, hasher)

	cmd := Command{userID: foundUser.ID().String(), currentPassword: "old12345", newPassword: "new12345!", occurredAt: internalTestTime.Add(time.Minute)}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "verify password")
}

func TestExecuteReturnsErrorOnHasherHashFailure(t *testing.T) {
	foundUser := seedInternalUser(t)
	users := &stubUserRepo{user: foundUser}
	hasher := &stubHasher{verifyOK: true, hashErr: errors.New("hash broken")}
	handler := NewHandler(users, &stubSessionRepo{}, hasher)

	cmd := Command{userID: foundUser.ID().String(), currentPassword: "old12345", newPassword: "new12345!", occurredAt: internalTestTime.Add(time.Minute)}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash password")
}

func TestExecuteReturnsErrorOnChangePasswordDomainFailure(t *testing.T) {
	foundUser := seedInternalUser(t)
	users := &stubUserRepo{user: foundUser}
	hasher := &stubHasher{verifyOK: true, hashResult: foundUser.PasswordHash()}
	handler := NewHandler(users, &stubSessionRepo{}, hasher)

	cmd := Command{userID: foundUser.ID().String(), currentPassword: "old12345", newPassword: "new12345!", occurredAt: time.Time{}}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "change password")
}

func TestExecuteReturnsErrorOnUserSaveFailure(t *testing.T) {
	foundUser := seedInternalUser(t)
	users := &stubUserRepo{user: foundUser, saveErr: errors.New("save broken")}
	hasher := &stubHasher{verifyOK: true, hashResult: foundUser.PasswordHash()}
	handler := NewHandler(users, &stubSessionRepo{}, hasher)

	cmd := Command{userID: foundUser.ID().String(), currentPassword: "old12345", newPassword: "new12345!", occurredAt: internalTestTime.Add(time.Minute)}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save user")
}

func TestExecuteReturnsErrorOnSessionDeleteFailure(t *testing.T) {
	foundUser := seedInternalUser(t)
	users := &stubUserRepo{user: foundUser}
	sessions := &stubSessionRepo{deleteErr: errors.New("delete broken")}
	hasher := &stubHasher{verifyOK: true, hashResult: foundUser.PasswordHash()}
	handler := NewHandler(users, sessions, hasher)

	cmd := Command{userID: foundUser.ID().String(), currentPassword: "old12345", newPassword: "new12345!", occurredAt: internalTestTime.Add(time.Minute)}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete sessions")
}

func seedInternalUser(t *testing.T) user.User {
	t.Helper()
	tenantID := tenant.NewTenantID(uuid.New())
	email, _ := user.NewEmail("alice@example.com")
	hash, _ := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	foundUser, err := user.NewUser(tenantID, email, "Alice", user.RoleAdmin, hash, false, internalTestTime)
	require.NoError(t, err)
	foundUser.ClearEvents()
	return foundUser
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

type stubHasher struct {
	verifyOK   bool
	verifyErr  error
	hashResult user.PasswordHash
	hashErr    error
}

func (h *stubHasher) Hash(_ string) (user.PasswordHash, error) { return h.hashResult, h.hashErr }
func (h *stubHasher) Verify(_ string, _ user.PasswordHash) (bool, error) {
	return h.verifyOK, h.verifyErr
}
