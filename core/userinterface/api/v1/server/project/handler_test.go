package project_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	projectcreate "github.com/usegavel/gavel/core/application/project/create"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	projectlist "github.com/usegavel/gavel/core/application/project/list"
	updatelanguages "github.com/usegavel/gavel/core/application/project/updatelanguages"
	updatequalitygate "github.com/usegavel/gavel/core/application/project/updatequalitygate"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	domainproject "github.com/usegavel/gavel/core/domain/project/model"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/project"
)

func newTestHandler(listFinder *fakeListFinder, getByKeyFinder *fakeGetByKeyFinder) *project.Handler {
	repo := memproject.NewProjectRepository()
	return project.New(project.Deps{
		ListProjects:             projectlist.NewHandler(listFinder),
		GetProject:               projectgetbykey.NewHandler(getByKeyFinder),
		CreateProject:            projectcreate.NewHandler(repo),
		UpdateProjectLanguages:   updatelanguages.NewHandler(repo),
		UpdateProjectQualityGate: updatequalitygate.NewHandler(repo),
	})
}

func TestListProjects_ReturnsItems(t *testing.T) {
	summary := testProjectSummary()
	listFinder := &fakeListFinder{items: []projectlist.ProjectSummary{summary}, total: 1}
	handler := newTestHandler(listFinder, &fakeGetByKeyFinder{})

	resp, err := handler.ListProjects(authContext(), gen.ListProjectsRequestObject{})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListProjects200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
	assert.Equal(t, "core", jsonResp.Items[0].Key)
	assert.Equal(t, "Core Module", jsonResp.Items[0].Name)
	assert.Equal(t, "main", jsonResp.Items[0].DefaultBranch)
	assert.Equal(t, "pass", jsonResp.Items[0].LatestVerdict)
	assert.Equal(t, int32(42), jsonResp.Items[0].TotalFindings)
}

func TestListProjects_EmptyList(t *testing.T) {
	listFinder := &fakeListFinder{items: nil, total: 0}
	handler := newTestHandler(listFinder, &fakeGetByKeyFinder{})

	resp, err := handler.ListProjects(authContext(), gen.ListProjectsRequestObject{})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListProjects200JSONResponse)
	require.True(t, ok)
	assert.Empty(t, jsonResp.Items)
	assert.Nil(t, jsonResp.NextCursor)
}

func TestListProjects_WithPagination(t *testing.T) {
	summaries := make([]projectlist.ProjectSummary, 20)
	for idx := range summaries {
		summaries[idx] = testProjectSummary()
	}
	listFinder := &fakeListFinder{items: summaries, total: 50}
	handler := newTestHandler(listFinder, &fakeGetByKeyFinder{})

	limit := gen.Limit(20)
	resp, err := handler.ListProjects(authContext(), gen.ListProjectsRequestObject{
		Params: gen.ListProjectsParams{Limit: &limit},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListProjects200JSONResponse)
	require.True(t, ok)
	assert.Len(t, jsonResp.Items, 20)
	require.NotNil(t, jsonResp.NextCursor)
}

func TestGetProject_ReturnsDetail(t *testing.T) {
	detail := testProjectDetail()
	getByKeyFinder := &fakeGetByKeyFinder{detail: detail}
	handler := newTestHandler(&fakeListFinder{}, getByKeyFinder)

	resp, err := handler.GetProject(authContext(), gen.GetProjectRequestObject{Key: "core"})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetProject200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	assert.Equal(t, "core", jsonResp.Key)
	assert.Equal(t, "Core Module", jsonResp.Name)
	assert.Equal(t, "main", jsonResp.DefaultBranch)
	assert.Equal(t, "pass", jsonResp.LatestVerdict)
	assert.Equal(t, int32(42), jsonResp.TotalFindings)
	assert.Equal(t, "//core/...", jsonResp.TargetPattern)
	assert.Equal(t, []string{"go"}, jsonResp.Languages)
	require.Len(t, jsonResp.QualityGateRules, 1)
	assert.Equal(t, "code_quality", jsonResp.QualityGateRules[0].Subtype)
	assert.Equal(t, int32(10), jsonResp.SeverityCounts["error"])
}

func TestGetProject_NotFound(t *testing.T) {
	getByKeyFinder := &fakeGetByKeyFinder{err: errFake}
	handler := newTestHandler(&fakeListFinder{}, getByKeyFinder)

	resp, err := handler.GetProject(authContext(), gen.GetProjectRequestObject{Key: "missing"})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateProject_Success(t *testing.T) {
	repo := memproject.NewProjectRepository()
	handler := project.New(project.Deps{
		ListProjects:             projectlist.NewHandler(&fakeListFinder{}),
		GetProject:               projectgetbykey.NewHandler(&fakeGetByKeyFinder{}),
		CreateProject:            projectcreate.NewHandler(repo),
		UpdateProjectLanguages:   updatelanguages.NewHandler(repo),
		UpdateProjectQualityGate: updatequalitygate.NewHandler(repo),
	})

	tp := "//new/..."
	body := gen.CreateProjectRequest{
		Key:           "new-proj",
		Name:          "New Project",
		TargetPattern: &tp,
	}
	resp, err := handler.CreateProject(authContext(), gen.CreateProjectRequestObject{Body: &body})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.CreateProject201JSONResponse)
	require.True(t, ok, "expected 201 JSON response, got %T", resp)
	assert.NotEmpty(t, jsonResp.ProjectId.String())
}

func TestCreateProject_NilBody(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetByKeyFinder{})

	resp, err := handler.CreateProject(authContext(), gen.CreateProjectRequestObject{Body: nil})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateProject400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestCreateProject_InvalidKey(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetByKeyFinder{})

	body := gen.CreateProjectRequest{Key: "", Name: "Unnamed"}
	resp, err := handler.CreateProject(authContext(), gen.CreateProjectRequestObject{Body: &body})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateProject400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestCreateProject_ConflictOnDuplicate(t *testing.T) {
	repo := newConflictRepo()
	handler := project.New(project.Deps{
		ListProjects:             projectlist.NewHandler(&fakeListFinder{}),
		GetProject:               projectgetbykey.NewHandler(&fakeGetByKeyFinder{}),
		CreateProject:            projectcreate.NewHandler(repo),
		UpdateProjectLanguages:   updatelanguages.NewHandler(repo),
		UpdateProjectQualityGate: updatequalitygate.NewHandler(repo),
	})

	tp := "//dup/..."
	body := gen.CreateProjectRequest{Key: "dup", Name: "Dup", TargetPattern: &tp}
	resp, err := handler.CreateProject(authContext(), gen.CreateProjectRequestObject{Body: &body})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateProject409JSONResponse)
	assert.True(t, ok, "expected 409 response, got %T", resp)
}

func TestGetProject_NotFoundReturns404(t *testing.T) {
	getByKeyFinder := &fakeGetByKeyFinder{err: errNotFound}
	handler := newTestHandler(&fakeListFinder{}, getByKeyFinder)

	resp, err := handler.GetProject(authContext(), gen.GetProjectRequestObject{Key: "missing"})

	require.NoError(t, err)
	_, ok := resp.(gen.GetProject404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestUpdateProjectQualityGate_NilBody(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetByKeyFinder{detail: testProjectDetail()})

	resp, err := handler.UpdateProjectQualityGate(authContext(), gen.UpdateProjectQualityGateRequestObject{
		Key:  "core",
		Body: nil,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UpdateProjectQualityGate400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestUpdateProjectQualityGate_NotFoundReturns404(t *testing.T) {
	getByKeyFinder := &fakeGetByKeyFinder{err: errNotFound}
	handler := newTestHandler(&fakeListFinder{}, getByKeyFinder)

	body := gen.QualityGate{Rules: []gen.QualityGateRule{}}
	resp, err := handler.UpdateProjectQualityGate(authContext(), gen.UpdateProjectQualityGateRequestObject{
		Key:  "missing",
		Body: &body,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UpdateProjectQualityGate404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestUpdateProjectQualityGate_Success(t *testing.T) {
	detail := testProjectDetail()
	repo := memproject.NewProjectRepository()
	tid, err := tenant.ParseTenantID(testTenant)
	require.NoError(t, err)
	proj, err := domainproject.NewProject(tid, detail.Key, detail.Name, detail.TargetPattern)
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), proj))

	handler := project.New(project.Deps{
		ListProjects:             projectlist.NewHandler(&fakeListFinder{}),
		GetProject:               projectgetbykey.NewHandler(&fakeGetByKeyFinder{detail: detailWithID(proj.ID().String())}),
		CreateProject:            projectcreate.NewHandler(repo),
		UpdateProjectLanguages:   updatelanguages.NewHandler(repo),
		UpdateProjectQualityGate: updatequalitygate.NewHandler(repo),
	})

	minRes := int32(1)
	body := gen.QualityGate{Rules: []gen.QualityGateRule{
		{
			Subtype: "code_quality",
			Strategy: gen.QualityGateStrategy{
				Type: "count_by_severity",
				CountBySeverity: &gen.CountBySeverityStrategy{
					MaxError:   0,
					MaxWarning: 10,
					MaxNote:    50,
				},
			},
			MinResolved: &minRes,
		},
		{
			Subtype: "coverage",
			Strategy: gen.QualityGateStrategy{
				Type:          "min_percentage",
				MinPercentage: &gen.MinPercentageStrategy{Min: 90.0},
			},
		},
		{
			Subtype: "architecture",
			Strategy: gen.QualityGateStrategy{
				Type:          "max_violations",
				MaxViolations: &gen.MaxViolationsStrategy{Max: 0},
			},
		},
	}}
	resp, err := handler.UpdateProjectQualityGate(authContext(), gen.UpdateProjectQualityGateRequestObject{
		Key:  "core",
		Body: &body,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UpdateProjectQualityGate204Response)
	assert.True(t, ok, "expected 204 response, got %T", resp)
}

func TestUpdateProjectLanguages_NilBody(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetByKeyFinder{detail: testProjectDetail()})

	resp, err := handler.UpdateProjectLanguages(authContext(), gen.UpdateProjectLanguagesRequestObject{
		Key:  "core",
		Body: nil,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UpdateProjectLanguages400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestUpdateProjectLanguages_NotFoundReturns404(t *testing.T) {
	getByKeyFinder := &fakeGetByKeyFinder{err: errNotFound}
	handler := newTestHandler(&fakeListFinder{}, getByKeyFinder)

	body := gen.UpdateLanguagesRequest{Languages: []string{"go"}}
	resp, err := handler.UpdateProjectLanguages(authContext(), gen.UpdateProjectLanguagesRequestObject{
		Key:  "missing",
		Body: &body,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UpdateProjectLanguages404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestUpdateProjectLanguages_Success(t *testing.T) {
	detail := testProjectDetail()
	repo := memproject.NewProjectRepository()
	tid2, err := tenant.ParseTenantID(testTenant)
	require.NoError(t, err)
	proj, err := domainproject.NewProject(tid2, detail.Key, detail.Name, detail.TargetPattern)
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), proj))

	handler := project.New(project.Deps{
		ListProjects:             projectlist.NewHandler(&fakeListFinder{}),
		GetProject:               projectgetbykey.NewHandler(&fakeGetByKeyFinder{detail: detailWithID(proj.ID().String())}),
		CreateProject:            projectcreate.NewHandler(repo),
		UpdateProjectLanguages:   updatelanguages.NewHandler(repo),
		UpdateProjectQualityGate: updatequalitygate.NewHandler(repo),
	})

	body := gen.UpdateLanguagesRequest{Languages: []string{"go", "typescript"}}
	resp, err := handler.UpdateProjectLanguages(authContext(), gen.UpdateProjectLanguagesRequestObject{
		Key:  "core",
		Body: &body,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UpdateProjectLanguages204Response)
	assert.True(t, ok, "expected 204 response, got %T", resp)
}

func TestListProjects_FinderError(t *testing.T) {
	listFinder := &fakeListFinder{err: errFake}
	handler := newTestHandler(listFinder, &fakeGetByKeyFinder{})

	resp, err := handler.ListProjects(authContext(), gen.ListProjectsRequestObject{})

	require.Error(t, err)
	assert.Nil(t, resp)
}
