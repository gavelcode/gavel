package iam_test

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
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

const (
	validHash     = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	validSecret   = "gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	validSession  = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	validSession2 = "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBA"
)

func newTenantForTest(t *testing.T, slugRaw string) tenantmodel.Tenant {
	t.Helper()
	slug, err := tenantmodel.NewSlug(slugRaw)
	require.NoError(t, err)
	tenant, err := tenantmodel.NewTenant(slug, "Display "+slugRaw, testTime)
	require.NoError(t, err)
	return tenant
}

func newUserForTest(t *testing.T, tenantID tenantmodel.TenantID, emailRaw string) usermodel.User {
	t.Helper()
	email, err := usermodel.NewEmail(emailRaw)
	require.NoError(t, err)
	hash, err := usermodel.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	require.NoError(t, err)
	u, err := usermodel.NewUser(tenantID, email, "Alice", usermodel.RoleAdmin, hash, false, testTime)
	require.NoError(t, err)
	return u
}

func TestTenantRepository(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewTenantRepository()

	tenant := newTenantForTest(t, "acme")
	require.NoError(t, repo.Save(ctx, tenant))

	got, err := repo.ByID(ctx, tenant.ID())
	require.NoError(t, err)
	assert.True(t, tenant.ID().Equal(got.ID()))

	gotBySlug, err := repo.BySlug(ctx, tenant.Slug())
	require.NoError(t, err)
	assert.True(t, tenant.ID().Equal(gotBySlug.ID()))

	_, err = repo.ByID(ctx, tenantmodel.TenantID{})
	require.Error(t, err)

	missing := tenantmodel.NewTenantID(uuid.New())
	_, err = repo.ByID(ctx, missing)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenantmodel.ErrTenantNotFound)
}

func TestTenantRepositoryRejectsDuplicateSlug(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewTenantRepository()

	first := newTenantForTest(t, "acme")
	require.NoError(t, repo.Save(ctx, first))

	second := newTenantForTest(t, "acme")
	err := repo.Save(ctx, second)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenantmodel.ErrSlugTaken)
}

func TestTenantRepositorySaveAllowsUpdate(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewTenantRepository()

	tenant := newTenantForTest(t, "acme")
	require.NoError(t, repo.Save(ctx, tenant))
	require.NoError(t, tenant.Suspend(testTime.Add(time.Hour)))
	require.NoError(t, repo.Save(ctx, tenant), "save must accept updates to the same tenant id")
}

func TestUserRepository(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewUserRepository()

	tenantID := tenantmodel.NewTenantID(uuid.New())
	user := newUserForTest(t, tenantID, "alice@example.com")
	require.NoError(t, repo.Save(ctx, user))

	got, err := repo.ByID(ctx, user.ID())
	require.NoError(t, err)
	assert.True(t, user.ID().Equal(got.ID()))

	gotByEmail, err := repo.ByEmail(ctx, tenantID, user.Email())
	require.NoError(t, err)
	assert.True(t, user.ID().Equal(gotByEmail.ID()))

	missing := usermodel.NewUserID(uuid.New())
	_, err = repo.ByID(ctx, missing)
	require.Error(t, err)
	assert.ErrorIs(t, err, usermodel.ErrUserNotFound)
}

func TestUserRepositoryRejectsDuplicateEmailWithinTenant(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewUserRepository()

	tenantID := tenantmodel.NewTenantID(uuid.New())
	first := newUserForTest(t, tenantID, "alice@example.com")
	require.NoError(t, repo.Save(ctx, first))

	second := newUserForTest(t, tenantID, "alice@example.com")
	err := repo.Save(ctx, second)
	require.Error(t, err)
	assert.ErrorIs(t, err, usermodel.ErrEmailAlreadyInUse)
}

func TestUserRepositoryAllowsSameEmailAcrossTenants(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewUserRepository()

	tenantA := tenantmodel.NewTenantID(uuid.New())
	tenantB := tenantmodel.NewTenantID(uuid.New())
	require.NoError(t, repo.Save(ctx, newUserForTest(t, tenantA, "alice@example.com")))
	require.NoError(t, repo.Save(ctx, newUserForTest(t, tenantB, "alice@example.com")))
}

func TestUserRepositoryCountByTenant(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewUserRepository()

	tenantID := tenantmodel.NewTenantID(uuid.New())
	require.NoError(t, repo.Save(ctx, newUserForTest(t, tenantID, "a@example.com")))
	require.NoError(t, repo.Save(ctx, newUserForTest(t, tenantID, "b@example.com")))

	otherTenant := tenantmodel.NewTenantID(uuid.New())
	require.NoError(t, repo.Save(ctx, newUserForTest(t, otherTenant, "a@example.com")))

	count, err := repo.CountByTenant(ctx, tenantID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	count, err = repo.CountByTenant(ctx, otherTenant)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSessionRepository(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewSessionRepository()

	tok, _ := session.NewToken(validSession)
	uid := usermodel.NewUserID(uuid.New())
	sess, err := session.NewSession(tok, uid, "ua", "ip", testTime, testTime.Add(24*time.Hour))
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, sess))

	got, err := repo.ByTokenHash(ctx, sess.TokenHash())
	require.NoError(t, err)
	assert.True(t, sess.UserID().Equal(got.UserID()))

	missing, _ := session.NewTokenHash(validHash)
	_, err = repo.ByTokenHash(ctx, missing)
	require.Error(t, err)
	assert.ErrorIs(t, err, session.ErrNotFound)
}

func TestSessionRepositoryDeleteAllForUser(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewSessionRepository()

	uid := usermodel.NewUserID(uuid.New())
	tok1, _ := session.NewToken(validSession)
	session1, _ := session.NewSession(tok1, uid, "ua", "ip", testTime, testTime.Add(24*time.Hour))
	require.NoError(t, repo.Save(ctx, session1))

	tok2, _ := session.NewToken(validSession2)
	otherUser := usermodel.NewUserID(uuid.New())
	session2, _ := session.NewSession(tok2, otherUser, "ua", "ip", testTime, testTime.Add(24*time.Hour))
	require.NoError(t, repo.Save(ctx, session2))

	require.NoError(t, repo.DeleteAllForUser(ctx, uid))
	_, err := repo.ByTokenHash(ctx, session1.TokenHash())
	require.Error(t, err, "uid's session must be gone")

	_, err = repo.ByTokenHash(ctx, session2.TokenHash())
	require.NoError(t, err, "other user's session must remain untouched")
}

func TestSessionRepositoryDeleteExpired(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewSessionRepository()

	tok, _ := session.NewToken(validSession)
	uid := usermodel.NewUserID(uuid.New())
	sess, _ := session.NewSession(tok, uid, "ua", "ip", testTime, testTime.Add(time.Hour))
	require.NoError(t, repo.Save(ctx, sess))

	removed, err := repo.DeleteExpired(ctx, testTime.Add(2*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(1), removed)

	_, err = repo.ByTokenHash(ctx, sess.TokenHash())
	require.Error(t, err)
}

func TestAPITokenRepository(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewAPITokenRepository()

	secret, _ := apitoken.NewSecret(validSecret)
	tenantID := tenantmodel.NewTenantID(uuid.New())
	uid := usermodel.NewUserID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeRead}

	tok, err := apitoken.NewAPIToken(secret, tenantID, uid, "ci-bot", scopes, testTime, testTime.Add(time.Hour))
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, tok))

	got, err := repo.ByID(ctx, tok.ID())
	require.NoError(t, err)
	assert.True(t, tok.ID().Equal(got.ID()))

	gotByHash, err := repo.ByTokenHash(ctx, tok.TokenHash())
	require.NoError(t, err)
	assert.True(t, tok.ID().Equal(gotByHash.ID()))

	missingID := apitoken.NewAPITokenID(uuid.New())
	_, err = repo.ByID(ctx, missingID)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrNotFound)

	missingHash, _ := apitoken.NewSecretHash(validHash)
	_, err = repo.ByTokenHash(ctx, missingHash)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrNotFound)
}

func TestAPITokenRepositoryListByUser(t *testing.T) {
	ctx := context.Background()
	repo := memiam.NewAPITokenRepository()

	secret, _ := apitoken.NewSecret(validSecret)
	tenantID := tenantmodel.NewTenantID(uuid.New())
	uid := usermodel.NewUserID(uuid.New())
	scopes := apitoken.Scopes{apitoken.ScopeRead}

	tok1, _ := apitoken.NewAPIToken(secret, tenantID, uid, "ci-bot", scopes, testTime, testTime.Add(time.Hour))
	tok2, _ := apitoken.NewAPIToken(secret, tenantID, uid, "release-bot", scopes, testTime.Add(time.Minute), testTime.Add(time.Hour))
	require.NoError(t, repo.Save(ctx, tok1))
	require.NoError(t, repo.Save(ctx, tok2))

	otherUser := usermodel.NewUserID(uuid.New())
	tok3, _ := apitoken.NewAPIToken(secret, tenantID, otherUser, "personal", scopes, testTime, testTime.Add(time.Hour))
	require.NoError(t, repo.Save(ctx, tok3))

	tokens, err := repo.ListByUser(ctx, uid)
	require.NoError(t, err)
	assert.Len(t, tokens, 2)
}
