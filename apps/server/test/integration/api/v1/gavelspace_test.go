package v1integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListGavelspaces_Empty(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/gavelspaces", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var body struct {
		Items      []any   `json:"items"`
		NextCursor *string `json:"next_cursor"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Empty(t, body.Items)
	require.Nil(t, body.NextCursor)
}

func TestCreateGavelspace_AdminSuccess(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPost, "/gavelspaces", map[string]string{"name": "acme"}, cookie)
	require.Equal(t, http.StatusCreated, res.Code, res.Body.String())

	var body struct {
		Name string `json:"name"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Equal(t, "acme", body.Name)
}

func TestCreateGavelspace_RequiresAdmin(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)

	res := f.do(t, http.MethodPost, "/gavelspaces", map[string]string{"name": "acme"}, cookie)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestGetGavelspace_NotFound(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/gavelspaces/missing", nil, cookie)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestGetGavelspace_Success(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)

	createRes := fixture.do(t, http.MethodPost, "/gavelspaces", map[string]string{"name": "acme"}, cookie)
	require.Equal(t, http.StatusCreated, createRes.Code)

	res := fixture.do(t, http.MethodGet, "/gavelspaces/acme", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var body struct {
		Name     string `json:"name"`
		Projects []any  `json:"projects"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Equal(t, "acme", body.Name)
	require.Empty(t, body.Projects)
}

func TestRegisterGavelspaceProject_Success(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)

	createRes := fixture.do(t, http.MethodPost, "/gavelspaces", map[string]string{"name": "acme"}, cookie)
	require.Equal(t, http.StatusCreated, createRes.Code)

	projectID := uuid.NewString()
	res := fixture.do(t, http.MethodPost, "/gavelspaces/acme/projects", map[string]any{
		"project_id":     projectID,
		"target_pattern": "//apps/server/...",
	}, cookie)
	require.Equal(t, http.StatusNoContent, res.Code, res.Body.String())

	getRes := fixture.do(t, http.MethodGet, "/gavelspaces/acme", nil, cookie)
	require.Equal(t, http.StatusOK, getRes.Code)
	var detail struct {
		Projects []struct {
			ID string `json:"id"`
		} `json:"projects"`
	}
	mustDecode(t, getRes.Body.Bytes(), &detail)
	require.Len(t, detail.Projects, 1)
	require.Equal(t, projectID, detail.Projects[0].ID)
}

func TestRegisterGavelspaceProject_RequiresAdmin(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)

	res := f.do(t, http.MethodPost, "/gavelspaces/acme/projects", map[string]any{
		"project_id":     uuid.NewString(),
		"target_pattern": "//x/...",
	}, cookie)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestRegisterGavelspaceProject_NotFound(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPost, "/gavelspaces/missing/projects", map[string]any{
		"project_id":     uuid.NewString(),
		"target_pattern": "//x/...",
	}, cookie)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestRemoveGavelspaceProject_Success(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)

	createRes := fixture.do(t, http.MethodPost, "/gavelspaces", map[string]string{"name": "acme"}, cookie)
	require.Equal(t, http.StatusCreated, createRes.Code)

	projectID := uuid.NewString()
	reg := fixture.do(t, http.MethodPost, "/gavelspaces/acme/projects", map[string]any{
		"project_id":     projectID,
		"target_pattern": "//apps/...",
	}, cookie)
	require.Equal(t, http.StatusNoContent, reg.Code)

	del := fixture.do(t, http.MethodDelete, "/gavelspaces/acme/projects/"+projectID, nil, cookie)
	require.Equal(t, http.StatusNoContent, del.Code, del.Body.String())
}
