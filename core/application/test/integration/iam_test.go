package appintegration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/changepassword"
	"github.com/usegavel/gavel/core/application/iam/createuser"
	"github.com/usegavel/gavel/core/application/iam/deactivateuser"
	"github.com/usegavel/gavel/core/application/iam/issuetoken"
	"github.com/usegavel/gavel/core/application/iam/listmytokens"
	"github.com/usegavel/gavel/core/application/iam/login"
	"github.com/usegavel/gavel/core/application/iam/resolveprincipal"
	"github.com/usegavel/gavel/core/application/iam/revoketoken"
	tenantcreate "github.com/usegavel/gavel/core/application/iam/tenant/create"
	tenantsuspend "github.com/usegavel/gavel/core/application/iam/tenant/suspend"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var testTime = time.Date(2026, time.June, 16, 12, 0, 0, 0, time.UTC)

type iamFixture struct {
	tenantCreate *tenantcreate.Handler
	createUser   *createuser.Handler
	login        *login.Handler
	issueToken   *issuetoken.Handler
	revokeToken  *revoketoken.Handler
	listTokens   *listmytokens.Handler
	changePw     *changepassword.Handler
	deactivate   *deactivateuser.Handler
	resolve      *resolveprincipal.Handler
	suspend      *tenantsuspend.Handler
}

func newIAMFixture(t *testing.T) iamFixture {
	t.Helper()

	tenants := memiam.NewTenantRepository()
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	tokens := memiam.NewAPITokenRepository()
	hasher := memiam.NewFakeHasher()
	secrets := memiam.NewFakeSecretGenerator()

	return iamFixture{
		tenantCreate: tenantcreate.NewHandler(tenants),
		createUser:   createuser.NewHandler(tenants, users, hasher),
		login:        login.NewHandler(tenants, users, sessions, hasher, secrets),
		issueToken:   issuetoken.NewHandler(users, tokens, secrets),
		revokeToken:  revoketoken.NewHandler(tokens),
		listTokens:   listmytokens.NewHandler(tokens),
		changePw:     changepassword.NewHandler(users, sessions, hasher),
		deactivate:   deactivateuser.NewHandler(users, sessions),
		resolve:      resolveprincipal.NewHandler(users, sessions, tokens),
		suspend:      tenantsuspend.NewHandler(tenants),
	}
}

func mustCreateTenant(t *testing.T, f iamFixture) string {
	t.Helper()
	cmd, err := tenantcreate.NewCommand("acme", "Acme Corp", testTime)
	require.NoError(t, err)
	result, err := f.tenantCreate.Execute(context.Background(), cmd)
	require.NoError(t, err)
	return result.TenantID
}

func mustCreateUser(t *testing.T, f iamFixture, tenantID, email, password string) string {
	t.Helper()
	cmd, err := createuser.NewCommand(tenantID, email, "Test User", "admin", password, false, testTime)
	require.NoError(t, err)
	result, err := f.createUser.Execute(context.Background(), cmd)
	require.NoError(t, err)
	return result.UserID
}

func mustLogin(t *testing.T, f iamFixture, tenantSlug, email, password string) login.Result {
	t.Helper()
	cmd, err := login.NewCommand(tenantSlug, email, password, "test-agent", "127.0.0.1", testTime, 24*time.Hour)
	require.NoError(t, err)
	result, err := f.login.Execute(context.Background(), cmd)
	require.NoError(t, err)
	return result
}

func TestIAMLifecycle_CreateUserAndLogin(t *testing.T) {
	fixture := newIAMFixture(t)
	ctx := context.Background()

	tenantID := mustCreateTenant(t, fixture)
	userID := mustCreateUser(t, fixture, tenantID, "alice@example.com", "password123")

	cmd, err := login.NewCommand("acme", "alice@example.com", "password123", "test-agent", "127.0.0.1", testTime, 24*time.Hour)
	require.NoError(t, err)

	result, err := fixture.login.Execute(ctx, cmd)
	require.NoError(t, err)

	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, "alice@example.com", result.Email)
	assert.NotEmpty(t, result.SessionToken)
	assert.Equal(t, testTime.Add(24*time.Hour), result.SessionExpiresAt)
}

func TestIAMLifecycle_TokenIssueAndRevoke(t *testing.T) {
	fixture := newIAMFixture(t)
	ctx := context.Background()

	tenantID := mustCreateTenant(t, fixture)
	userID := mustCreateUser(t, fixture, tenantID, "bob@example.com", "password123")

	issueCmd, err := issuetoken.NewCommand(userID, "ci-token", []string{"ingest"}, testTime, time.Time{})
	require.NoError(t, err)

	issueResult, err := fixture.issueToken.Execute(ctx, issueCmd)
	require.NoError(t, err)
	assert.NotEmpty(t, issueResult.TokenID)
	assert.NotEmpty(t, issueResult.PlainSecret)
	assert.Equal(t, userID, issueResult.UserID)
	assert.Equal(t, tenantID, issueResult.TenantID)

	listQuery, err := listmytokens.NewQuery(userID, testTime)
	require.NoError(t, err)

	listResult, err := fixture.listTokens.Execute(ctx, listQuery)
	require.NoError(t, err)
	require.Len(t, listResult.Tokens, 1)
	assert.Equal(t, issueResult.TokenID, listResult.Tokens[0].ID)
	assert.False(t, listResult.Tokens[0].IsRevoked)

	revokeCmd, err := revoketoken.NewCommand(issueResult.TokenID, userID, testTime)
	require.NoError(t, err)

	_, err = fixture.revokeToken.Execute(ctx, revokeCmd)
	require.NoError(t, err)

	listResult, err = fixture.listTokens.Execute(ctx, listQuery)
	require.NoError(t, err)
	require.Len(t, listResult.Tokens, 1)
	assert.True(t, listResult.Tokens[0].IsRevoked)
}

func TestIAMLifecycle_ChangePassword(t *testing.T) {
	fixture := newIAMFixture(t)
	ctx := context.Background()

	tenantID := mustCreateTenant(t, fixture)
	userID := mustCreateUser(t, fixture, tenantID, "carol@example.com", "oldpass123")

	_ = mustLogin(t, fixture, "acme", "carol@example.com", "oldpass123")

	changePwCmd, err := changepassword.NewCommand(userID, "oldpass123", "newpass456", testTime)
	require.NoError(t, err)

	_, err = fixture.changePw.Execute(ctx, changePwCmd)
	require.NoError(t, err)

	newLoginCmd, err := login.NewCommand("acme", "carol@example.com", "newpass456", "test-agent", "127.0.0.1", testTime, 24*time.Hour)
	require.NoError(t, err)

	result, err := fixture.login.Execute(ctx, newLoginCmd)
	require.NoError(t, err)
	assert.Equal(t, userID, result.UserID)

	oldLoginCmd, err := login.NewCommand("acme", "carol@example.com", "oldpass123", "test-agent", "127.0.0.1", testTime, 24*time.Hour)
	require.NoError(t, err)

	_, err = fixture.login.Execute(ctx, oldLoginCmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestIAMLifecycle_DeactivateUser(t *testing.T) {
	fixture := newIAMFixture(t)
	ctx := context.Background()

	tenantID := mustCreateTenant(t, fixture)
	userID := mustCreateUser(t, fixture, tenantID, "dave@example.com", "password123")

	_ = mustLogin(t, fixture, "acme", "dave@example.com", "password123")

	deactivateCmd, err := deactivateuser.NewCommand(userID, testTime)
	require.NoError(t, err)

	_, err = fixture.deactivate.Execute(ctx, deactivateCmd)
	require.NoError(t, err)

	loginCmd, err := login.NewCommand("acme", "dave@example.com", "password123", "test-agent", "127.0.0.1", testTime, 24*time.Hour)
	require.NoError(t, err)

	_, err = fixture.login.Execute(ctx, loginCmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestIAMLifecycle_TenantSuspend(t *testing.T) {
	fixture := newIAMFixture(t)
	ctx := context.Background()

	tenantID := mustCreateTenant(t, fixture)
	_ = mustCreateUser(t, fixture, tenantID, "eve@example.com", "password123")

	_ = mustLogin(t, fixture, "acme", "eve@example.com", "password123")

	suspendCmd, err := tenantsuspend.NewCommand(tenantID, testTime)
	require.NoError(t, err)

	_, err = fixture.suspend.Execute(ctx, suspendCmd)
	require.NoError(t, err)

	loginCmd, err := login.NewCommand("acme", "eve@example.com", "password123", "test-agent", "127.0.0.1", testTime, 24*time.Hour)
	require.NoError(t, err)

	_, err = fixture.login.Execute(ctx, loginCmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestIAMLifecycle_ResolvePrincipal_Session(t *testing.T) {
	fixture := newIAMFixture(t)
	ctx := context.Background()

	tenantID := mustCreateTenant(t, fixture)
	userID := mustCreateUser(t, fixture, tenantID, "frank@example.com", "password123")

	loginResult := mustLogin(t, fixture, "acme", "frank@example.com", "password123")

	query, err := resolveprincipal.NewQuery(loginResult.SessionToken, "", testTime)
	require.NoError(t, err)

	principal, err := fixture.resolve.Execute(ctx, query)
	require.NoError(t, err)

	assert.Equal(t, userID, principal.UserID)
	assert.Equal(t, tenantID, principal.TenantID)
	assert.Equal(t, "frank@example.com", principal.Email)
	assert.Equal(t, "admin", principal.Role)
	assert.False(t, principal.ViaAPIToken)
	assert.Empty(t, principal.APITokenID)
}

func TestIAMLifecycle_ResolvePrincipal_APIToken(t *testing.T) {
	fixture := newIAMFixture(t)
	ctx := context.Background()

	tenantID := mustCreateTenant(t, fixture)
	userID := mustCreateUser(t, fixture, tenantID, "grace@example.com", "password123")

	issueCmd, err := issuetoken.NewCommand(userID, "deploy-token", []string{"ingest"}, testTime, time.Time{})
	require.NoError(t, err)

	issueResult, err := fixture.issueToken.Execute(ctx, issueCmd)
	require.NoError(t, err)

	query, err := resolveprincipal.NewQuery("", issueResult.PlainSecret, testTime)
	require.NoError(t, err)

	principal, err := fixture.resolve.Execute(ctx, query)
	require.NoError(t, err)

	assert.Equal(t, userID, principal.UserID)
	assert.Equal(t, tenantID, principal.TenantID)
	assert.Equal(t, "grace@example.com", principal.Email)
	assert.Equal(t, "admin", principal.Role)
	assert.True(t, principal.ViaAPIToken)
	assert.Equal(t, issueResult.TokenID, principal.APITokenID)
	assert.Contains(t, principal.Scopes, "ingest")
}
