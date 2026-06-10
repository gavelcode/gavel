package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/login"
	"github.com/usegavel/gavel/core/application/iam/resolveprincipal"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

var fixedNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func newAuthFixture(t *testing.T) (*auth.Middleware, string) {
	t.Helper()

	tenantRepo := memiam.NewTenantRepository()
	userRepo := memiam.NewUserRepository()
	sessionRepo := memiam.NewSessionRepository()
	tokenRepo := memiam.NewAPITokenRepository()
	hasher := memiam.NewFakeHasher()
	secretGen := memiam.NewFakeSecretGenerator()

	slug, err := tenant.NewSlug("test")
	require.NoError(t, err)
	testTenant, err := tenant.NewTenant(slug, "Test Tenant", fixedNow)
	require.NoError(t, err)
	require.NoError(t, tenantRepo.Save(context.Background(), testTenant))

	email, err := user.NewEmail("auth@example.com")
	require.NoError(t, err)
	role, err := user.NewRole("admin")
	require.NoError(t, err)
	hash, err := hasher.Hash("password123")
	require.NoError(t, err)
	testUser, err := user.NewUser(testTenant.ID(), email, "Auth User", role, hash, false, fixedNow)
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(context.Background(), testUser))

	loginHandler := login.NewHandler(tenantRepo, userRepo, sessionRepo, hasher, secretGen)
	cmd, err := login.NewCommand("test", "auth@example.com", "password123", "TestBot/1.0", "127.0.0.1", fixedNow, 24*time.Hour)
	require.NoError(t, err)
	loginResult, err := loginHandler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	resolveHandler := resolveprincipal.NewHandler(userRepo, sessionRepo, tokenRepo)
	cookie := auth.SessionCookie{Name: "session", TTL: 24 * time.Hour}
	middleware := auth.NewMiddleware(resolveHandler, cookie, func() time.Time { return fixedNow })

	return middleware, loginResult.SessionToken
}

func TestRequireRoleAllowsMatchingRole(t *testing.T) {
	principal := &auth.Principal{Role: "admin"}
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireRole("admin")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRoleRejectsWrongRole(t *testing.T) {
	principal := &auth.Principal{Role: "viewer"}
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireRole("admin")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireRoleRejectsNoPrincipal(t *testing.T) {
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireRole("admin")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireScopeOrRoleAllowsSessionWithRole(t *testing.T) {
	principal := &auth.Principal{Role: "admin", ViaAPIToken: false}
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireScopeOrRole("ingest", "admin")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireScopeOrRoleAllowsTokenWithScope(t *testing.T) {
	principal := &auth.Principal{ViaAPIToken: true, Scopes: []string{"ingest"}}
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireScopeOrRole("ingest", "admin")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireScopeOrRoleRejectsTokenMissingScope(t *testing.T) {
	principal := &auth.Principal{ViaAPIToken: true, Scopes: []string{"read"}}
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireScopeOrRole("ingest", "admin")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireScopeOrRoleRejectsSessionWrongRole(t *testing.T) {
	principal := &auth.Principal{Role: "viewer", ViaAPIToken: false}
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireScopeOrRole("ingest", "admin")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireScopeAllowsSessionUser(t *testing.T) {
	principal := &auth.Principal{Role: "admin", ViaAPIToken: false}
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireScope("ingest")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireScopeAllowsTokenWithScope(t *testing.T) {
	principal := &auth.Principal{ViaAPIToken: true, Scopes: []string{"ingest", "read"}}
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireScope("ingest")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireScopeRejectsTokenMissingScope(t *testing.T) {
	principal := &auth.Principal{ViaAPIToken: true, Scopes: []string{"read"}}
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireScope("ingest")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireScopeRejectsNoPrincipal(t *testing.T) {
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireScope("ingest")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireScopeOrRoleRejectsNoPrincipal(t *testing.T) {
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := (&auth.Middleware{}).RequireScopeOrRole("ingest", "admin")
	handler := middleware(inner)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticateWithValidCookieSetsPrincipal(t *testing.T) {
	middleware, sessionToken := newAuthFixture(t)
	var capturedPrincipal *auth.Principal
	inner := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedPrincipal, _ = auth.PrincipalFromContext(req.Context())
	})

	handler := middleware.Authenticate(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, capturedPrincipal)
	assert.Equal(t, "auth@example.com", capturedPrincipal.Email)
	assert.Equal(t, "admin", capturedPrincipal.Role)
	assert.False(t, capturedPrincipal.ViaAPIToken)
}

func TestAuthenticateWithNoCredentialsReturns401(t *testing.T) {
	mw, _ := newAuthFixture(t)
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	handler := mw.Authenticate(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticateWithInvalidCookieReturns401(t *testing.T) {
	mw, _ := newAuthFixture(t)
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	handler := mw.Authenticate(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticateWithInvalidBearerReturns401(t *testing.T) {
	mw, _ := newAuthFixture(t)
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	handler := mw.Authenticate(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestNewMiddlewarePanicsOnNilResolver(t *testing.T) {
	assert.Panics(t, func() {
		auth.NewMiddleware(nil, auth.SessionCookie{}, nil)
	})
}

func TestNewMiddlewareDefaultsClock(t *testing.T) {
	userRepo := memiam.NewUserRepository()
	sessionRepo := memiam.NewSessionRepository()
	tokenRepo := memiam.NewAPITokenRepository()
	resolveHandler := resolveprincipal.NewHandler(userRepo, sessionRepo, tokenRepo)

	mw := auth.NewMiddleware(resolveHandler, auth.SessionCookie{Name: "s"}, nil)

	assert.NotNil(t, mw)
}
