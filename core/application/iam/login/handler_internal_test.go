package login

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

var internalTestTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func TestNewHandlerPanicsOnNilTenants(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(nil, &stubUserRepo{}, &stubSessionRepo{}, &stubHasher{}, &stubSecrets{})
	})
}

func TestNewHandlerPanicsOnNilUsers(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubTenantRepo{}, nil, &stubSessionRepo{}, &stubHasher{}, &stubSecrets{})
	})
}

func TestNewHandlerPanicsOnNilSessions(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubTenantRepo{}, &stubUserRepo{}, nil, &stubHasher{}, &stubSecrets{})
	})
}

func TestNewHandlerPanicsOnNilHasher(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubTenantRepo{}, &stubUserRepo{}, &stubSessionRepo{}, nil, &stubSecrets{})
	})
}

func TestNewHandlerPanicsOnNilSecrets(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubTenantRepo{}, &stubUserRepo{}, &stubSessionRepo{}, &stubHasher{}, nil)
	})
}

func TestExecuteReturnsErrInvalidCredentialsOnBadSlug(t *testing.T) {
	tn := seedInternalTenant(t)
	u := seedInternalUser(t, tn.ID())
	handler := NewHandler(
		&stubTenantRepo{tenant: tn},
		&stubUserRepo{user: u},
		&stubSessionRepo{},
		&stubHasher{verifyOK: true},
		&stubSecrets{},
	)
	cmd := Command{tenantSlug: "INVALID SLUG!", email: "alice@example.com", plainPassword: "hunter22", occurredAt: internalTestTime, sessionTTL: time.Hour}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestExecuteReturnsErrInvalidCredentialsOnBadEmail(t *testing.T) {
	tn := seedInternalTenant(t)
	u := seedInternalUser(t, tn.ID())
	handler := NewHandler(
		&stubTenantRepo{tenant: tn},
		&stubUserRepo{user: u},
		&stubSessionRepo{},
		&stubHasher{verifyOK: true},
		&stubSecrets{},
	)
	cmd := Command{tenantSlug: "acme", email: "not-an-email", plainPassword: "hunter22", occurredAt: internalTestTime, sessionTTL: time.Hour}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestExecuteReturnsErrorOnTenantRepoInfraFailure(t *testing.T) {
	handler := NewHandler(
		&stubTenantRepo{err: errors.New("db broken")},
		&stubUserRepo{},
		&stubSessionRepo{},
		&stubHasher{},
		&stubSecrets{},
	)
	cmd := Command{tenantSlug: "acme", email: "alice@example.com", plainPassword: "hunter22", occurredAt: internalTestTime, sessionTTL: time.Hour}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "find tenant")
}

func TestExecuteReturnsErrorOnUserRepoInfraFailure(t *testing.T) {
	tn := seedInternalTenant(t)
	handler := NewHandler(
		&stubTenantRepo{tenant: tn},
		&stubUserRepo{err: errors.New("db broken")},
		&stubSessionRepo{},
		&stubHasher{},
		&stubSecrets{},
	)
	cmd := Command{tenantSlug: "acme", email: "alice@example.com", plainPassword: "hunter22", occurredAt: internalTestTime, sessionTTL: time.Hour}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "find user")
}

func TestExecuteReturnsErrorOnHasherVerifyFailure(t *testing.T) {
	tn := seedInternalTenant(t)
	u := seedInternalUser(t, tn.ID())
	handler := NewHandler(
		&stubTenantRepo{tenant: tn},
		&stubUserRepo{user: u},
		&stubSessionRepo{},
		&stubHasher{verifyErr: errors.New("hasher broken")},
		&stubSecrets{},
	)
	cmd := Command{tenantSlug: "acme", email: "alice@example.com", plainPassword: "hunter22", occurredAt: internalTestTime, sessionTTL: time.Hour}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "verify password")
}

func TestExecuteReturnsErrorOnSecretGeneratorFailure(t *testing.T) {
	tn := seedInternalTenant(t)
	u := seedInternalUser(t, tn.ID())
	handler := NewHandler(
		&stubTenantRepo{tenant: tn},
		&stubUserRepo{user: u},
		&stubSessionRepo{},
		&stubHasher{verifyOK: true},
		&stubSecrets{tokenErr: errors.New("entropy depleted")},
	)
	cmd := Command{tenantSlug: "acme", email: "alice@example.com", plainPassword: "hunter22", occurredAt: internalTestTime, sessionTTL: time.Hour}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "generate session token")
}

func TestExecuteReturnsErrorOnSessionSaveFailure(t *testing.T) {
	tn := seedInternalTenant(t)
	u := seedInternalUser(t, tn.ID())
	tok, _ := session.NewToken("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	handler := NewHandler(
		&stubTenantRepo{tenant: tn},
		&stubUserRepo{user: u},
		&stubSessionRepo{saveErr: errors.New("save broken")},
		&stubHasher{verifyOK: true},
		&stubSecrets{token: tok},
	)
	cmd := Command{tenantSlug: "acme", email: "alice@example.com", plainPassword: "hunter22", occurredAt: internalTestTime, sessionTTL: time.Hour}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save session")
}

func TestExecuteReturnsErrorOnNewSessionFailure(t *testing.T) {
	tn := seedInternalTenant(t)
	u := seedInternalUser(t, tn.ID())
	tok, _ := session.NewToken("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	handler := NewHandler(
		&stubTenantRepo{tenant: tn},
		&stubUserRepo{user: u},
		&stubSessionRepo{},
		&stubHasher{verifyOK: true},
		&stubSecrets{token: tok},
	)
	cmd := Command{tenantSlug: "acme", email: "alice@example.com", plainPassword: "hunter22", occurredAt: time.Time{}, sessionTTL: time.Hour}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new session")
}


func TestExecuteReturnsErrorOnUserSaveFailure(t *testing.T) {
	tn := seedInternalTenant(t)
	u := seedInternalUser(t, tn.ID())
	tok, _ := session.NewToken("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	handler := NewHandler(
		&stubTenantRepo{tenant: tn},
		&stubUserRepo{user: u, saveErr: errors.New("save broken")},
		&stubSessionRepo{},
		&stubHasher{verifyOK: true},
		&stubSecrets{token: tok},
	)
	cmd := Command{tenantSlug: "acme", email: "alice@example.com", plainPassword: "hunter22", occurredAt: internalTestTime, sessionTTL: time.Hour}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save user")
}

func seedInternalTenant(t *testing.T) tenant.Tenant {
	t.Helper()
	slug, err := tenant.NewSlug("acme")
	require.NoError(t, err)
	tn, err := tenant.NewTenant(slug, "Acme", internalTestTime)
	require.NoError(t, err)
	tn.ClearEvents()
	return tn
}

func seedInternalUser(t *testing.T, tenantID tenant.TenantID) user.User {
	t.Helper()
	email, _ := user.NewEmail("alice@example.com")
	hash, _ := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	u, err := user.NewUser(tenantID, email, "Alice", user.RoleAdmin, hash, false, internalTestTime)
	require.NoError(t, err)
	u.ClearEvents()
	return u
}

type stubTenantRepo struct {
	tenant tenant.Tenant
	err    error
}

func (r *stubTenantRepo) Save(_ context.Context, _ tenant.Tenant) error { return nil }
func (r *stubTenantRepo) ByID(_ context.Context, _ tenant.TenantID) (tenant.Tenant, error) {
	return r.tenant, r.err
}
func (r *stubTenantRepo) BySlug(_ context.Context, _ tenant.Slug) (tenant.Tenant, error) {
	if r.err != nil {
		return tenant.Tenant{}, r.err
	}
	return r.tenant, nil
}

type stubUserRepo struct {
	user    user.User
	err     error
	saveErr error
}

func (r *stubUserRepo) Save(_ context.Context, _ user.User) error { return r.saveErr }
func (r *stubUserRepo) ByID(_ context.Context, _ user.UserID) (user.User, error) {
	return r.user, r.err
}
func (r *stubUserRepo) ByEmail(_ context.Context, _ tenant.TenantID, _ user.Email) (user.User, error) {
	if r.err != nil {
		return user.User{}, r.err
	}
	return r.user, nil
}
func (r *stubUserRepo) CountByTenant(_ context.Context, _ tenant.TenantID) (int, error) {
	return 0, nil
}

type stubSessionRepo struct {
	saveErr error
}

func (r *stubSessionRepo) Save(_ context.Context, _ session.Session) error { return r.saveErr }
func (r *stubSessionRepo) ByTokenHash(_ context.Context, _ session.TokenHash) (session.Session, error) {
	return session.Session{}, nil
}
func (r *stubSessionRepo) DeleteAllForUser(_ context.Context, _ user.UserID) error { return nil }
func (r *stubSessionRepo) DeleteExpired(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

type stubHasher struct {
	verifyOK  bool
	verifyErr error
}

func (h *stubHasher) Hash(_ string) (user.PasswordHash, error) { return user.PasswordHash{}, nil }
func (h *stubHasher) Verify(_ string, _ user.PasswordHash) (bool, error) {
	return h.verifyOK, h.verifyErr
}

type stubSecrets struct {
	token    session.Token
	tokenErr error
}

func (s *stubSecrets) NewSessionToken() (session.Token, error) { return s.token, s.tokenErr }
func (s *stubSecrets) NewAPITokenSecret() (apitoken.Secret, error) {
	return apitoken.Secret{}, nil
}
