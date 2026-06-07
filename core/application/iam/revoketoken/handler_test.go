package revoketoken_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/revoketoken"
	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
	userA    = user.NewUserID(uuid.New())
	userB    = user.NewUserID(uuid.New())
)

const validSecret = "gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

func seedToken(t *testing.T, tokens *memiam.APITokenRepository, userID user.UserID) apitoken.APIToken {
	t.Helper()
	secret, _ := apitoken.NewSecret(validSecret)
	tenantID := tenant.NewTenantID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeRead}
	tok, err := apitoken.NewAPIToken(secret, tenantID, userID, "ci-bot", scopes, testTime, testTime.Add(time.Hour))
	require.NoError(t, err)
	tok.ClearEvents()
	require.NoError(t, tokens.Save(context.Background(), tok))
	return tok
}

func TestExecuteRevokesOwnToken(t *testing.T) {
	tokens := memiam.NewAPITokenRepository()
	tok := seedToken(t, tokens, userA)

	cmd, err := revoketoken.NewCommand(tok.ID().String(), userA.String(), testTime.Add(time.Minute))
	require.NoError(t, err)

	result, err := revoketoken.NewHandler(tokens).Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, tok.ID().String(), result.TokenID)
	require.Len(t, result.Events, 1)
	assert.Equal(t, apitoken.EventNameRevoked, result.Events[0].Name)

	got, _ := tokens.ByID(context.Background(), tok.ID())
	assert.True(t, got.IsRevoked())
}

func TestExecuteRejectsForeignToken(t *testing.T) {
	tokens := memiam.NewAPITokenRepository()
	tok := seedToken(t, tokens, userA)

	cmd, _ := revoketoken.NewCommand(tok.ID().String(), userB.String(), testTime.Add(time.Minute))
	_, err := revoketoken.NewHandler(tokens).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, revoketoken.ErrUnauthorized)
}

func TestExecuteRejectsMissingToken(t *testing.T) {
	tokens := memiam.NewAPITokenRepository()
	cmd, _ := revoketoken.NewCommand(uuid.NewString(), userA.String(), testTime.Add(time.Minute))
	_, err := revoketoken.NewHandler(tokens).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrNotFound)
}

func TestExecuteRejectsDoubleRevoke(t *testing.T) {
	tokens := memiam.NewAPITokenRepository()
	tok := seedToken(t, tokens, userA)

	cmd, _ := revoketoken.NewCommand(tok.ID().String(), userA.String(), testTime.Add(time.Minute))
	_, err := revoketoken.NewHandler(tokens).Execute(context.Background(), cmd)
	require.NoError(t, err)

	_, err = revoketoken.NewHandler(tokens).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrInvalid)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	_, err := revoketoken.NewCommand("", userA.String(), testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, revoketoken.ErrInvalidCommand)

	_, err = revoketoken.NewCommand("at-1", "", testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, revoketoken.ErrInvalidCommand)

	_, err = revoketoken.NewCommand("at-1", userA.String(), time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, revoketoken.ErrInvalidCommand)
}
