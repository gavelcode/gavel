package listmytokens

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

var internalTestTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func TestNewHandlerPanicsOnNilTokens(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(nil)
	})
}

func TestExecuteReturnsErrorOnInvalidUserID(t *testing.T) {
	handler := NewHandler(&stubTokenRepo{})
	q := Query{userID: "not-a-uuid", now: internalTestTime}
	_, err := handler.Execute(context.Background(), q)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestExecuteReturnsErrorOnListByUserFailure(t *testing.T) {
	handler := NewHandler(&stubTokenRepo{listErr: errors.New("db broken")})
	userID := user.NewUserID(uuid.New())
	q := Query{userID: userID.String(), now: internalTestTime}
	_, err := handler.Execute(context.Background(), q)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list tokens")
}

func TestExecutePopulatesLastUsedAtAndExpiresAt(t *testing.T) {
	tenantID := tenant.NewTenantID(uuid.New())
	userID := user.NewUserID(uuid.New())
	tokenID := apitoken.NewAPITokenID(uuid.New())
	secret, _ := apitoken.NewSecret("gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	hash := apitoken.HashSecret(secret)
	scopes := apitoken.Scopes{apitoken.ScopeRead}
	expiresAt := internalTestTime.Add(24 * time.Hour)
	lastUsed := internalTestTime.Add(time.Hour)

	tok, err := apitoken.ReconstituteAPIToken(
		tokenID, tenantID, userID, "ci-bot", hash, "gav_AAAA", scopes,
		internalTestTime, expiresAt, lastUsed, false,
	)
	require.NoError(t, err)

	handler := NewHandler(&stubTokenRepo{tokens: []apitoken.APIToken{tok}})
	q := Query{userID: userID.String(), now: internalTestTime.Add(2 * time.Hour)}
	res, err := handler.Execute(context.Background(), q)
	require.NoError(t, err)
	require.Len(t, res.Tokens, 1)

	s := res.Tokens[0]
	require.NotNil(t, s.LastUsedAt)
	assert.Equal(t, lastUsed, *s.LastUsedAt)
	require.NotNil(t, s.ExpiresAt)
	assert.Equal(t, expiresAt, *s.ExpiresAt)
}

type stubTokenRepo struct {
	tokens  []apitoken.APIToken
	listErr error
}

func (r *stubTokenRepo) Save(_ context.Context, _ apitoken.APIToken) error { return nil }
func (r *stubTokenRepo) ByID(_ context.Context, _ apitoken.APITokenID) (apitoken.APIToken, error) {
	return apitoken.APIToken{}, nil
}
func (r *stubTokenRepo) ByTokenHash(_ context.Context, _ apitoken.SecretHash) (apitoken.APIToken, error) {
	return apitoken.APIToken{}, nil
}
func (r *stubTokenRepo) ListByUser(_ context.Context, _ user.UserID) ([]apitoken.APIToken, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.tokens, nil
}
