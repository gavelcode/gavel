package revoketoken

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

func TestExecuteReturnsErrorOnInvalidTokenID(t *testing.T) {
	handler := NewHandler(&stubTokenRepo{})
	cmd := Command{tokenID: "not-a-uuid", callerUserID: uuid.NewString(), occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrInvalid)
}

func TestExecuteReturnsErrorOnInvalidCallerUserID(t *testing.T) {
	handler := NewHandler(&stubTokenRepo{})
	cmd := Command{tokenID: uuid.NewString(), callerUserID: "not-a-uuid", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestExecuteReturnsErrorOnTokenSaveFailure(t *testing.T) {
	userID := user.NewUserID(uuid.New())
	tok := seedInternalToken(t, userID)
	repo := &stubTokenRepo{token: tok, saveErr: errors.New("save broken")}
	handler := NewHandler(repo)

	cmd := Command{tokenID: tok.ID().String(), callerUserID: userID.String(), occurredAt: internalTestTime.Add(time.Minute)}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save token")
}

func seedInternalToken(t *testing.T, userID user.UserID) apitoken.APIToken {
	t.Helper()
	secret, _ := apitoken.NewSecret("gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	tenantID := tenant.NewTenantID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeRead}
	tok, err := apitoken.NewAPIToken(secret, tenantID, userID, "ci-bot", scopes, internalTestTime, internalTestTime.Add(time.Hour))
	require.NoError(t, err)
	tok.ClearEvents()
	return tok
}

type stubTokenRepo struct {
	token   apitoken.APIToken
	saveErr error
}

func (r *stubTokenRepo) Save(_ context.Context, _ apitoken.APIToken) error { return r.saveErr }
func (r *stubTokenRepo) ByID(_ context.Context, _ apitoken.APITokenID) (apitoken.APIToken, error) {
	return r.token, nil
}
func (r *stubTokenRepo) ByTokenHash(_ context.Context, _ apitoken.SecretHash) (apitoken.APIToken, error) {
	return r.token, nil
}
func (r *stubTokenRepo) ListByUser(_ context.Context, _ user.UserID) ([]apitoken.APIToken, error) {
	return nil, nil
}
