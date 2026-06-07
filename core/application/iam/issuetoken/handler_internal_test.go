package issuetoken

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
		NewHandler(nil, &stubTokenRepo{}, &stubSecrets{})
	})
}

func TestNewHandlerPanicsOnNilTokens(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubUserRepo{}, nil, &stubSecrets{})
	})
}

func TestNewHandlerPanicsOnNilSecrets(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubUserRepo{}, &stubTokenRepo{}, nil)
	})
}

func TestExecuteReturnsErrorOnInvalidUserID(t *testing.T) {
	handler := NewHandler(&stubUserRepo{}, &stubTokenRepo{}, &stubSecrets{})
	cmd := Command{userID: "not-a-uuid", name: "ci", scopes: []string{"read"}, occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestExecuteReturnsErrorOnInvalidScope(t *testing.T) {
	foundUser := seedInternalUser(t)
	handler := NewHandler(&stubUserRepo{user: foundUser}, &stubTokenRepo{}, &stubSecrets{})
	cmd := Command{userID: foundUser.ID().String(), name: "ci", scopes: []string{"nuke"}, occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrInvalid)
}

func TestExecuteReturnsErrorOnSecretGeneratorFailure(t *testing.T) {
	foundUser := seedInternalUser(t)
	secrets := &stubSecrets{err: errors.New("entropy depleted")}
	handler := NewHandler(&stubUserRepo{user: foundUser}, &stubTokenRepo{}, secrets)

	cmd := Command{userID: foundUser.ID().String(), name: "ci", scopes: []string{"read"}, occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "generate api token secret")
}

func TestExecuteReturnsErrorOnNewAPITokenDomainFailure(t *testing.T) {
	foundUser := seedInternalUser(t)
	secret, _ := apitoken.NewSecret("gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	secrets := &stubSecrets{secret: secret}
	handler := NewHandler(&stubUserRepo{user: foundUser}, &stubTokenRepo{}, secrets)

	cmd := Command{userID: foundUser.ID().String(), name: "", scopes: []string{"read"}, occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new api token")
}

func TestExecuteReturnsErrorOnTokenSaveFailure(t *testing.T) {
	foundUser := seedInternalUser(t)
	secret, _ := apitoken.NewSecret("gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	secrets := &stubSecrets{secret: secret}
	tokenRepo := &stubTokenRepo{saveErr: errors.New("save broken")}
	handler := NewHandler(&stubUserRepo{user: foundUser}, tokenRepo, secrets)

	cmd := Command{userID: foundUser.ID().String(), name: "ci", scopes: []string{"read"}, occurredAt: internalTestTime, expiresAt: internalTestTime.Add(time.Hour)}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save api token")
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
	user user.User
}

func (r *stubUserRepo) Save(_ context.Context, _ user.User) error { return nil }
func (r *stubUserRepo) ByID(_ context.Context, _ user.UserID) (user.User, error) {
	return r.user, nil
}
func (r *stubUserRepo) ByEmail(_ context.Context, _ tenant.TenantID, _ user.Email) (user.User, error) {
	return r.user, nil
}
func (r *stubUserRepo) CountByTenant(_ context.Context, _ tenant.TenantID) (int, error) {
	return 0, nil
}

type stubTokenRepo struct {
	saveErr error
}

func (r *stubTokenRepo) Save(_ context.Context, _ apitoken.APIToken) error { return r.saveErr }
func (r *stubTokenRepo) ByID(_ context.Context, _ apitoken.APITokenID) (apitoken.APIToken, error) {
	return apitoken.APIToken{}, nil
}
func (r *stubTokenRepo) ByTokenHash(_ context.Context, _ apitoken.SecretHash) (apitoken.APIToken, error) {
	return apitoken.APIToken{}, nil
}
func (r *stubTokenRepo) ListByUser(_ context.Context, _ user.UserID) ([]apitoken.APIToken, error) {
	return nil, nil
}

type stubSecrets struct {
	secret apitoken.Secret
	err    error
}

func (s *stubSecrets) NewAPITokenSecret() (apitoken.Secret, error) { return s.secret, s.err }
func (s *stubSecrets) NewSessionToken() (session.Token, error)     { return session.Token{}, nil }
