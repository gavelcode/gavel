package v1integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateUser_Success(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPost, "/admin/users", map[string]string{
		"email":        "newuser@example.com",
		"display_name": "New",
		"password":     "newpass!123",
		"role":         "maintainer",
	}, cookie)
	require.Equal(t, http.StatusCreated, res.Code, res.Body.String())

	var body struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.NotEmpty(t, body.ID)
	require.Equal(t, "newuser@example.com", body.Email)
	require.Equal(t, "maintainer", body.Role)
}

func TestCreateUser_RequiresAdmin(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)

	res := f.do(t, http.MethodPost, "/admin/users", map[string]string{
		"email":        "x@example.com",
		"display_name": "X",
		"password":     "passpass!",
	}, cookie)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPost, "/admin/users", map[string]string{
		"email":        viewerEmail,
		"display_name": "dup",
		"password":     "passpass!",
		"role":         "viewer",
	}, cookie)
	require.Equal(t, http.StatusConflict, res.Code)
}

func TestCreateUser_Unauthenticated(t *testing.T) {
	f := newTestFixture(t)

	res := f.do(t, http.MethodPost, "/admin/users", map[string]string{
		"email":        "x@example.com",
		"display_name": "X",
		"password":     "passpass!",
	}, nil)
	require.Equal(t, http.StatusUnauthorized, res.Code)
}
