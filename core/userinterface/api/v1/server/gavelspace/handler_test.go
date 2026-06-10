package gavelspace_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gscreate "github.com/usegavel/gavel/core/application/gavelspace/create"
	gsget "github.com/usegavel/gavel/core/application/gavelspace/get"
	gslist "github.com/usegavel/gavel/core/application/gavelspace/list"
	gsregisterproject "github.com/usegavel/gavel/core/application/gavelspace/registerproject"
	gsremoveproject "github.com/usegavel/gavel/core/application/gavelspace/removeproject"
	memgs "github.com/usegavel/gavel/core/infrastructure/gavelspace/memory"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/gavelspace"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func newTestHandler(listFinder *fakeListFinder, getFinder *fakeGetFinder) *gavelspace.Handler {
	repo := memgs.NewGavelspaceRepository()
	return gavelspace.New(gavelspace.Deps{
		ListGavelspaces:           gslist.NewHandler(listFinder),
		CreateGavelspace:          gscreate.NewHandler(repo),
		GetGavelspace:             gsget.NewHandler(getFinder),
		RegisterGavelspaceProject: gsregisterproject.NewHandler(repo),
		RemoveGavelspaceProject:   gsremoveproject.NewHandler(repo),
	})
}

func TestListGavelspaces_ReturnsItems(t *testing.T) {
	summary := testSummary()
	listFinder := &fakeListFinder{items: []gslist.GavelspaceSummary{summary}, total: 1}
	handler := newTestHandler(listFinder, &fakeGetFinder{})

	resp, err := handler.ListGavelspaces(context.Background(), gen.ListGavelspacesRequestObject{})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListGavelspaces200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
	assert.Equal(t, "gavel", jsonResp.Items[0].Name)
	assert.Equal(t, int32(3), jsonResp.Items[0].ProjectCount)
}

func TestListGavelspaces_EmptyList(t *testing.T) {
	listFinder := &fakeListFinder{items: nil, total: 0}
	handler := newTestHandler(listFinder, &fakeGetFinder{})

	resp, err := handler.ListGavelspaces(context.Background(), gen.ListGavelspacesRequestObject{})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListGavelspaces200JSONResponse)
	require.True(t, ok)
	assert.Empty(t, jsonResp.Items)
	assert.Nil(t, jsonResp.NextCursor)
}

func TestListGavelspaces_FinderError(t *testing.T) {
	listFinder := &fakeListFinder{err: errFake}
	handler := newTestHandler(listFinder, &fakeGetFinder{})

	resp, err := handler.ListGavelspaces(context.Background(), gen.ListGavelspacesRequestObject{})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetGavelspace_ReturnsDetail(t *testing.T) {
	detail := testDetail()
	getFinder := &fakeGetFinder{detail: detail}
	handler := newTestHandler(&fakeListFinder{}, getFinder)

	resp, err := handler.GetGavelspace(context.Background(), gen.GetGavelspaceRequestObject{Name: "gavel"})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetGavelspace200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	assert.Equal(t, "gavel", jsonResp.Name)
	require.Len(t, jsonResp.Projects, 1)
	assert.Equal(t, "core", jsonResp.Projects[0].Key)
	assert.Equal(t, "Core", jsonResp.Projects[0].Name)
	assert.Equal(t, "pass", jsonResp.Projects[0].LatestVerdict)
}

func TestGetGavelspace_NotFoundReturns404(t *testing.T) {
	getFinder := &fakeGetFinder{err: errNotFound}
	handler := newTestHandler(&fakeListFinder{}, getFinder)

	resp, err := handler.GetGavelspace(context.Background(), gen.GetGavelspaceRequestObject{Name: "missing"})

	require.NoError(t, err)
	_, ok := resp.(gen.GetGavelspace404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestCreateGavelspace_Success(t *testing.T) {
	repo := memgs.NewGavelspaceRepository()
	handler := gavelspace.New(gavelspace.Deps{
		ListGavelspaces:           gslist.NewHandler(&fakeListFinder{}),
		CreateGavelspace:          gscreate.NewHandler(repo),
		GetGavelspace:             gsget.NewHandler(&fakeGetFinder{}),
		RegisterGavelspaceProject: gsregisterproject.NewHandler(repo),
		RemoveGavelspaceProject:   gsremoveproject.NewHandler(repo),
	})

	body := gen.CreateGavelspaceRequest{Name: "test-gs"}
	resp, err := handler.CreateGavelspace(context.Background(), gen.CreateGavelspaceRequestObject{Body: &body})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.CreateGavelspace201JSONResponse)
	require.True(t, ok, "expected 201 JSON response, got %T", resp)
	assert.Equal(t, "test-gs", jsonResp.Name)
}

func TestCreateGavelspace_NilBody(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{})

	resp, err := handler.CreateGavelspace(context.Background(), gen.CreateGavelspaceRequestObject{Body: nil})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateGavelspace400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestCreateGavelspace_ConflictOnDuplicate(t *testing.T) {
	repo := &conflictRepo{inner: memgs.NewGavelspaceRepository()}
	handler := gavelspace.New(gavelspace.Deps{
		ListGavelspaces:           gslist.NewHandler(&fakeListFinder{}),
		CreateGavelspace:          gscreate.NewHandler(repo),
		GetGavelspace:             gsget.NewHandler(&fakeGetFinder{}),
		RegisterGavelspaceProject: gsregisterproject.NewHandler(repo),
		RemoveGavelspaceProject:   gsremoveproject.NewHandler(repo),
	})

	body := gen.CreateGavelspaceRequest{Name: "dup-gs"}
	resp, err := handler.CreateGavelspace(context.Background(), gen.CreateGavelspaceRequestObject{Body: &body})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateGavelspace409JSONResponse)
	assert.True(t, ok, "expected 409 response, got %T", resp)
}

func TestRegisterGavelspaceProject_NilBody(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{})

	resp, err := handler.RegisterGavelspaceProject(context.Background(), gen.RegisterGavelspaceProjectRequestObject{
		Name: "gavel",
		Body: nil,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.RegisterGavelspaceProject400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestRemoveGavelspaceProject_NotFound(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{})

	resp, err := handler.RemoveGavelspaceProject(context.Background(), gen.RemoveGavelspaceProjectRequestObject{
		Name:      "gavel",
		ProjectId: httpx.ParseUUIDOrZero("00000000-0000-0000-0000-000000000001"),
	})

	require.NoError(t, err)
	_, ok := resp.(gen.RemoveGavelspaceProject404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestCreateGavelspace_InvalidNameReturns400(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{})

	body := gen.CreateGavelspaceRequest{Name: ""}
	resp, err := handler.CreateGavelspace(context.Background(), gen.CreateGavelspaceRequestObject{Body: &body})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateGavelspace400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestGetGavelspace_FinderErrorReturnsError(t *testing.T) {
	getFinder := &fakeGetFinder{err: errFake}
	handler := newTestHandler(&fakeListFinder{}, getFinder)

	resp, err := handler.GetGavelspace(context.Background(), gen.GetGavelspaceRequestObject{Name: "gavel"})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestRegisterGavelspaceProject_SuccessReturns204(t *testing.T) {
	repo := memgs.NewGavelspaceRepository()
	handler := gavelspace.New(gavelspace.Deps{
		ListGavelspaces:           gslist.NewHandler(&fakeListFinder{}),
		CreateGavelspace:          gscreate.NewHandler(repo),
		GetGavelspace:             gsget.NewHandler(&fakeGetFinder{}),
		RegisterGavelspaceProject: gsregisterproject.NewHandler(repo),
		RemoveGavelspaceProject:   gsremoveproject.NewHandler(repo),
	})

	createBody := gen.CreateGavelspaceRequest{Name: "test-reg"}
	_, err := handler.CreateGavelspace(context.Background(), gen.CreateGavelspaceRequestObject{Body: &createBody})
	require.NoError(t, err)

	projectID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	regBody := gen.RegisterGavelspaceProjectRequest{ProjectId: projectID, TargetPattern: "//core/..."}
	resp, err := handler.RegisterGavelspaceProject(context.Background(), gen.RegisterGavelspaceProjectRequestObject{
		Name: "test-reg", Body: &regBody,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.RegisterGavelspaceProject204Response)
	assert.True(t, ok, "expected 204 response, got %T", resp)
}

func TestRegisterGavelspaceProject_NotFoundReturns404(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{})

	projectID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	regBody := gen.RegisterGavelspaceProjectRequest{ProjectId: projectID, TargetPattern: "//core/..."}
	resp, err := handler.RegisterGavelspaceProject(context.Background(), gen.RegisterGavelspaceProjectRequestObject{
		Name: "missing", Body: &regBody,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.RegisterGavelspaceProject404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestRegisterGavelspaceProject_DuplicateReturns409(t *testing.T) {
	repo := memgs.NewGavelspaceRepository()
	handler := gavelspace.New(gavelspace.Deps{
		ListGavelspaces:           gslist.NewHandler(&fakeListFinder{}),
		CreateGavelspace:          gscreate.NewHandler(repo),
		GetGavelspace:             gsget.NewHandler(&fakeGetFinder{}),
		RegisterGavelspaceProject: gsregisterproject.NewHandler(repo),
		RemoveGavelspaceProject:   gsremoveproject.NewHandler(repo),
	})

	createBody := gen.CreateGavelspaceRequest{Name: "dup-reg"}
	_, err := handler.CreateGavelspace(context.Background(), gen.CreateGavelspaceRequestObject{Body: &createBody})
	require.NoError(t, err)

	projectID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	regBody := gen.RegisterGavelspaceProjectRequest{ProjectId: projectID, TargetPattern: "//core/..."}
	_, err = handler.RegisterGavelspaceProject(context.Background(), gen.RegisterGavelspaceProjectRequestObject{
		Name: "dup-reg", Body: &regBody,
	})
	require.NoError(t, err)

	resp, err := handler.RegisterGavelspaceProject(context.Background(), gen.RegisterGavelspaceProjectRequestObject{
		Name: "dup-reg", Body: &regBody,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.RegisterGavelspaceProject409JSONResponse)
	assert.True(t, ok, "expected 409 response, got %T", resp)
}

func TestRemoveGavelspaceProject_SuccessReturns204(t *testing.T) {
	repo := memgs.NewGavelspaceRepository()
	handler := gavelspace.New(gavelspace.Deps{
		ListGavelspaces:           gslist.NewHandler(&fakeListFinder{}),
		CreateGavelspace:          gscreate.NewHandler(repo),
		GetGavelspace:             gsget.NewHandler(&fakeGetFinder{}),
		RegisterGavelspaceProject: gsregisterproject.NewHandler(repo),
		RemoveGavelspaceProject:   gsremoveproject.NewHandler(repo),
	})

	createBody := gen.CreateGavelspaceRequest{Name: "rm-test"}
	_, err := handler.CreateGavelspace(context.Background(), gen.CreateGavelspaceRequestObject{Body: &createBody})
	require.NoError(t, err)

	projectID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	regBody := gen.RegisterGavelspaceProjectRequest{ProjectId: projectID, TargetPattern: "//core/..."}
	_, err = handler.RegisterGavelspaceProject(context.Background(), gen.RegisterGavelspaceProjectRequestObject{
		Name: "rm-test", Body: &regBody,
	})
	require.NoError(t, err)

	resp, err := handler.RemoveGavelspaceProject(context.Background(), gen.RemoveGavelspaceProjectRequestObject{
		Name:      "rm-test",
		ProjectId: projectID,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.RemoveGavelspaceProject204Response)
	assert.True(t, ok, "expected 204 response, got %T", resp)
}
