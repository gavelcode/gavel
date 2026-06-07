package resolveprincipal_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/resolveprincipal"
	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

const (
	validSession = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	validSecret  = "gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
)

type setup struct {
	users    *memiam.UserRepository
	sessions *memiam.SessionRepository
	tokens   *memiam.APITokenRepository
	user     user.User
	handler  *resolveprincipal.Handler
}

func newSetup(t *testing.T) *setup {
	t.Helper()
	users := memiam.NewUserRepository()
	sessions := memiam.NewSessionRepository()
	tokens := memiam.NewAPITokenRepository()

	tenantID := tenant.NewTenantID(uuid.New())
	email, _ := user.NewEmail("alice@example.com")
	hash, _ := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	seededUser, err := user.NewUser(tenantID, email, "Alice", user.RoleAdmin, hash, false, testTime)
	require.NoError(t, err)
	seededUser.ClearEvents()
	require.NoError(t, users.Save(context.Background(), seededUser))

	return &setup{
		users:    users,
		sessions: sessions,
		tokens:   tokens,
		user:     seededUser,
		handler:  resolveprincipal.NewHandler(users, sessions, tokens),
	}
}

func TestExecuteResolvesViaCookie(t *testing.T) {
	setup := newSetup(t)
	tok, _ := session.NewToken(validSession)
	sess, _ := session.NewSession(tok, setup.user.ID(), "ua", "ip", testTime, testTime.Add(time.Hour))
	sess.ClearEvents()
	require.NoError(t, setup.sessions.Save(context.Background(), sess))

	query, err := resolveprincipal.NewQuery(validSession, "", testTime.Add(time.Minute))
	require.NoError(t, err)
	p, err := setup.handler.Execute(context.Background(), query)
	require.NoError(t, err)
	assert.Equal(t, setup.user.ID().String(), p.UserID)
	assert.False(t, p.ViaAPIToken)
	assert.Empty(t, p.Scopes)
}

func TestExecuteResolvesViaBearer(t *testing.T) {
	setup := newSetup(t)
	secret, _ := apitoken.NewSecret(validSecret)
	tenantID := tenant.NewTenantID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeIngest, apitoken.ScopeRead}
	tok, err := apitoken.NewAPIToken(secret, tenantID, setup.user.ID(), "ci-bot", scopes, testTime, testTime.Add(time.Hour))
	require.NoError(t, err)
	tok.ClearEvents()
	require.NoError(t, setup.tokens.Save(context.Background(), tok))

	query, _ := resolveprincipal.NewQuery("", validSecret, testTime.Add(time.Minute))
	p, err := setup.handler.Execute(context.Background(), query)
	require.NoError(t, err)
	assert.Equal(t, setup.user.ID().String(), p.UserID)
	assert.True(t, p.ViaAPIToken)
	assert.Equal(t, tok.ID().String(), p.APITokenID)
	assert.ElementsMatch(t, []string{"ingest", "read"}, p.Scopes)
}

func TestExecuteRejectsNoCredentials(t *testing.T) {
	setup := newSetup(t)
	query, _ := resolveprincipal.NewQuery("", "", testTime.Add(time.Minute))
	_, err := setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)
}

func TestExecuteRejectsExpiredSession(t *testing.T) {
	setup := newSetup(t)
	tok, _ := session.NewToken(validSession)
	sess, _ := session.NewSession(tok, setup.user.ID(), "ua", "ip", testTime, testTime.Add(time.Hour))
	sess.ClearEvents()
	require.NoError(t, setup.sessions.Save(context.Background(), sess))

	query, _ := resolveprincipal.NewQuery(validSession, "", testTime.Add(2*time.Hour))
	_, err := setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)
}

func TestExecuteRejectsRevokedToken(t *testing.T) {
	setup := newSetup(t)
	secret, _ := apitoken.NewSecret(validSecret)
	tenantID := tenant.NewTenantID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeRead}
	tok, _ := apitoken.NewAPIToken(secret, tenantID, setup.user.ID(), "ci-bot", scopes, testTime, testTime.Add(time.Hour))
	require.NoError(t, tok.Revoke(testTime.Add(time.Minute)))
	tok.ClearEvents()
	require.NoError(t, setup.tokens.Save(context.Background(), tok))

	query, _ := resolveprincipal.NewQuery("", validSecret, testTime.Add(2*time.Minute))
	_, err := setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)
}

func TestExecuteRejectsInactiveUser(t *testing.T) {
	setup := newSetup(t)
	require.NoError(t, setup.user.Deactivate(testTime.Add(time.Minute)))
	setup.user.ClearEvents()
	require.NoError(t, setup.users.Save(context.Background(), setup.user))

	tok, _ := session.NewToken(validSession)
	sess, _ := session.NewSession(tok, setup.user.ID(), "ua", "ip", testTime, testTime.Add(time.Hour))
	sess.ClearEvents()
	require.NoError(t, setup.sessions.Save(context.Background(), sess))

	query, _ := resolveprincipal.NewQuery(validSession, "", testTime.Add(2*time.Minute))
	_, err := setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)
}

func TestExecuteRejectsMalformedCredentials(t *testing.T) {
	setup := newSetup(t)
	query, _ := resolveprincipal.NewQuery("not-a-token", "", testTime.Add(time.Minute))
	_, err := setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)

	query, _ = resolveprincipal.NewQuery("", "not-a-bearer", testTime.Add(time.Minute))
	_, err = setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)
}

func TestExecuteBearerTakesPrecedenceOverCookie(t *testing.T) {
	setup := newSetup(t)
	secret, _ := apitoken.NewSecret(validSecret)
	tenantID := tenant.NewTenantID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeRead}
	tok, _ := apitoken.NewAPIToken(secret, tenantID, setup.user.ID(), "ci-bot", scopes, testTime, testTime.Add(time.Hour))
	tok.ClearEvents()
	require.NoError(t, setup.tokens.Save(context.Background(), tok))

	query, _ := resolveprincipal.NewQuery(validSession, validSecret, testTime.Add(time.Minute))
	p, err := setup.handler.Execute(context.Background(), query)
	require.NoError(t, err)
	assert.True(t, p.ViaAPIToken, "when both credentials are presented, bearer wins")
}

func TestExecuteRejectsBearerTokenNotFound(t *testing.T) {
	setup := newSetup(t)
	unknownSecret := "gav_BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"
	query, _ := resolveprincipal.NewQuery("", unknownSecret, testTime.Add(time.Minute))
	_, err := setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)
}

func TestExecuteRejectsBearerTokenUserNotFound(t *testing.T) {
	setup := newSetup(t)
	secret, _ := apitoken.NewSecret(validSecret)
	tenantID := tenant.NewTenantID(uuid.New())
	orphanUserID := user.NewUserID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeRead}
	tok, err := apitoken.NewAPIToken(secret, tenantID, orphanUserID, "ci-bot", scopes, testTime, testTime.Add(time.Hour))
	require.NoError(t, err)
	tok.ClearEvents()
	require.NoError(t, setup.tokens.Save(context.Background(), tok))

	query, _ := resolveprincipal.NewQuery("", validSecret, testTime.Add(time.Minute))
	_, err = setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)
}

func TestExecuteRejectsCookieSessionNotFound(t *testing.T) {
	setup := newSetup(t)
	unknownCookie := "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"
	query, _ := resolveprincipal.NewQuery(unknownCookie, "", testTime.Add(time.Minute))
	_, err := setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)
}

func TestExecuteRejectsCookieSessionUserNotFound(t *testing.T) {
	setup := newSetup(t)
	tok, _ := session.NewToken(validSession)
	orphanUserID := user.NewUserID(uuid.New())
	sess, _ := session.NewSession(tok, orphanUserID, "ua", "ip", testTime, testTime.Add(time.Hour))
	sess.ClearEvents()
	require.NoError(t, setup.sessions.Save(context.Background(), sess))

	query, _ := resolveprincipal.NewQuery(validSession, "", testTime.Add(time.Minute))
	_, err := setup.handler.Execute(context.Background(), query)
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrUnauthenticated)
}

func TestNewQueryRejectsZeroTime(t *testing.T) {
	_, err := resolveprincipal.NewQuery("c", "b", time.Time{})
	require.Error(t, err)
	assert.ErrorIs(t, err, resolveprincipal.ErrInvalidQuery)
}
