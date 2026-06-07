package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	tenantmodel "github.com/usegavel/gavel/core/domain/iam/model/tenant"
	usermodel "github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/infrastructure/iam/postgres"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
)

var testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

const (
	validHash    = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	validSecret  = "gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	validSession = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
)

func setupDB(t *testing.T) *database.DB { return testkit.TestDB(t) }

func seedTenant(t *testing.T, database *database.DB, slugRaw string) tenantmodel.Tenant {
	t.Helper()
	slug, err := tenantmodel.NewSlug(slugRaw)
	require.NoError(t, err)
	tenant, err := tenantmodel.NewTenant(slug, "Display "+slugRaw, testTime)
	require.NoError(t, err)
	tenant.ClearEvents()
	require.NoError(t, postgres.NewTenantRepo(database).Save(context.Background(), tenant))
	return tenant
}

func seedUser(t *testing.T, database *database.DB, tenantID tenantmodel.TenantID, emailRaw string) usermodel.User {
	t.Helper()
	email, err := usermodel.NewEmail(emailRaw)
	require.NoError(t, err)
	hash, err := usermodel.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	require.NoError(t, err)
	user, err := usermodel.NewUser(tenantID, email, "Alice", usermodel.RoleAdmin, hash, false, testTime)
	require.NoError(t, err)
	user.ClearEvents()
	require.NoError(t, postgres.NewUserRepo(database).Save(context.Background(), user))
	return user
}

func TestTenantRepoRoundTrip(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewTenantRepo(database)
	tenant := seedTenant(t, database, "acme")

	got, err := repo.ByID(context.Background(), tenant.ID())
	require.NoError(t, err)
	assert.True(t, tenant.ID().Equal(got.ID()))
	assert.Equal(t, tenant.Slug().String(), got.Slug().String())
	assert.Equal(t, tenant.DisplayName(), got.DisplayName())
	assert.True(t, got.Status().IsActive())
	assert.True(t, tenant.CreatedAt().Equal(got.CreatedAt()))

	bySlug, err := repo.BySlug(context.Background(), tenant.Slug())
	require.NoError(t, err)
	assert.True(t, tenant.ID().Equal(bySlug.ID()))

	missingID := tenantmodel.NewTenantID(uuid.New())
	_, err = repo.ByID(context.Background(), missingID)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenantmodel.ErrTenantNotFound)
}

func TestTenantRepoUpdateOnConflict(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewTenantRepo(database)
	tenant := seedTenant(t, database, "acme")

	require.NoError(t, tenant.Suspend(testTime.Add(time.Hour)))
	tenant.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), tenant))

	got, _ := repo.ByID(context.Background(), tenant.ID())
	assert.False(t, got.Status().IsActive())
}

func TestUserRepoRoundTrip(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewUserRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	got, err := repo.ByID(context.Background(), user.ID())
	require.NoError(t, err)
	assert.True(t, user.ID().Equal(got.ID()))
	assert.True(t, user.TenantID().Equal(got.TenantID()))
	assert.Equal(t, user.Email().String(), got.Email().String())
	assert.True(t, got.IsActive())

	byEmail, err := repo.ByEmail(context.Background(), tenant.ID(), user.Email())
	require.NoError(t, err)
	assert.True(t, user.ID().Equal(byEmail.ID()))

	count, err := repo.CountByTenant(context.Background(), tenant.ID())
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	missingID := usermodel.NewUserID(uuid.New())
	_, err = repo.ByID(context.Background(), missingID)
	require.Error(t, err)
	assert.ErrorIs(t, err, usermodel.ErrUserNotFound)
}

func TestUserRepoRejectsDuplicateEmailWithinTenant(t *testing.T) {
	database := setupDB(t)
	tenant := seedTenant(t, database, "acme")
	_ = seedUser(t, database, tenant.ID(), "alice@example.com")

	dup, err := usermodel.NewEmail("alice@example.com")
	require.NoError(t, err)
	hash, _ := usermodel.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	user, err := usermodel.NewUser(tenant.ID(), dup, "Other Alice", usermodel.RoleViewer, hash, false, testTime)
	require.NoError(t, err)
	user.ClearEvents()
	err = postgres.NewUserRepo(database).Save(context.Background(), user)
	require.Error(t, err)
	assert.ErrorIs(t, err, usermodel.ErrEmailAlreadyInUse)
}

func TestSessionRepoRoundTrip(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewSessionRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	tok, err := session.NewToken(validSession)
	require.NoError(t, err)
	sess, err := session.NewSession(tok, user.ID(), "Mozilla/5.0", "203.0.113.42", testTime, testTime.Add(24*time.Hour))
	require.NoError(t, err)
	sess.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess))

	got, err := repo.ByTokenHash(context.Background(), sess.TokenHash())
	require.NoError(t, err)
	assert.True(t, user.ID().Equal(got.UserID()))
	assert.Equal(t, "Mozilla/5.0", got.UserAgent())
	assert.False(t, got.IsRevoked())

	missing, _ := session.NewTokenHash(validHash)
	_, err = repo.ByTokenHash(context.Background(), missing)
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrNotFound)
}

func TestSessionRepoDeleteAllForUser(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewSessionRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	tok, _ := session.NewToken(validSession)
	sess, _ := session.NewSession(tok, user.ID(), "ua", "ip", testTime, testTime.Add(time.Hour))
	sess.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess))

	require.NoError(t, repo.DeleteAllForUser(context.Background(), user.ID()))
	_, err := repo.ByTokenHash(context.Background(), sess.TokenHash())
	require.Error(t, err)
}

func TestSessionRepoDeleteExpired(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewSessionRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	tok, _ := session.NewToken(validSession)
	sess, _ := session.NewSession(tok, user.ID(), "ua", "ip", testTime, testTime.Add(time.Hour))
	sess.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess))

	n, err := repo.DeleteExpired(context.Background(), testTime.Add(2*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestAPITokenRepoRoundTrip(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewAPITokenRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	secret, _ := apitoken.NewSecret(validSecret)
	scopes := apitoken.Scopes{apitoken.ScopeIngest, apitoken.ScopeRead}
	token, err := apitoken.NewAPIToken(secret, tenant.ID(), user.ID(), "ci-bot", scopes, testTime, testTime.Add(30*24*time.Hour))
	require.NoError(t, err)
	token.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), token))

	got, err := repo.ByID(context.Background(), token.ID())
	require.NoError(t, err)
	assert.True(t, token.ID().Equal(got.ID()))
	assert.True(t, tenant.ID().Equal(got.TenantID()))
	assert.True(t, user.ID().Equal(got.UserID()))
	assert.Equal(t, "ci-bot", got.Name())
	assert.True(t, token.TokenHash().Equal(got.TokenHash()))
	assert.Equal(t, secret.Prefix(), got.TokenPrefix())
	gotScopes := got.Scopes()
	assert.Len(t, gotScopes, 2)
	assert.True(t, gotScopes.Contains(apitoken.ScopeIngest))
	assert.True(t, gotScopes.Contains(apitoken.ScopeRead))

	byHash, err := repo.ByTokenHash(context.Background(), token.TokenHash())
	require.NoError(t, err)
	assert.True(t, token.ID().Equal(byHash.ID()))

	list, err := repo.ListByUser(context.Background(), user.ID())
	require.NoError(t, err)
	assert.Len(t, list, 1)

	missingID := apitoken.NewAPITokenID(uuid.New())
	_, err = repo.ByID(context.Background(), missingID)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrNotFound)
}

func TestAPITokenRepoZeroExpiresAtMeansNeverExpires(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewAPITokenRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	secret, _ := apitoken.NewSecret(validSecret)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}
	token, err := apitoken.NewAPIToken(secret, tenant.ID(), user.ID(), "perpetual", scopes, testTime, time.Time{})
	require.NoError(t, err)
	token.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), token))

	got, err := repo.ByID(context.Background(), token.ID())
	require.NoError(t, err)
	assert.True(t, got.ExpiresAt().IsZero(), "zero expiresAt must round-trip through NULL")
}

func TestTenantRepoBySlugNotFound(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewTenantRepo(database)

	slug, err := tenantmodel.NewSlug("nonexistent")
	require.NoError(t, err)

	_, err = repo.BySlug(context.Background(), slug)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenantmodel.ErrTenantNotFound)
}

func TestTenantRepoRejectsDuplicateSlug(t *testing.T) {
	database := setupDB(t)
	_ = seedTenant(t, database, "acme")

	slug, err := tenantmodel.NewSlug("acme")
	require.NoError(t, err)
	dup, err := tenantmodel.NewTenant(slug, "Another Acme", testTime)
	require.NoError(t, err)
	dup.ClearEvents()

	err = postgres.NewTenantRepo(database).Save(context.Background(), dup)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenantmodel.ErrSlugTaken)
}

func TestUserRepoByEmailNotFound(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewUserRepo(database)
	tenant := seedTenant(t, database, "acme")

	email, err := usermodel.NewEmail("nobody@example.com")
	require.NoError(t, err)

	_, err = repo.ByEmail(context.Background(), tenant.ID(), email)
	require.Error(t, err)
	assert.ErrorIs(t, err, usermodel.ErrUserNotFound)
}

func TestUserRepoUpdateLastLogin(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewUserRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	loginTime := testTime.Add(2 * time.Hour)
	require.NoError(t, user.TouchLogin(loginTime))
	user.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), user))

	got, err := repo.ByID(context.Background(), user.ID())
	require.NoError(t, err)
	assert.True(t, loginTime.Equal(got.LastLoginAt()))
}

func TestUserRepoCountByTenantMultipleUsers(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewUserRepo(database)
	tenant := seedTenant(t, database, "acme")
	other := seedTenant(t, database, "other")
	_ = seedUser(t, database, tenant.ID(), "alice@example.com")
	_ = seedUser(t, database, tenant.ID(), "bob@example.com")
	_ = seedUser(t, database, other.ID(), "carol@example.com")

	count, err := repo.CountByTenant(context.Background(), tenant.ID())
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	otherCount, err := repo.CountByTenant(context.Background(), other.ID())
	require.NoError(t, err)
	assert.Equal(t, 1, otherCount)
}

func TestAPITokenRepoByTokenHashNotFound(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewAPITokenRepo(database)

	hash, err := apitoken.NewSecretHash(validHash)
	require.NoError(t, err)

	_, err = repo.ByTokenHash(context.Background(), hash)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrNotFound)
}

func TestAPITokenRepoLastUsedAtRoundTrip(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewAPITokenRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	secret, _ := apitoken.NewSecret(validSecret)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}
	token, err := apitoken.NewAPIToken(secret, tenant.ID(), user.ID(), "ci-bot", scopes, testTime, testTime.Add(30*24*time.Hour))
	require.NoError(t, err)

	usedAt := testTime.Add(time.Hour)
	require.NoError(t, token.TouchUsed(usedAt))
	token.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), token))

	got, err := repo.ByID(context.Background(), token.ID())
	require.NoError(t, err)
	assert.True(t, usedAt.Equal(got.LastUsedAt()))
}

func TestAPITokenRepoListByUserEmpty(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewAPITokenRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	list, err := repo.ListByUser(context.Background(), user.ID())
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestAPITokenRepoListByUserOrdering(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewAPITokenRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	secret1, _ := apitoken.NewSecret("gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	scopes := apitoken.Scopes{apitoken.ScopeIngest}
	tok1, err := apitoken.NewAPIToken(secret1, tenant.ID(), user.ID(), "first", scopes, testTime, testTime.Add(30*24*time.Hour))
	require.NoError(t, err)
	tok1.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), tok1))

	secret2, _ := apitoken.NewSecret("gav_BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	tok2, err := apitoken.NewAPIToken(secret2, tenant.ID(), user.ID(), "second", scopes, testTime.Add(time.Hour), testTime.Add(31*24*time.Hour))
	require.NoError(t, err)
	tok2.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), tok2))

	list, err := repo.ListByUser(context.Background(), user.ID())
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "second", list[0].Name())
	assert.Equal(t, "first", list[1].Name())
}

func TestAPITokenRepoRevokedTokenRoundTrip(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewAPITokenRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	secret, _ := apitoken.NewSecret(validSecret)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}
	token, err := apitoken.NewAPIToken(secret, tenant.ID(), user.ID(), "ci-bot", scopes, testTime, testTime.Add(30*24*time.Hour))
	require.NoError(t, err)
	token.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), token))

	require.NoError(t, token.Revoke(testTime.Add(time.Hour)))
	token.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), token))

	got, err := repo.ByID(context.Background(), token.ID())
	require.NoError(t, err)
	assert.True(t, got.IsRevoked())
}

func TestSessionRepoRevokedSessionRoundTrip(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewSessionRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	tok, _ := session.NewToken(validSession)
	sess, _ := session.NewSession(tok, user.ID(), "Mozilla/5.0", "203.0.113.42", testTime, testTime.Add(24*time.Hour))
	sess.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess))

	require.NoError(t, sess.Revoke(testTime.Add(time.Hour)))
	sess.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess))

	got, err := repo.ByTokenHash(context.Background(), sess.TokenHash())
	require.NoError(t, err)
	assert.True(t, got.IsRevoked())
}

func TestSessionRepoDeleteExpiredIncludesRevoked(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewSessionRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")

	tok, _ := session.NewToken(validSession)
	sess, _ := session.NewSession(tok, user.ID(), "ua", "ip", testTime, testTime.Add(24*time.Hour))
	sess.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess))

	require.NoError(t, sess.Revoke(testTime.Add(time.Minute)))
	sess.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess))

	n, err := repo.DeleteExpired(context.Background(), testTime)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n, "revoked sessions should be cleaned up regardless of expiry")
}

func TestSessionRepoDeleteAllForUserMultipleSessions(t *testing.T) {
	database := setupDB(t)
	repo := postgres.NewSessionRepo(database)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")
	other := seedUser(t, database, tenant.ID(), "bob@example.com")

	tok1, _ := session.NewToken("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	sess1, _ := session.NewSession(tok1, user.ID(), "ua1", "ip1", testTime, testTime.Add(time.Hour))
	sess1.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess1))

	tok2, _ := session.NewToken("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	sess2, _ := session.NewSession(tok2, user.ID(), "ua2", "ip2", testTime, testTime.Add(time.Hour))
	sess2.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess2))

	tok3, _ := session.NewToken("CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC")
	sess3, _ := session.NewSession(tok3, other.ID(), "ua3", "ip3", testTime, testTime.Add(time.Hour))
	sess3.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), sess3))

	require.NoError(t, repo.DeleteAllForUser(context.Background(), user.ID()))

	_, err := repo.ByTokenHash(context.Background(), sess1.TokenHash())
	assert.ErrorIs(t, err, session.ErrNotFound)
	_, err = repo.ByTokenHash(context.Background(), sess2.TokenHash())
	assert.ErrorIs(t, err, session.ErrNotFound)

	got, err := repo.ByTokenHash(context.Background(), sess3.TokenHash())
	require.NoError(t, err)
	assert.True(t, other.ID().Equal(got.UserID()))
}

func seedAPIToken(t *testing.T, database *database.DB, tenantID tenantmodel.TenantID, userID usermodel.UserID, secretStr string) apitoken.APIToken {
	t.Helper()
	secret, err := apitoken.NewSecret(secretStr)
	require.NoError(t, err)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}
	token, err := apitoken.NewAPIToken(secret, tenantID, userID, "test-token", scopes, testTime, testTime.Add(30*24*time.Hour))
	require.NoError(t, err)
	token.ClearEvents()
	require.NoError(t, postgres.NewAPITokenRepo(database).Save(context.Background(), token))
	return token
}

func TestTenantSaveReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)
	slug, err := tenantmodel.NewSlug("cancel-save")
	require.NoError(t, err)
	tenant, err := tenantmodel.NewTenant(slug, "Cancel Save", testTime)
	require.NoError(t, err)
	tenant.ClearEvents()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = postgres.NewTenantRepo(database).Save(ctx, tenant)
	assert.Error(t, err)
}

func TestTenantByIDReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := postgres.NewTenantRepo(database).ByID(ctx, tenantmodel.NewTenantID(uuid.New()))
	assert.Error(t, err)
	assert.NotErrorIs(t, err, tenantmodel.ErrTenantNotFound)
}

func TestUserSaveReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)
	email, err := usermodel.NewEmail("cancel@example.com")
	require.NoError(t, err)
	hash, err := usermodel.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	require.NoError(t, err)
	user, err := usermodel.NewUser(tenantmodel.NewTenantID(uuid.New()), email, "Cancel", usermodel.RoleAdmin, hash, false, testTime)
	require.NoError(t, err)
	user.ClearEvents()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = postgres.NewUserRepo(database).Save(ctx, user)
	assert.Error(t, err)
}

func TestUserCountByTenantReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := postgres.NewUserRepo(database).CountByTenant(ctx, tenantmodel.NewTenantID(uuid.New()))
	assert.Error(t, err)
}

func TestUserByIDReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := postgres.NewUserRepo(database).ByID(ctx, usermodel.NewUserID(uuid.New()))
	assert.Error(t, err)
	assert.NotErrorIs(t, err, usermodel.ErrUserNotFound)
}

func TestSessionSaveReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)
	tok, err := session.NewToken(validSession)
	require.NoError(t, err)
	sess, err := session.NewSession(tok, usermodel.NewUserID(uuid.New()), "ua", "ip", testTime, testTime.Add(time.Hour))
	require.NoError(t, err)
	sess.ClearEvents()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = postgres.NewSessionRepo(database).Save(ctx, sess)
	assert.Error(t, err)
}

func TestSessionByTokenHashReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)
	hash, err := session.NewTokenHash(validHash)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = postgres.NewSessionRepo(database).ByTokenHash(ctx, hash)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, session.ErrNotFound)
}

func TestSessionDeleteAllForUserReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := postgres.NewSessionRepo(database).DeleteAllForUser(ctx, usermodel.NewUserID(uuid.New()))
	assert.Error(t, err)
}

func TestSessionDeleteExpiredReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := postgres.NewSessionRepo(database).DeleteExpired(ctx, testTime)
	assert.Error(t, err)
}

func TestAPITokenSaveReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)
	secret, err := apitoken.NewSecret(validSecret)
	require.NoError(t, err)
	scopes := apitoken.Scopes{apitoken.ScopeIngest}
	token, err := apitoken.NewAPIToken(secret, tenantmodel.NewTenantID(uuid.New()), usermodel.NewUserID(uuid.New()), "cancel", scopes, testTime, testTime.Add(time.Hour))
	require.NoError(t, err)
	token.ClearEvents()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = postgres.NewAPITokenRepo(database).Save(ctx, token)
	assert.Error(t, err)
}

func TestAPITokenListByUserReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := postgres.NewAPITokenRepo(database).ListByUser(ctx, usermodel.NewUserID(uuid.New()))
	assert.Error(t, err)
}

func TestAPITokenByIDReturnsErrorOnCancelledContext(t *testing.T) {
	database := setupDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := postgres.NewAPITokenRepo(database).ByID(ctx, apitoken.NewAPITokenID(uuid.New()))
	assert.Error(t, err)
	assert.NotErrorIs(t, err, apitoken.ErrNotFound)
}

func TestTenantByIDReturnsErrorOnCorruptedSlug(t *testing.T) {
	database := setupDB(t)
	tenant := seedTenant(t, database, "acme")
	ctx := context.Background()

	_, err := database.ExecContext(ctx,
		"UPDATE iam_tenants SET slug = '' WHERE id = ?", tenant.ID().UUID())
	require.NoError(t, err)

	_, err = postgres.NewTenantRepo(database).ByID(ctx, tenant.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slug")
}

func TestUserByIDReturnsErrorOnCorruptedEmail(t *testing.T) {
	database := setupDB(t)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")
	ctx := context.Background()

	_, err := database.ExecContext(ctx,
		"UPDATE iam_users SET email = '' WHERE id = ?", user.ID().UUID())
	require.NoError(t, err)

	_, err = postgres.NewUserRepo(database).ByID(ctx, user.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email")
}

func TestUserByIDReturnsErrorOnCorruptedPasswordHash(t *testing.T) {
	database := setupDB(t)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")
	ctx := context.Background()

	_, err := database.ExecContext(ctx,
		"UPDATE iam_users SET password_hash = 'x' WHERE id = ?", user.ID().UUID())
	require.NoError(t, err)

	_, err = postgres.NewUserRepo(database).ByID(ctx, user.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password hash")
}

func TestAPITokenListByUserReturnsErrorOnCorruptedTokenHash(t *testing.T) {
	database := setupDB(t)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")
	seedAPIToken(t, database, tenant.ID(), user.ID(), validSecret)
	ctx := context.Background()

	_, err := database.ExecContext(ctx,
		"UPDATE iam_api_tokens SET token_hash = 'x' WHERE user_id = ?", user.ID().UUID())
	require.NoError(t, err)

	_, err = postgres.NewAPITokenRepo(database).ListByUser(ctx, user.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token hash")
}

func TestAPITokenListByUserReturnsErrorOnCorruptedScopesJSON(t *testing.T) {
	database := setupDB(t)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")
	seedAPIToken(t, database, tenant.ID(), user.ID(), validSecret)
	ctx := context.Background()

	_, err := database.ExecContext(ctx,
		`UPDATE iam_api_tokens SET scopes = '"not-array"' WHERE user_id = ?`, user.ID().UUID())
	require.NoError(t, err)

	_, err = postgres.NewAPITokenRepo(database).ListByUser(ctx, user.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal scopes")
}

func TestAPITokenListByUserReturnsErrorOnInvalidScope(t *testing.T) {
	database := setupDB(t)
	tenant := seedTenant(t, database, "acme")
	user := seedUser(t, database, tenant.ID(), "alice@example.com")
	seedAPIToken(t, database, tenant.ID(), user.ID(), validSecret)
	ctx := context.Background()

	_, err := database.ExecContext(ctx,
		`UPDATE iam_api_tokens SET scopes = '["invalid"]' WHERE user_id = ?`, user.ID().UUID())
	require.NoError(t, err)

	_, err = postgres.NewAPITokenRepo(database).ListByUser(ctx, user.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scope")
}
