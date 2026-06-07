package listmytokens_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/listmytokens"
	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

const validSecret = "gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

func TestExecuteReturnsUserTokens(t *testing.T) {
	tokens := memiam.NewAPITokenRepository()
	tenantID := tenant.NewTenantID(uuid.New())
	userID := user.NewUserID(uuid.New())
	secret, _ := apitoken.NewSecret(validSecret)

	tok, err := apitoken.NewAPIToken(secret, tenantID, userID, "ci-bot", apitoken.Scopes{apitoken.ScopeRead}, testTime, testTime.Add(time.Hour))
	require.NoError(t, err)
	tok.ClearEvents()
	require.NoError(t, tokens.Save(context.Background(), tok))

	q, err := listmytokens.NewQuery(userID.String(), testTime.Add(10*time.Minute))
	require.NoError(t, err)
	res, err := listmytokens.NewHandler(tokens).Execute(context.Background(), q)
	require.NoError(t, err)
	require.Len(t, res.Tokens, 1)
	s := res.Tokens[0]
	assert.Equal(t, tok.ID().String(), s.ID)
	assert.Equal(t, "ci-bot", s.Name)
	assert.Equal(t, []string{"read"}, s.Scopes)
	assert.False(t, s.IsRevoked)
	assert.False(t, s.IsExpired)
}

func TestNewQueryRejectsInvalidInputs(t *testing.T) {
	_, err := listmytokens.NewQuery("", testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, listmytokens.ErrInvalidQuery)

	_, err = listmytokens.NewQuery("u-1", time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, listmytokens.ErrInvalidQuery)
}
