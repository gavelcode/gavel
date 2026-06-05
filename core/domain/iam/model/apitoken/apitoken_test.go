package apitoken_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

func mustAPITokenSecret(t *testing.T) apitoken.Secret {
	t.Helper()
	s, err := apitoken.NewSecret(validSecret)
	require.NoError(t, err)
	return s
}

func TestNewAPIToken(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	scopes := apitoken.Scopes{apitoken.ScopeRead, apitoken.ScopeIngest}
	createdAt := iamTestTime
	expiresAt := iamTestTime.Add(30 * 24 * time.Hour)

	tok, err := apitoken.NewAPIToken(secret, tenantID, userID, "ci-bot", scopes, createdAt, expiresAt)
	require.NoError(t, err)

	assert.True(t, tenantID.Equal(tok.TenantID()))
	assert.True(t, userID.Equal(tok.UserID()))
	assert.Equal(t, "ci-bot", tok.Name())
	expectedHash := apitoken.HashSecret(secret)
	assert.True(t, expectedHash.Equal(tok.TokenHash()))
	assert.Equal(t, secret.Prefix(), tok.TokenPrefix())
	assert.True(t, apitoken.Scopes(scopes).Contains(apitoken.ScopeRead))
	assert.Equal(t, createdAt, tok.CreatedAt())
	assert.Equal(t, expiresAt, tok.ExpiresAt())
	assert.True(t, tok.LastUsedAt().IsZero())
	assert.False(t, tok.IsRevoked())
}

func TestNewAPITokenAcceptsZeroExpiresAtAsNeverExpires(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}

	tok, err := apitoken.NewAPIToken(secret, tenantID, userID, "perpetual", scopes, iamTestTime, time.Time{})
	require.NoError(t, err)
	assert.True(t, tok.ExpiresAt().IsZero())
	assert.False(t, tok.IsExpired(iamTestTime.AddDate(10, 0, 0)), "zero expiresAt means never expires")
}

func TestNewAPITokenRejectsInvalidInputs(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}
	createdAt := iamTestTime

	cases := []struct {
		name     string
		secret   apitoken.Secret
		tenantID tenant.TenantID
		userID   user.UserID
		toolName string
		scopes   apitoken.Scopes
		usedAt       time.Time
		exp      time.Time
	}{
		{name: "empty name", secret: secret, tenantID: tenantID, userID: userID, toolName: "  ", scopes: scopes, usedAt: createdAt, exp: createdAt.Add(time.Hour)},
		{name: "empty scopes", secret: secret, tenantID: tenantID, userID: userID, toolName: "n", scopes: nil, usedAt: createdAt, exp: createdAt.Add(time.Hour)},
		{name: "zero createdAt", secret: secret, tenantID: tenantID, userID: userID, toolName: "n", scopes: scopes, usedAt: time.Time{}, exp: createdAt.Add(time.Hour)},
		{name: "expiresAt before createdAt", secret: secret, tenantID: tenantID, userID: userID, toolName: "n", scopes: scopes, usedAt: createdAt, exp: createdAt.Add(-time.Hour)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := apitoken.NewAPIToken(tc.secret, tc.tenantID, tc.userID, tc.toolName, tc.scopes, tc.usedAt, tc.exp)
			require.Error(t, err)
			assert.ErrorIs(t, err, apitoken.ErrInvalid)
		})
	}
}

func TestNewAPITokenRecordsTokenIssuedEvent(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}

	tok, _ := apitoken.NewAPIToken(secret, tenantID, userID, "ci-bot", scopes, iamTestTime, iamTestTime.Add(time.Hour))

	events := tok.Events()
	require.Len(t, events, 1)
	issued, ok := events[0].(apitoken.Issued)
	require.True(t, ok)
	assert.True(t, issued.TokenID().Equal(tok.ID()))
	assert.True(t, issued.UserID().Equal(userID))
	assert.True(t, issued.TenantID().Equal(tenantID))
	assert.Equal(t, "ci-bot", issued.Name())
	assert.Equal(t, iamTestTime, issued.OccurredAt())
}

func TestReconstituteAPITokenRejectsInvalidInputs(t *testing.T) {
	tokenID := apitoken.NewAPITokenID(uuid.New())
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	hash, _ := apitoken.NewSecretHash(validAPITokenHash)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}

	cases := []struct {
		name        string
		tokenPrefix string
		toolName    string
		createdAt   time.Time
	}{
		{name: "empty tokenPrefix", tokenPrefix: "  ", toolName: "bot", createdAt: iamTestTime},
		{name: "empty name", tokenPrefix: "gav_AAAAAAAA", toolName: "  ", createdAt: iamTestTime},
		{name: "zero createdAt", tokenPrefix: "gav_AAAAAAAA", toolName: "bot", createdAt: time.Time{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := apitoken.ReconstituteAPIToken(tokenID, tenantID, userID, tc.toolName, hash, tc.tokenPrefix, scopes, tc.createdAt, iamTestTime.Add(time.Hour), time.Time{}, false)
			require.Error(t, err)
			assert.ErrorIs(t, err, apitoken.ErrInvalid)
		})
	}
}

func TestReconstituteAPIToken(t *testing.T) {
	tokenID := apitoken.NewAPITokenID(uuid.New())
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	hash, _ := apitoken.NewSecretHash(validAPITokenHash)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}

	tok, err := apitoken.ReconstituteAPIToken(tokenID, tenantID, userID, "ci-bot", hash, "gav_AAAAAAAA", scopes, iamTestTime, iamTestTime.Add(time.Hour), iamTestTime.Add(30*time.Minute), false)
	require.NoError(t, err)
	assert.True(t, tokenID.Equal(tok.ID()))
	assert.Empty(t, tok.Events())
}

func TestAPITokenIsExpired(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}
	createdAt := iamTestTime
	expiresAt := iamTestTime.Add(time.Hour)

	tok, _ := apitoken.NewAPIToken(secret, tenantID, userID, "ci-bot", scopes, createdAt, expiresAt)
	assert.False(t, tok.IsExpired(createdAt))
	assert.False(t, tok.IsExpired(expiresAt.Add(-time.Minute)))
	assert.True(t, tok.IsExpired(expiresAt))
}

func TestAPITokenTouchUsed(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}

	tok, _ := apitoken.NewAPIToken(secret, tenantID, userID, "ci-bot", scopes, iamTestTime, iamTestTime.Add(time.Hour))
	tok.ClearEvents()

	usedAt := iamTestTime.Add(30 * time.Minute)
	require.NoError(t, tok.TouchUsed(usedAt))
	assert.Equal(t, usedAt, tok.LastUsedAt())
	assert.Empty(t, tok.Events(), "TouchUsed must not record events")
}

func TestAPITokenRevoke(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}

	tok, _ := apitoken.NewAPIToken(secret, tenantID, userID, "ci-bot", scopes, iamTestTime, iamTestTime.Add(time.Hour))
	tok.ClearEvents()

	usedAt := iamTestTime.Add(10 * time.Minute)
	require.NoError(t, tok.Revoke(usedAt))
	assert.True(t, tok.IsRevoked())
	assert.True(t, tok.IsExpired(usedAt), "revoked token is expired regardless of expiresAt")

	events := tok.Events()
	require.Len(t, events, 1)
	revoked, ok := events[0].(apitoken.Revoked)
	require.True(t, ok)
	assert.True(t, revoked.TokenID().Equal(tok.ID()))
	assert.Equal(t, usedAt, revoked.OccurredAt())

	require.Error(t, tok.Revoke(usedAt.Add(time.Hour)), "double revoke must be rejected")
}

func TestAPITokenTouchUsedRejectsZeroTimestamp(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tok, _ := apitoken.NewAPIToken(secret, mustTenantID(t), mustUserID(t), "bot", apitoken.Scopes{apitoken.ScopeIngest}, iamTestTime, iamTestTime.Add(time.Hour))

	err := tok.TouchUsed(time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrInvalid)
}

func TestAPITokenRevokeRejectsZeroTimestamp(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tok, _ := apitoken.NewAPIToken(secret, mustTenantID(t), mustUserID(t), "bot", apitoken.Scopes{apitoken.ScopeIngest}, iamTestTime, iamTestTime.Add(time.Hour))

	err := tok.Revoke(time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrInvalid)
}

func TestAPITokenHasScope(t *testing.T) {
	secret := mustAPITokenSecret(t)
	tenantID := mustTenantID(t)
	userID := mustUserID(t)
	scopes := apitoken.Scopes{apitoken.ScopeIngest, apitoken.ScopeRead}

	tok, _ := apitoken.NewAPIToken(secret, tenantID, userID, "ci-bot", scopes, iamTestTime, iamTestTime.Add(time.Hour))
	assert.True(t, tok.HasScope(apitoken.ScopeIngest))
	assert.True(t, tok.HasScope(apitoken.ScopeRead))
	assert.False(t, tok.HasScope(apitoken.ScopeAdmin))
}
