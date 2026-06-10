package pleading_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pleadingfile "github.com/usegavel/gavel/core/application/pleading/file"
	pleadingget "github.com/usegavel/gavel/core/application/pleading/get"
	pleadinglist "github.com/usegavel/gavel/core/application/pleading/list"
	pleadingresolve "github.com/usegavel/gavel/core/application/pleading/resolve"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	mempleading "github.com/usegavel/gavel/core/infrastructure/pleading/memory"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/pleading"
)

func newTestHandler(listFinder *fakeListFinder, getFinder *fakeGetFinder, getByKeyFinder *fakeGetByKeyFinder) *pleading.Handler {
	repo := mempleading.NewPleadingRepository()
	return pleading.New(pleading.Deps{
		ListPleadings:       pleadinglist.NewHandler(listFinder),
		GetPleading:         pleadingget.NewHandler(getFinder),
		FilePleading:        pleadingfile.NewHandler(repo),
		ResolvePleading:     pleadingresolve.NewHandler(repo),
		ResolveProjectByKey: projectgetbykey.NewHandler(getByKeyFinder),
	})
}

func TestListPleadings_ReturnsItems(t *testing.T) {
	summary := testPleadingSummary()
	listFinder := &fakeListFinder{items: []pleadinglist.PleadingSummary{summary}, total: 1}
	handler := newTestHandler(listFinder, &fakeGetFinder{}, &fakeGetByKeyFinder{})

	resp, err := handler.ListPleadings(context.Background(), gen.ListPleadingsRequestObject{Params: gen.ListPleadingsParams{}})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListPleadings200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
	assert.Equal(t, int32(42), jsonResp.Items[0].Number)
	assert.Equal(t, "Add login feature", jsonResp.Items[0].Title)
	assert.Equal(t, "dev@example.com", jsonResp.Items[0].Petitioner)
	assert.Equal(t, "feat/login", jsonResp.Items[0].SourceBranch)
	assert.Equal(t, "main", jsonResp.Items[0].TargetBranch)
	assert.Equal(t, "abc123", jsonResp.Items[0].CommitSha)
	assert.Equal(t, gen.PleadingStatus("open"), jsonResp.Items[0].Status)
}

func TestListPleadings_EmptyList(t *testing.T) {
	listFinder := &fakeListFinder{items: nil, total: 0}
	handler := newTestHandler(listFinder, &fakeGetFinder{}, &fakeGetByKeyFinder{})

	resp, err := handler.ListPleadings(context.Background(), gen.ListPleadingsRequestObject{Params: gen.ListPleadingsParams{}})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListPleadings200JSONResponse)
	require.True(t, ok)
	assert.Empty(t, jsonResp.Items)
	assert.Nil(t, jsonResp.NextCursor)
}

func TestListPleadings_FinderError(t *testing.T) {
	listFinder := &fakeListFinder{err: errFake}
	handler := newTestHandler(listFinder, &fakeGetFinder{}, &fakeGetByKeyFinder{})

	resp, err := handler.ListPleadings(context.Background(), gen.ListPleadingsRequestObject{Params: gen.ListPleadingsParams{}})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetPleading_ReturnsDetail(t *testing.T) {
	detail := testPleadingDetail()
	getFinder := &fakeGetFinder{detail: detail}
	handler := newTestHandler(&fakeListFinder{}, getFinder, &fakeGetByKeyFinder{})

	resp, err := handler.GetPleading(context.Background(), gen.GetPleadingRequestObject{
		Id: httpx.ParseUUIDOrZero("44444444-4444-4444-4444-444444444444"),
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetPleading200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	assert.Equal(t, int32(42), jsonResp.Number)
	assert.Equal(t, "Add login feature", jsonResp.Title)
	assert.Equal(t, "dev@example.com", jsonResp.Petitioner)
	assert.Equal(t, "feat/login", jsonResp.SourceBranch)
	assert.Equal(t, "main", jsonResp.TargetBranch)
	assert.Equal(t, "abc123", jsonResp.CommitSha)
	assert.Equal(t, gen.PleadingStatus("open"), jsonResp.Status)
}

func TestGetPleading_NotFoundReturns404(t *testing.T) {
	getFinder := &fakeGetFinder{err: errNotFound}
	handler := newTestHandler(&fakeListFinder{}, getFinder, &fakeGetByKeyFinder{})

	resp, err := handler.GetPleading(context.Background(), gen.GetPleadingRequestObject{
		Id: httpx.ParseUUIDOrZero("44444444-4444-4444-4444-444444444444"),
	})

	require.NoError(t, err)
	_, ok := resp.(gen.GetPleading404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestResolvePleading_NilBodyReturns400(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{}, &fakeGetByKeyFinder{})

	resp, err := handler.ResolvePleading(context.Background(), gen.ResolvePleadingRequestObject{
		Id:   httpx.ParseUUIDOrZero("44444444-4444-4444-4444-444444444444"),
		Body: nil,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.ResolvePleading400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestListProjectPleadings_ReturnsItems(t *testing.T) {
	summary := testPleadingSummary()
	listFinder := &fakeListFinder{items: []pleadinglist.PleadingSummary{summary}, total: 1}
	getByKeyFinder := &fakeGetByKeyFinder{detail: testProjectDetail()}
	handler := newTestHandler(listFinder, &fakeGetFinder{}, getByKeyFinder)

	resp, err := handler.ListProjectPleadings(context.Background(), gen.ListProjectPleadingsRequestObject{Key: "core"})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListProjectPleadings200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
	assert.Equal(t, int32(42), jsonResp.Items[0].Number)
	assert.Equal(t, "Add login feature", jsonResp.Items[0].Title)
}

func TestListProjectPleadings_ProjectNotFoundReturns404(t *testing.T) {
	getByKeyFinder := &fakeGetByKeyFinder{err: errNotFound}
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{}, getByKeyFinder)

	resp, err := handler.ListProjectPleadings(context.Background(), gen.ListProjectPleadingsRequestObject{Key: "missing"})

	require.NoError(t, err)
	_, ok := resp.(gen.ListProjectPleadings404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestFileProjectPleading_NilBodyReturns400(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{}, &fakeGetByKeyFinder{})

	resp, err := handler.FileProjectPleading(context.Background(), gen.FileProjectPleadingRequestObject{
		Key:  "core",
		Body: nil,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.FileProjectPleading400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestFileProjectPleading_ProjectNotFoundReturns404(t *testing.T) {
	getByKeyFinder := &fakeGetByKeyFinder{err: errNotFound}
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{}, getByKeyFinder)

	body := gen.FilePleadingRequest{
		Number:       42,
		Title:        "Add login",
		Petitioner:   "dev@example.com",
		SourceBranch: "feat/login",
		TargetBranch: "main",
		CommitSha:    "abc123",
	}
	resp, err := handler.FileProjectPleading(context.Background(), gen.FileProjectPleadingRequestObject{
		Key:  "missing",
		Body: &body,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.FileProjectPleading404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestListPleadings_WithOptionalParams(t *testing.T) {
	summary := testPleadingSummary()
	listFinder := &fakeListFinder{items: []pleadinglist.PleadingSummary{summary}, total: 1}
	handler := newTestHandler(listFinder, &fakeGetFinder{}, &fakeGetByKeyFinder{})

	projectUUID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	gavelspace := "gavel"
	status := gen.ListPleadingsParamsStatus("open")
	resp, err := handler.ListPleadings(context.Background(), gen.ListPleadingsRequestObject{
		Params: gen.ListPleadingsParams{
			ProjectId:  &projectUUID,
			Gavelspace: &gavelspace,
			Status:     &status,
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListPleadings200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
}

func TestGetPleading_FinderErrorReturnsError(t *testing.T) {
	getFinder := &fakeGetFinder{err: errFake}
	handler := newTestHandler(&fakeListFinder{}, getFinder, &fakeGetByKeyFinder{})

	resp, err := handler.GetPleading(context.Background(), gen.GetPleadingRequestObject{
		Id: httpx.ParseUUIDOrZero("44444444-4444-4444-4444-444444444444"),
	})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestResolvePleading_SuccessReturns204(t *testing.T) {
	getByKeyFinder := &fakeGetByKeyFinder{detail: testProjectDetail()}
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{}, getByKeyFinder)

	body := gen.FilePleadingRequest{
		Number: 42, Title: "PR Title", Petitioner: "dev@example.com",
		SourceBranch: "feat/test", TargetBranch: "main", CommitSha: "abc123",
	}
	fileResp, err := handler.FileProjectPleading(context.Background(), gen.FileProjectPleadingRequestObject{
		Key: "core", Body: &body,
	})
	require.NoError(t, err)
	created, found := fileResp.(gen.FileProjectPleading201JSONResponse)
	require.True(t, found, "expected 201, got %T", fileResp)

	outcome := gen.ResolvePleadingRequestOutcome("merged")
	resp, err := handler.ResolvePleading(context.Background(), gen.ResolvePleadingRequestObject{
		Id:   created.PleadingId,
		Body: &gen.ResolvePleadingJSONRequestBody{Outcome: outcome},
	})

	require.NoError(t, err)
	_, found = resp.(gen.ResolvePleading204Response)
	assert.True(t, found, "expected 204 response, got %T", resp)
}

func TestResolvePleading_NotFoundReturns404(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{}, &fakeGetByKeyFinder{})

	outcome := gen.ResolvePleadingRequestOutcome("merged")
	resp, err := handler.ResolvePleading(context.Background(), gen.ResolvePleadingRequestObject{
		Id:   httpx.ParseUUIDOrZero("99999999-9999-9999-9999-999999999999"),
		Body: &gen.ResolvePleadingJSONRequestBody{Outcome: outcome},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.ResolvePleading404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestResolvePleading_InvalidOutcomeReturns400(t *testing.T) {
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{}, &fakeGetByKeyFinder{})

	outcome := gen.ResolvePleadingRequestOutcome("invalid-outcome")
	resp, err := handler.ResolvePleading(context.Background(), gen.ResolvePleadingRequestObject{
		Id:   httpx.ParseUUIDOrZero("44444444-4444-4444-4444-444444444444"),
		Body: &gen.ResolvePleadingJSONRequestBody{Outcome: outcome},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.ResolvePleading400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestFileProjectPleading_SuccessReturns201(t *testing.T) {
	getByKeyFinder := &fakeGetByKeyFinder{detail: testProjectDetail()}
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{}, getByKeyFinder)

	body := gen.FilePleadingRequest{
		Number: 1, Title: "First PR", Petitioner: "alice@example.com",
		SourceBranch: "feat/one", TargetBranch: "main", CommitSha: "def456",
	}
	resp, err := handler.FileProjectPleading(context.Background(), gen.FileProjectPleadingRequestObject{
		Key: "core", Body: &body,
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.FileProjectPleading201JSONResponse)
	require.True(t, ok, "expected 201 response, got %T", resp)
	assert.NotEqual(t, httpx.ParseUUIDOrZero("00000000-0000-0000-0000-000000000000"), jsonResp.PleadingId)
}

func TestFileProjectPleading_InvalidCommandReturns400(t *testing.T) {
	getByKeyFinder := &fakeGetByKeyFinder{detail: testProjectDetail()}
	handler := newTestHandler(&fakeListFinder{}, &fakeGetFinder{}, getByKeyFinder)

	body := gen.FilePleadingRequest{
		Number: 0, Title: "", Petitioner: "",
		SourceBranch: "", TargetBranch: "", CommitSha: "",
	}
	resp, err := handler.FileProjectPleading(context.Background(), gen.FileProjectPleadingRequestObject{
		Key: "core", Body: &body,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.FileProjectPleading400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestListProjectPleadings_WithStatusParam(t *testing.T) {
	summary := testPleadingSummary()
	listFinder := &fakeListFinder{items: []pleadinglist.PleadingSummary{summary}, total: 1}
	getByKeyFinder := &fakeGetByKeyFinder{detail: testProjectDetail()}
	handler := newTestHandler(listFinder, &fakeGetFinder{}, getByKeyFinder)

	status := gen.ListProjectPleadingsParamsStatus("open")
	resp, err := handler.ListProjectPleadings(context.Background(), gen.ListProjectPleadingsRequestObject{
		Key:    "core",
		Params: gen.ListProjectPleadingsParams{Status: &status},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListProjectPleadings200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
}

func TestListProjectPleadings_FinderErrorReturnsError(t *testing.T) {
	listFinder := &fakeListFinder{err: errFake}
	getByKeyFinder := &fakeGetByKeyFinder{detail: testProjectDetail()}
	handler := newTestHandler(listFinder, &fakeGetFinder{}, getByKeyFinder)

	resp, err := handler.ListProjectPleadings(context.Background(), gen.ListProjectPleadingsRequestObject{Key: "core"})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestListPleadings_WithGateResult(t *testing.T) {
	summary := testPleadingSummaryWithGate()
	listFinder := &fakeListFinder{items: []pleadinglist.PleadingSummary{summary}, total: 1}
	handler := newTestHandler(listFinder, &fakeGetFinder{}, &fakeGetByKeyFinder{})

	projectUUID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	resp, err := handler.ListPleadings(context.Background(), gen.ListPleadingsRequestObject{
		Params: gen.ListPleadingsParams{ProjectId: &projectUUID},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListPleadings200JSONResponse)
	require.True(t, ok, "expected 200, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
	require.NotNil(t, jsonResp.Items[0].GateResult)
	assert.True(t, jsonResp.Items[0].GateResult.Passed)
	require.NotNil(t, jsonResp.Items[0].GateResult.Conditions)
	assert.Len(t, *jsonResp.Items[0].GateResult.Conditions, 1)
}

func TestGetPleading_WithGateResult(t *testing.T) {
	detail := testPleadingDetailWithGate()
	getFinder := &fakeGetFinder{detail: detail}
	handler := newTestHandler(&fakeListFinder{}, getFinder, &fakeGetByKeyFinder{})

	resp, err := handler.GetPleading(context.Background(), gen.GetPleadingRequestObject{
		Id: httpx.ParseUUIDOrZero("44444444-4444-4444-4444-444444444444"),
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetPleading200JSONResponse)
	require.True(t, ok, "expected 200, got %T", resp)
	require.NotNil(t, jsonResp.GateResult)
	assert.True(t, jsonResp.GateResult.Passed)
	require.NotNil(t, jsonResp.GateResult.Conditions)
	assert.Len(t, *jsonResp.GateResult.Conditions, 1)
}
