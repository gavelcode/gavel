package v1integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListMyTokens_EmptyForFreshUser(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/me/tokens", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var body struct {
		Items      []any   `json:"items"`
		NextCursor *string `json:"next_cursor"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Empty(t, body.Items)
	require.Nil(t, body.NextCursor)
}

func TestListMyTokens_Unauthenticated(t *testing.T) {
	f := newTestFixture(t)
	res := f.do(t, http.MethodGet, "/me/tokens", nil, nil)
	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestCreateMyToken_Success(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPost, "/me/tokens", map[string]any{
		"name":   "ci",
		"scopes": []string{"ingest"},
	}, cookie)
	require.Equal(t, http.StatusCreated, res.Code, res.Body.String())

	var body struct {
		ID     string   `json:"id"`
		Name   string   `json:"name"`
		Token  string   `json:"token"`
		Prefix string   `json:"prefix"`
		Scopes []string `json:"scopes"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.NotEmpty(t, body.ID)
	require.Equal(t, "ci", body.Name)
	require.NotEmpty(t, body.Token)
	require.NotEmpty(t, body.Prefix)
	require.Equal(t, []string{"ingest"}, body.Scopes)
}

func TestCreateMyToken_AdminScopeRequiresAdminRole(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)

	res := f.do(t, http.MethodPost, "/me/tokens", map[string]any{
		"name":   "evil",
		"scopes": []string{"admin"},
	}, cookie)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestDeleteMyToken_Success(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)

	createRes := fixture.do(t, http.MethodPost, "/me/tokens", map[string]any{
		"name":   "scratch",
		"scopes": []string{"ingest"},
	}, cookie)
	require.Equal(t, http.StatusCreated, createRes.Code)
	var created struct {
		ID string `json:"id"`
	}
	mustDecode(t, createRes.Body.Bytes(), &created)

	delRes := fixture.do(t, http.MethodDelete, "/me/tokens/"+created.ID, nil, cookie)
	require.Equal(t, http.StatusNoContent, delRes.Code, delRes.Body.String())
}

func TestDeleteMyToken_NotFound(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodDelete, "/me/tokens/3bdf1c30-1234-4abc-9def-0123456789ab", nil, cookie)
	require.Equal(t, http.StatusNotFound, res.Code)
}
