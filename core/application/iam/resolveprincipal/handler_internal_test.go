package resolveprincipal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

var internalTestTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func TestNewHandlerPanicsOnNilUsers(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(nil, &stubSessionRepo{}, &stubTokenRepo{})
	})
}

func TestNewHandlerPanicsOnNilSessions(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubUserRepo{}, nil, &stubTokenRepo{})
	})
}

func TestNewHandlerPanicsOnNilTokens(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubUserRepo{}, &stubSessionRepo{}, nil)
	})
}

func TestResolveBearerReturnsErrorOnTokenRepoInfraFailure(t *testing.T) {
	handler := NewHandler(
		&stubUserRepo{},
		&stubSessionRepo{},
		&stubTokenRepo{byHashErr: errors.New("db broken")},
	)
	q := Query{bearerToken: "gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), q)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "find token")
}

func TestResolveBearerReturnsErrorOnUserRepoInfraFailure(t *testing.T) {
	secret, _ := apitoken.NewSecret("gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	hash := apitoken.HashSecret(secret)
	tenantID := tenant.NewTenantID(uuid.New())
	userID := user.NewUserID(uuid.New())
	tokenID := apitoken.NewAPITokenID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeRead}
	expiresAt := internalTestTime.Add(24 * time.Hour)

	tok, err := apitoken.ReconstituteAPIToken(tokenID, tenantID, userID, "ci-bot", hash, "gav_AAAA", scopes, internalTestTime, expiresAt, time.Time{}, false)
	require.NoError(t, err)

	handler := NewHandler(
		&stubUserRepo{err: errors.New("db broken")},
		&stubSessionRepo{},
		&stubTokenRepo{token: tok},
	)
	q := Query{bearerToken: "gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", occurredAt: internalTestTime.Add(time.Minute)}
	_, err = handler.Execute(context.Background(), q)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "find token user")
}

func TestResolveBearerRejectsInactiveUser(t *testing.T) {
	secret, _ := apitoken.NewSecret("gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	hash := apitoken.HashSecret(secret)
	tenantID := tenant.NewTenantID(uuid.New())
	userID := user.NewUserID(uuid.New())
	tokenID := apitoken.NewAPITokenID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeRead}
	expiresAt := internalTestTime.Add(24 * time.Hour)

	tok, err := apitoken.ReconstituteAPIToken(tokenID, tenantID, userID, "ci-bot", hash, "gav_AAAA", scopes, internalTestTime, expiresAt, time.Time{}, false)
	require.NoError(t, err)

	foundUser := seedInternalUser(t, tenantID, true)
	require.NoError(t, foundUser.Deactivate(internalTestTime.Add(time.Minute)))
	foundUser.ClearEvents()

	handler := NewHandler(
		&stubUserRepo{user: foundUser},
		&stubSessionRepo{},
		&stubTokenRepo{token: tok},
	)
	q := Query{bearerToken: "gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", occurredAt: internalTestTime.Add(2 * time.Minute)}
	_, err = handler.Execute(context.Background(), q)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthenticated)
}

func TestResolveCookieReturnsErrorOnSessionRepoInfraFailure(t *testing.T) {
	handler := NewHandler(
		&stubUserRepo{},
		&stubSessionRepo{byHashErr: errors.New("db broken")},
		&stubTokenRepo{},
	)
	q := Query{sessionCookie: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), q)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "find session")
}

func TestResolveCookieReturnsErrorOnUserRepoInfraFailure(t *testing.T) {
	tok, _ := session.NewToken("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	userID := user.NewUserID(uuid.New())
	sess, err := session.NewSession(tok, userID, "ua", "ip", internalTestTime, internalTestTime.Add(time.Hour))
	require.NoError(t, err)
	sess.ClearEvents()

	handler := NewHandler(
		&stubUserRepo{err: errors.New("db broken")},
		&stubSessionRepo{session: sess},
		&stubTokenRepo{},
	)
	q := Query{sessionCookie: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", occurredAt: internalTestTime.Add(time.Minute)}
	_, err = handler.Execute(context.Background(), q)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "find session user")
}

func seedInternalUser(t *testing.T, tenantID tenant.TenantID, active bool) user.User {
	t.Helper()
	email, _ := user.NewEmail("alice@example.com")
	hash, _ := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	foundUser, err := user.NewUser(tenantID, email, "Alice", user.RoleAdmin, hash, false, internalTestTime)
	require.NoError(t, err)
	foundUser.ClearEvents()
	if !active {
		require.NoError(t, foundUser.Deactivate(internalTestTime.Add(time.Minute)))
		foundUser.ClearEvents()
	}
	return foundUser
}

type stubUserRepo struct {
	user user.User
	err  error
}

func (r *stubUserRepo) Save(_ context.Context, _ user.User) error { return nil }
func (r *stubUserRepo) ByID(_ context.Context, _ user.UserID) (user.User, error) {
	if r.err != nil {
		return user.User{}, r.err
	}
	return r.user, nil
}
func (r *stubUserRepo) ByEmail(_ context.Context, _ tenant.TenantID, _ user.Email) (user.User, error) {
	return r.user, r.err
}
func (r *stubUserRepo) CountByTenant(_ context.Context, _ tenant.TenantID) (int, error) {
	return 0, nil
}

type stubSessionRepo struct {
	session   session.Session
	byHashErr error
}

func (r *stubSessionRepo) Save(_ context.Context, _ session.Session) error { return nil }
func (r *stubSessionRepo) ByTokenHash(_ context.Context, _ session.TokenHash) (session.Session, error) {
	if r.byHashErr != nil {
		return session.Session{}, r.byHashErr
	}
	return r.session, nil
}
func (r *stubSessionRepo) DeleteAllForUser(_ context.Context, _ user.UserID) error { return nil }
func (r *stubSessionRepo) DeleteExpired(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

type stubTokenRepo struct {
	token     apitoken.APIToken
	byHashErr error
}

func (r *stubTokenRepo) Save(_ context.Context, _ apitoken.APIToken) error { return nil }
func (r *stubTokenRepo) ByID(_ context.Context, _ apitoken.APITokenID) (apitoken.APIToken, error) {
	return r.token, nil
}
func (r *stubTokenRepo) ByTokenHash(_ context.Context, _ apitoken.SecretHash) (apitoken.APIToken, error) {
	if r.byHashErr != nil {
		return apitoken.APIToken{}, r.byHashErr
	}
	return r.token, nil
}
func (r *stubTokenRepo) ListByUser(_ context.Context, _ user.UserID) ([]apitoken.APIToken, error) {
	return nil, nil
}
