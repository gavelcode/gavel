package iam_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/changepassword"
	"github.com/usegavel/gavel/core/application/iam/createuser"
	"github.com/usegavel/gavel/core/application/iam/issuetoken"
	"github.com/usegavel/gavel/core/application/iam/listmytokens"
	"github.com/usegavel/gavel/core/application/iam/login"
	"github.com/usegavel/gavel/core/application/iam/logout"
	"github.com/usegavel/gavel/core/application/iam/revoketoken"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/iam"
)

var fixedNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

const testTenantSlug = "test"

func nowFunc() time.Time { return fixedNow }

type testFixture struct {
	handler  *iam.Handler
	userID   string
	tenantID string
}

func newTestFixture(t *testing.T) testFixture {
	t.Helper()

	tenantRepo := memiam.NewTenantRepository()
	userRepo := memiam.NewUserRepository()
	sessionRepo := memiam.NewSessionRepository()
	tokenRepo := memiam.NewAPITokenRepository()
	hasher := memiam.NewFakeHasher()
	secretGen := memiam.NewFakeSecretGenerator()

	slug, err := tenant.NewSlug(testTenantSlug)
	require.NoError(t, err)
	testTenant, err := tenant.NewTenant(slug, "Test Tenant", fixedNow)
	require.NoError(t, err)
	require.NoError(t, tenantRepo.Save(context.Background(), testTenant))

	email, err := user.NewEmail("test@example.com")
	require.NoError(t, err)
	role, err := user.NewRole("admin")
	require.NoError(t, err)
	hash, err := hasher.Hash("password123")
	require.NoError(t, err)
	testUser, err := user.NewUser(testTenant.ID(), email, "Test User", role, hash, false, fixedNow)
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(context.Background(), testUser))

	handler := iam.New(iam.Deps{
		Login:          login.NewHandler(tenantRepo, userRepo, sessionRepo, hasher, secretGen),
		Logout:         logout.NewHandler(sessionRepo),
		ChangePassword: changepassword.NewHandler(userRepo, sessionRepo, hasher),
		IssueToken:     issuetoken.NewHandler(userRepo, tokenRepo, secretGen),
		RevokeToken:    revoketoken.NewHandler(tokenRepo),
		ListMyTokens:   listmytokens.NewHandler(tokenRepo),
		CreateUser:     createuser.NewHandler(tenantRepo, userRepo, hasher),
		Cookie:         auth.SessionCookie{Name: "session", TTL: 24 * time.Hour},
		DefaultTenant:  testTenantSlug,
		Now:            nowFunc,
	})

	return testFixture{
		handler:  handler,
		userID:   testUser.ID().String(),
		tenantID: testTenant.ID().String(),
	}
}

func withPrincipal(ctx context.Context, userID, tenantID, role string) context.Context {
	return auth.WithPrincipal(ctx, &auth.Principal{
		UserID:   userID,
		TenantID: tenantID,
		Email:    "test@example.com",
		Role:     role,
	})
}

func TestCreateSession_NilBodyReturns400(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.CreateSession(context.Background(), gen.CreateSessionRequestObject{Body: nil})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateSession400JSONResponse)
	assert.True(t, ok, "expected 400, got %T", resp)
}

func TestCreateSession_InvalidCredentials(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.CreateSession(context.Background(), gen.CreateSessionRequestObject{
		Body: &gen.LoginRequest{
			Email:    openapi_types.Email("wrong@example.com"),
			Password: "wrong",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateSession401JSONResponse)
	assert.True(t, ok, "expected 401, got %T", resp)
}

func TestCreateSession_ValidCredentialsReturnsSuccess(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.CreateSession(context.Background(), gen.CreateSessionRequestObject{
		Body: &gen.LoginRequest{
			Email:    openapi_types.Email("test@example.com"),
			Password: "password123",
		},
	})

	require.NoError(t, err)
	_, is400 := resp.(gen.CreateSession400JSONResponse)
	_, is401 := resp.(gen.CreateSession401JSONResponse)
	assert.False(t, is400, "unexpected 400")
	assert.False(t, is401, "unexpected 401")
}

func TestGetMe_NoPrincipalReturns401(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.GetMe(context.Background(), gen.GetMeRequestObject{})

	require.NoError(t, err)
	_, ok := resp.(gen.GetMe401JSONResponse)
	assert.True(t, ok, "expected 401, got %T", resp)
}

func TestGetMe_WithPrincipalReturnsMe(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.GetMe(ctx, gen.GetMeRequestObject{})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetMe200JSONResponse)
	require.True(t, ok, "expected 200, got %T", resp)
	assert.Equal(t, "admin", string(jsonResp.Role))
	assert.Equal(t, openapi_types.Email("test@example.com"), jsonResp.Email)
}

func TestDeleteCurrentSession_Returns204(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.DeleteCurrentSession(context.Background(), gen.DeleteCurrentSessionRequestObject{})

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestChangeMyPassword_NoPrincipalReturns401(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.ChangeMyPassword(context.Background(), gen.ChangeMyPasswordRequestObject{
		Body: &gen.ChangeMyPasswordJSONRequestBody{CurrentPassword: "old", NewPassword: "new"},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.ChangeMyPassword401JSONResponse)
	assert.True(t, ok, "expected 401, got %T", resp)
}

func TestChangeMyPassword_NilBodyReturns400(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.ChangeMyPassword(ctx, gen.ChangeMyPasswordRequestObject{Body: nil})

	require.NoError(t, err)
	_, ok := resp.(gen.ChangeMyPassword400JSONResponse)
	assert.True(t, ok, "expected 400, got %T", resp)
}

func TestListMyTokens_NoPrincipalReturns401(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.ListMyTokens(context.Background(), gen.ListMyTokensRequestObject{})

	require.NoError(t, err)
	_, ok := resp.(gen.ListMyTokens401JSONResponse)
	assert.True(t, ok, "expected 401, got %T", resp)
}

func TestListMyTokens_WithPrincipalReturnsEmpty(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.ListMyTokens(ctx, gen.ListMyTokensRequestObject{})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListMyTokens200JSONResponse)
	require.True(t, ok, "expected 200, got %T", resp)
	assert.Empty(t, jsonResp.Items)
}

func TestCreateMyToken_NoPrincipalReturns401(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.CreateMyToken(context.Background(), gen.CreateMyTokenRequestObject{
		Body: &gen.CreateTokenRequest{Name: "test", Scopes: []gen.CreateTokenRequestScopes{"read"}},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateMyToken401JSONResponse)
	assert.True(t, ok, "expected 401, got %T", resp)
}

func TestCreateMyToken_NilBodyReturns400(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.CreateMyToken(ctx, gen.CreateMyTokenRequestObject{Body: nil})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateMyToken400JSONResponse)
	assert.True(t, ok, "expected 400, got %T", resp)
}

func TestCreateMyToken_AdminScopeByNonAdminReturns403(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "viewer")

	resp, err := fix.handler.CreateMyToken(ctx, gen.CreateMyTokenRequestObject{
		Body: &gen.CreateTokenRequest{Name: "test", Scopes: []gen.CreateTokenRequestScopes{"admin"}},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateMyToken403JSONResponse)
	assert.True(t, ok, "expected 403, got %T", resp)
}

func TestCreateMyToken_SuccessReturns201(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.CreateMyToken(ctx, gen.CreateMyTokenRequestObject{
		Body: &gen.CreateTokenRequest{Name: "my-token", Scopes: []gen.CreateTokenRequestScopes{"read", "ingest"}},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.CreateMyToken201JSONResponse)
	require.True(t, ok, "expected 201, got %T", resp)
	assert.Equal(t, "my-token", jsonResp.Name)
	assert.NotEmpty(t, jsonResp.Token)
	assert.NotEmpty(t, jsonResp.Prefix)
	assert.ElementsMatch(t, []string{"read", "ingest"}, jsonResp.Scopes)
}

func TestCreateMyToken_WithExpiryReturns201(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")
	days := 30

	resp, err := fix.handler.CreateMyToken(ctx, gen.CreateMyTokenRequestObject{
		Body: &gen.CreateTokenRequest{
			Name:          "expiring",
			Scopes:        []gen.CreateTokenRequestScopes{"read"},
			ExpiresInDays: &days,
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateMyToken201JSONResponse)
	assert.True(t, ok, "expected 201, got %T", resp)
}

func TestDeleteMyToken_NoPrincipalReturns401(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.DeleteMyToken(context.Background(), gen.DeleteMyTokenRequestObject{})

	require.NoError(t, err)
	_, ok := resp.(gen.DeleteMyToken401JSONResponse)
	assert.True(t, ok, "expected 401, got %T", resp)
}

func TestDeleteMyToken_NotFoundReturns404(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.DeleteMyToken(ctx, gen.DeleteMyTokenRequestObject{
		Id: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
	})

	require.NoError(t, err)
	_, ok := resp.(gen.DeleteMyToken404JSONResponse)
	assert.True(t, ok, "expected 404, got %T", resp)
}

func TestDeleteMyToken_SuccessReturns204(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	createResp, err := fix.handler.CreateMyToken(ctx, gen.CreateMyTokenRequestObject{
		Body: &gen.CreateTokenRequest{Name: "to-delete", Scopes: []gen.CreateTokenRequestScopes{"read"}},
	})
	require.NoError(t, err)
	created, found := createResp.(gen.CreateMyToken201JSONResponse)
	require.True(t, found)

	resp, err := fix.handler.DeleteMyToken(ctx, gen.DeleteMyTokenRequestObject{Id: created.Id})

	require.NoError(t, err)
	_, found = resp.(gen.DeleteMyToken204Response)
	assert.True(t, found, "expected 204, got %T", resp)
}

func TestCreateUser_NoPrincipalReturns401(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.CreateUser(context.Background(), gen.CreateUserRequestObject{
		Body: &gen.CreateUserRequest{Email: openapi_types.Email("new@example.com"), DisplayName: "New", Password: "pass123"},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateUser401JSONResponse)
	assert.True(t, ok, "expected 401, got %T", resp)
}

func TestCreateUser_NilBodyReturns400(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.CreateUser(ctx, gen.CreateUserRequestObject{Body: nil})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateUser400JSONResponse)
	assert.True(t, ok, "expected 400, got %T", resp)
}

func TestCreateUser_SuccessReturns201(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.CreateUser(ctx, gen.CreateUserRequestObject{
		Body: &gen.CreateUserRequest{
			Email:       openapi_types.Email("newuser@example.com"),
			DisplayName: "New User",
			Password:    "securepass123",
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.CreateUser201JSONResponse)
	require.True(t, ok, "expected 201, got %T", resp)
	assert.Equal(t, openapi_types.Email("newuser@example.com"), jsonResp.Email)
}

func TestCreateUser_DuplicateEmailReturns409(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.CreateUser(ctx, gen.CreateUserRequestObject{
		Body: &gen.CreateUserRequest{
			Email:       openapi_types.Email("test@example.com"),
			DisplayName: "Duplicate",
			Password:    "securepass123",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateUser409JSONResponse)
	assert.True(t, ok, "expected 409, got %T", resp)
}

func TestCreateSession_WithCustomTenant(t *testing.T) {
	fix := newTestFixture(t)

	customTenant := testTenantSlug
	resp, err := fix.handler.CreateSession(context.Background(), gen.CreateSessionRequestObject{
		Body: &gen.LoginRequest{
			Email:    openapi_types.Email("test@example.com"),
			Password: "password123",
			Tenant:   &customTenant,
		},
	})

	require.NoError(t, err)
	_, is400 := resp.(gen.CreateSession400JSONResponse)
	_, is401 := resp.(gen.CreateSession401JSONResponse)
	assert.False(t, is400, "unexpected 400")
	assert.False(t, is401, "unexpected 401")
}

func TestCreateSession_UnknownTenantReturns401(t *testing.T) {
	fix := newTestFixture(t)

	unknownTenant := "nonexistent"
	resp, err := fix.handler.CreateSession(context.Background(), gen.CreateSessionRequestObject{
		Body: &gen.LoginRequest{
			Email:    openapi_types.Email("test@example.com"),
			Password: "password123",
			Tenant:   &unknownTenant,
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateSession401JSONResponse)
	assert.True(t, ok, "expected 401 for unknown tenant, got %T", resp)
}

func TestChangeMyPassword_SuccessReturns204(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.ChangeMyPassword(ctx, gen.ChangeMyPasswordRequestObject{
		Body: &gen.ChangeMyPasswordJSONRequestBody{
			CurrentPassword: "password123",
			NewPassword:     "newpassword456",
		},
	})

	require.NoError(t, err)
	_, is400 := resp.(gen.ChangeMyPassword400JSONResponse)
	_, is401 := resp.(gen.ChangeMyPassword401JSONResponse)
	assert.False(t, is400, "unexpected 400")
	assert.False(t, is401, "unexpected 401")
}

func TestChangeMyPassword_WrongCurrentPasswordReturns401(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.ChangeMyPassword(ctx, gen.ChangeMyPasswordRequestObject{
		Body: &gen.ChangeMyPasswordJSONRequestBody{
			CurrentPassword: "wrong-password",
			NewPassword:     "newpassword456",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.ChangeMyPassword401JSONResponse)
	assert.True(t, ok, "expected 401 for wrong password, got %T", resp)
}

func TestCreateUser_InvalidEmailReturns400(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.CreateUser(ctx, gen.CreateUserRequestObject{
		Body: &gen.CreateUserRequest{
			Email:       openapi_types.Email("not-an-email"),
			DisplayName: "Bad Email",
			Password:    "securepass123",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateUser400JSONResponse)
	assert.True(t, ok, "expected 400 for invalid email, got %T", resp)
}

func TestCreateUser_WithExplicitRole(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	role := gen.CreateUserRequestRole("viewer")
	resp, err := fix.handler.CreateUser(ctx, gen.CreateUserRequestObject{
		Body: &gen.CreateUserRequest{
			Email:       openapi_types.Email("viewer@example.com"),
			DisplayName: "Viewer User",
			Password:    "securepass123",
			Role:        &role,
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.CreateUser201JSONResponse)
	require.True(t, ok, "expected 201, got %T", resp)
	assert.Equal(t, openapi_types.Email("viewer@example.com"), jsonResp.Email)
}

func TestDeleteMyToken_UnauthorizedReturns403(t *testing.T) {
	fix := newTestFixture(t)

	ctx1 := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")
	createResp, err := fix.handler.CreateMyToken(ctx1, gen.CreateMyTokenRequestObject{
		Body: &gen.CreateTokenRequest{Name: "owned", Scopes: []gen.CreateTokenRequestScopes{"read"}},
	})
	require.NoError(t, err)
	created, found := createResp.(gen.CreateMyToken201JSONResponse)
	require.True(t, found)

	otherUserID := "99999999-9999-9999-9999-999999999999"
	ctx2 := withPrincipal(context.Background(), otherUserID, fix.tenantID, "admin")
	resp, err := fix.handler.DeleteMyToken(ctx2, gen.DeleteMyTokenRequestObject{Id: created.Id})

	require.NoError(t, err)
	_, found = resp.(gen.DeleteMyToken403JSONResponse)
	assert.True(t, found, "expected 403 for wrong user, got %T", resp)
}

func TestCreateSession_VisitSetsCookie(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.CreateSession(context.Background(), gen.CreateSessionRequestObject{
		Body: &gen.LoginRequest{
			Email:    openapi_types.Email("test@example.com"),
			Password: "password123",
		},
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitCreateSessionResponse(rec))
	assert.Equal(t, http.StatusOK, rec.Code)
	cookies := rec.Result().Cookies()
	var found bool
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			found = true
			assert.NotEmpty(t, cookie.Value)
		}
	}
	assert.True(t, found, "expected session cookie to be set")
}

func TestDeleteCurrentSession_VisitClearsCookie(t *testing.T) {
	fix := newTestFixture(t)

	resp, err := fix.handler.DeleteCurrentSession(context.Background(), gen.DeleteCurrentSessionRequestObject{})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitDeleteCurrentSessionResponse(rec))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestChangeMyPassword_VisitClearsCookie(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	resp, err := fix.handler.ChangeMyPassword(ctx, gen.ChangeMyPasswordRequestObject{
		Body: &gen.ChangeMyPasswordJSONRequestBody{
			CurrentPassword: "password123",
			NewPassword:     "newpassword456",
		},
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitChangeMyPasswordResponse(rec))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestListMyTokens_AfterCreateReturnsToken(t *testing.T) {
	fix := newTestFixture(t)
	ctx := withPrincipal(context.Background(), fix.userID, fix.tenantID, "admin")

	_, err := fix.handler.CreateMyToken(ctx, gen.CreateMyTokenRequestObject{
		Body: &gen.CreateTokenRequest{Name: "listed-token", Scopes: []gen.CreateTokenRequestScopes{"read"}},
	})
	require.NoError(t, err)

	resp, err := fix.handler.ListMyTokens(ctx, gen.ListMyTokensRequestObject{})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListMyTokens200JSONResponse)
	require.True(t, ok, "expected 200, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
	assert.Equal(t, "listed-token", jsonResp.Items[0].Name)
	assert.NotEmpty(t, jsonResp.Items[0].Prefix)
}
