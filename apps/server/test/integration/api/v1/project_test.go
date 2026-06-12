package v1integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	projectview "github.com/usegavel/gavel/core/application/project/projectview"
)

func seedProject(f *testFixture, key, name string) string {
	projectID := uuid.NewString()
	f.projects.put(&projectview.ProjectDetail{
		ID:               projectID,
		Key:              key,
		Name:             name,
		DefaultBranch:    "main",
		LatestVerdict:    "pending",
		TotalFindings:    0,
		CreatedAt:        time.Date(2026, time.June, 6, 9, 0, 0, 0, time.UTC),
		TargetPattern:    "//...",
		Languages:        []string{"go"},
		QualityGateRules: []projectview.QualityGateRuleView{},
		SeverityCounts:   map[string]int{},
	})
	return projectID
}

func TestListProjects_Empty(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/projects", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var body struct {
		Items []any `json:"items"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Empty(t, body.Items)
}

func TestGetProject_Success(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)
	projectID := seedProject(f, "core", "Core")

	res := f.do(t, http.MethodGet, "/projects/core", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var body struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Equal(t, projectID, body.ID)
	require.Equal(t, "core", body.Key)
}

func TestGetProject_NotFound(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)
	res := f.do(t, http.MethodGet, "/projects/missing", nil, cookie)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestCreateProject_RequiresAdmin(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)
	res := f.do(t, http.MethodPost, "/projects", map[string]string{
		"key": "x", "name": "X", "target_pattern": "//x/...",
	}, cookie)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestCreateProject_AdminSuccess(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)
	res := f.do(t, http.MethodPost, "/projects", map[string]string{
		"key": "newproj", "name": "New", "target_pattern": "//new/...",
	}, cookie)
	require.Equal(t, http.StatusCreated, res.Code, res.Body.String())
	var body struct {
		ProjectID string `json:"project_id"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.NotEmpty(t, body.ProjectID)
}

func TestUpdateProjectQualityGate_RequiresAdmin(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)
	res := f.do(t, http.MethodPut, "/projects/core/quality-gate", map[string]any{"rules": []any{}}, cookie)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestUpdateProjectLanguages_RequiresAdmin(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)
	res := f.do(t, http.MethodPut, "/projects/core/languages", map[string]any{"languages": []string{"go"}}, cookie)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestListProjectCaseFiles_Empty(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)
	seedProject(f, "core", "Core")

	res := f.do(t, http.MethodGet, "/projects/core/casefiles", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var body struct {
		Items []any `json:"items"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Empty(t, body.Items)
}

func TestListProjectPleadings_Empty(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)
	seedProject(f, "core", "Core")

	res := f.do(t, http.MethodGet, "/projects/core/pleadings", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
}

func TestFileProjectPleading_RequiresAdmin(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)
	res := f.do(t, http.MethodPost, "/projects/core/pleadings", map[string]any{
		"number": 1, "title": "x", "petitioner": "p",
		"source_branch": "f", "target_branch": "main", "commit_sha": "abc",
	}, cookie)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestGetProjectBaseline_Success(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)

	createRes := fixture.do(t, http.MethodPost, "/projects", map[string]string{
		"key": "core", "name": "Core", "target_pattern": "//core/...",
	}, cookie)
	require.Equal(t, http.StatusCreated, createRes.Code, createRes.Body.String())
	var created struct{ ProjectID string `json:"project_id"` }
	mustDecode(t, createRes.Body.Bytes(), &created)

	p, err := fixture.projRepo.FindByName(context.Background(), "Core")
	require.NoError(t, err)
	p.UpdateBaseline("main", []string{"fp1"}, []string{"a1"}, nil, nil)
	require.NoError(t, fixture.projRepo.Save(context.Background(), p))

	fixture.projects.put(&projectview.ProjectDetail{
		ID: created.ProjectID, Key: "core", Name: "Core", DefaultBranch: "main",
		TargetPattern: "//core/...", Languages: []string{"go"},
		QualityGateRules: []projectview.QualityGateRuleView{}, SeverityCounts: map[string]int{},
	})

	res := fixture.do(t, http.MethodGet, "/projects/core/baseline", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var body struct {
		Fingerprints []string `json:"fingerprints"`
		HasPrevious  bool     `json:"has_previous"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Equal(t, []string{"fp1"}, body.Fingerprints)
	require.True(t, body.HasPrevious)
}
