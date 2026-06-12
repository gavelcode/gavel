package v1integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSession_Success(t *testing.T) {
	f := newTestFixture(t)

	res := f.do(t, http.MethodPost, "/sessions", map[string]string{
		"email":    adminEmail,
		"password": adminPassword,
	}, nil)

	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var profile struct {
		ID                 string `json:"id"`
		Email              string `json:"email"`
		Role               string `json:"role"`
		MustChangePassword bool   `json:"must_change_password"`
	}
	mustDecode(t, res.Body.Bytes(), &profile)
	require.Equal(t, adminEmail, profile.Email)
	require.Equal(t, "admin", profile.Role)
	require.False(t, profile.MustChangePassword)

	var found bool
	for _, c := range res.Result().Cookies() {
		if c.Name == sessionCookieName {
			found = true
			require.NotEmpty(t, c.Value)
			require.True(t, c.HttpOnly)
		}
	}
	require.True(t, found, "expected session cookie")
}

func TestCreateSession_InvalidPassword(t *testing.T) {
	f := newTestFixture(t)

	res := f.do(t, http.MethodPost, "/sessions", map[string]string{
		"email":    adminEmail,
		"password": "wrong",
	}, nil)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	require.Contains(t, res.Body.String(), "invalid credentials")
}

func TestCreateSession_MissingBody(t *testing.T) {
	f := newTestFixture(t)

	res := f.do(t, http.MethodPost, "/sessions", map[string]string{}, nil)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestGetMe_Unauthenticated(t *testing.T) {
	f := newTestFixture(t)
	res := f.do(t, http.MethodGet, "/me", nil, nil)
	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestGetMe_AfterLogin(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/me", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var profile struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	mustDecode(t, res.Body.Bytes(), &profile)
	require.Equal(t, adminEmail, profile.Email)
	require.Equal(t, "admin", profile.Role)
}

func TestDeleteCurrentSession_ClearsCookie(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodDelete, "/sessions/current", nil, cookie)
	require.Equal(t, http.StatusNoContent, res.Code)

	var cleared bool
	for _, c := range res.Result().Cookies() {
		if c.Name == sessionCookieName && c.MaxAge < 0 {
			cleared = true
		}
	}
	require.True(t, cleared, "expected session cookie to be cleared")
}

func TestDeleteCurrentSession_RevokesServerSide(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)

	res := fixture.do(t, http.MethodGet, "/me", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code)

	res = fixture.do(t, http.MethodDelete, "/sessions/current", nil, cookie)
	require.Equal(t, http.StatusNoContent, res.Code)

	res = fixture.do(t, http.MethodGet, "/me", nil, cookie)
	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestChangeMyPassword_Success(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, viewerEmail, viewerPassword)

	res := fixture.do(t, http.MethodPost, "/me/password", map[string]string{
		"current_password": viewerPassword,
		"new_password":     "newSecret!23",
	}, cookie)

	require.Equal(t, http.StatusNoContent, res.Code, res.Body.String())

	loginRes := fixture.do(t, http.MethodPost, "/sessions", map[string]string{
		"email":    viewerEmail,
		"password": "newSecret!23",
	}, nil)
	require.Equal(t, http.StatusOK, loginRes.Code)
}

func TestChangeMyPassword_WrongCurrent(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)

	res := f.do(t, http.MethodPost, "/me/password", map[string]string{
		"current_password": "wrong",
		"new_password":     "newSecret!23",
	}, cookie)

	require.Equal(t, http.StatusUnauthorized, res.Code)
}
