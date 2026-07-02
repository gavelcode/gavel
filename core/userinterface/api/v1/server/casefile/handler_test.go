package casefile_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	casefileget "github.com/usegavel/gavel/core/application/casefile/get"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	"github.com/usegavel/gavel/core/application/casefile/judge"
	casefilelist "github.com/usegavel/gavel/core/application/casefile/list"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	"github.com/usegavel/gavel/core/application/project/getbaseline"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	"github.com/usegavel/gavel/core/application/project/projectview"
	"github.com/usegavel/gavel/core/domain/project/model"
	memcf "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/casefile"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func newQueryHandler(
	listFinder *fakeListFinder,
	getFinder *fakeGetFinder,
	findingFinder *fakeFindingFinder,
	getByKeyFinder *fakeGetByKeyFinder,
) *casefile.Handler {
	return casefile.New(casefile.Deps{
		ListCaseFiles:       casefilelist.NewHandler(listFinder),
		GetCaseFile:         casefileget.NewHandler(getFinder),
		ListFindings:        findinglist.NewHandler(findingFinder),
		ResolveProjectByKey: projectgetbykey.NewHandler(getByKeyFinder),
	})
}

func newMinimalHandler() *casefile.Handler {
	return casefile.New(casefile.Deps{
		Now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})
}

func TestListCaseFiles_MissingProjectAndGavelspace(t *testing.T) {
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{},
	)

	resp, err := handler.ListCaseFiles(context.Background(), gen.ListCaseFilesRequestObject{})

	require.NoError(t, err)
	_, ok := resp.(gen.ListCaseFiles400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestListCaseFiles_WithProjectID(t *testing.T) {
	summary := testCaseFileSummary()
	handler := newQueryHandler(
		&fakeListFinder{items: []casefilelist.CaseFileSummary{summary}, total: 1},
		&fakeGetFinder{},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{},
	)

	projectUUID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	resp, err := handler.ListCaseFiles(context.Background(), gen.ListCaseFilesRequestObject{
		Params: gen.ListCaseFilesParams{ProjectId: &projectUUID},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListCaseFiles200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
	assert.Equal(t, "abc123", jsonResp.Items[0].CommitSha)
	assert.Equal(t, "main", jsonResp.Items[0].Branch)
	assert.Equal(t, "pass", jsonResp.Items[0].VerdictOutcome)
	assert.Equal(t, int32(10), jsonResp.Items[0].TotalFindings)
	assert.Equal(t, int32(2), jsonResp.Items[0].NewFindings)
	assert.Equal(t, int32(8), jsonResp.Items[0].ExistingFindings)
	assert.Equal(t, int32(3), jsonResp.Items[0].ResolvedFindings)
	require.NotNil(t, jsonResp.Items[0].CoveragePercent)
	assert.InDelta(t, 85.5, *jsonResp.Items[0].CoveragePercent, 0.01)
}

func TestListCaseFiles_EmptyList(t *testing.T) {
	handler := newQueryHandler(
		&fakeListFinder{items: nil, total: 0},
		&fakeGetFinder{},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{},
	)

	projectUUID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	resp, err := handler.ListCaseFiles(context.Background(), gen.ListCaseFilesRequestObject{
		Params: gen.ListCaseFilesParams{ProjectId: &projectUUID},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListCaseFiles200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	assert.Empty(t, jsonResp.Items)
	assert.Nil(t, jsonResp.NextCursor)
}

func TestListCaseFiles_FinderError(t *testing.T) {
	handler := newQueryHandler(
		&fakeListFinder{err: errFake},
		&fakeGetFinder{},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{},
	)

	projectUUID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	resp, err := handler.ListCaseFiles(context.Background(), gen.ListCaseFilesRequestObject{
		Params: gen.ListCaseFilesParams{ProjectId: &projectUUID},
	})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetCaseFile_ReturnsDetail(t *testing.T) {
	detail := testCaseFileDetail()
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{detail: detail},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{},
	)

	resp, err := handler.GetCaseFile(context.Background(), gen.GetCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222"),
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetCaseFile200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	assert.Equal(t, "abc123", jsonResp.CommitSha)
	assert.Equal(t, "main", jsonResp.Branch)
	assert.Equal(t, "pass", jsonResp.VerdictOutcome)
	assert.Equal(t, int32(10), jsonResp.TotalFindings)
	assert.Equal(t, int32(2), jsonResp.NewFindings)
	assert.Equal(t, int32(8), jsonResp.ExistingFindings)
	assert.Equal(t, int32(3), jsonResp.ResolvedFindings)
	require.NotNil(t, jsonResp.CoveragePercent)
	assert.InDelta(t, 85.5, *jsonResp.CoveragePercent, 0.01)
	require.Len(t, jsonResp.Evidences, 1)
	assert.Equal(t, "code_quality", jsonResp.Evidences[0].Subtype)
	assert.Equal(t, "golangci-lint", jsonResp.Evidences[0].Source)
	require.Len(t, jsonResp.Rulings, 1)
	assert.Equal(t, "code_quality", jsonResp.Rulings[0].Subtype)
	assert.True(t, jsonResp.Rulings[0].Passed)
	assert.Equal(t, "0 findings", jsonResp.Rulings[0].Detail)
}

func TestGetCaseFile_NotFoundReturns404(t *testing.T) {
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{err: errNotFound},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{},
	)

	resp, err := handler.GetCaseFile(context.Background(), gen.GetCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222"),
	})

	require.NoError(t, err)
	_, ok := resp.(gen.GetCaseFile404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestListFindings_ReturnsItems(t *testing.T) {
	finding := testFindingView()
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{},
		&fakeFindingFinder{items: []findinglist.FindingView{finding}, total: 1},
		&fakeGetByKeyFinder{},
	)

	cfID := httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222")
	resp, err := handler.ListFindings(context.Background(), gen.ListFindingsRequestObject{
		Params: gen.ListFindingsParams{CasefileId: &cfID},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListFindings200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
	assert.Equal(t, "golangci-lint", jsonResp.Items[0].Tool)
	assert.Equal(t, "errcheck", jsonResp.Items[0].RuleId)
	assert.Equal(t, "warning", jsonResp.Items[0].Severity)
	assert.Equal(t, "main.go", jsonResp.Items[0].FilePath)
	assert.Equal(t, int32(10), jsonResp.Items[0].Line)
	assert.Equal(t, "unchecked error", jsonResp.Items[0].Message)
	assert.Equal(t, "fp1", jsonResp.Items[0].Fingerprint)
	assert.Equal(t, "new", jsonResp.Items[0].Status)
	assert.Equal(t, "core", jsonResp.Items[0].ProjectKey)
}

func TestListFindings_EmptyResults(t *testing.T) {
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{},
		&fakeFindingFinder{items: nil, total: 0},
		&fakeGetByKeyFinder{},
	)

	cfID := httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222")
	resp, err := handler.ListFindings(context.Background(), gen.ListFindingsRequestObject{
		Params: gen.ListFindingsParams{CasefileId: &cfID},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListFindings200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	assert.Empty(t, jsonResp.Items)
	assert.Nil(t, jsonResp.NextCursor)
}

func TestListProjectCaseFiles_ReturnsItems(t *testing.T) {
	summary := testCaseFileSummary()
	handler := newQueryHandler(
		&fakeListFinder{items: []casefilelist.CaseFileSummary{summary}, total: 1},
		&fakeGetFinder{},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{detail: &projectview.ProjectDetail{
			ID:  "11111111-1111-1111-1111-111111111111",
			Key: "core",
		}},
	)

	resp, err := handler.ListProjectCaseFiles(context.Background(), gen.ListProjectCaseFilesRequestObject{
		Key: "core",
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListProjectCaseFiles200JSONResponse)
	require.True(t, ok, "expected 200 JSON response, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
	assert.Equal(t, "abc123", jsonResp.Items[0].CommitSha)
}

func TestListProjectCaseFiles_ProjectNotFoundReturns404(t *testing.T) {
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{err: errNotFound},
	)

	resp, err := handler.ListProjectCaseFiles(context.Background(), gen.ListProjectCaseFilesRequestObject{
		Key: "missing",
	})

	require.NoError(t, err)
	_, ok := resp.(gen.ListProjectCaseFiles404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestCreateCaseFile_NilBodyReturns400(t *testing.T) {
	handler := newMinimalHandler()

	resp, err := handler.CreateCaseFile(context.Background(), gen.CreateCaseFileRequestObject{
		Body: nil,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateCaseFile400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestIngestCaseFileEvidence_NilBodyReturns400(t *testing.T) {
	handler := newMinimalHandler()

	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id:   httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222"),
		Body: nil,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.IngestCaseFileEvidence400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestFinalizeCaseFile_NilBodyReturns400(t *testing.T) {
	handler := newMinimalHandler()

	resp, err := handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id:   httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222"),
		Body: nil,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.FinalizeCaseFile400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func newMutationHandler(t *testing.T) (*casefile.Handler, *memcf.CaseFileRepository, *memproject.ProjectRepository) {
	t.Helper()
	cfRepo := memcf.NewCaseFileRepository()
	projRepo := memproject.NewProjectRepository()

	createH := createcasefile.NewHandler(cfRepo, projRepo)
	ingestH := ingestevidence.NewHandler(cfRepo)
	classifyH := classify.NewHandler(cfRepo)
	judgeH := judge.NewHandler(cfRepo, projRepo)
	finalizeH := finalize.NewHandler(cfRepo, projRepo, classifyH, judgeH, nil)

	handler := casefile.New(casefile.Deps{
		CreateCaseFile:   createH,
		IngestEvidence:   ingestH,
		FinalizeCaseFile: finalizeH,
		Now:              func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})
	return handler, cfRepo, projRepo
}

func seedProject(t *testing.T, repo *memproject.ProjectRepository) model.Project {
	t.Helper()
	p, err := model.NewProject("core", "Core", "//core/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), p))
	return p
}

func TestCreateCaseFile_Success(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)

	resp, err := handler.CreateCaseFile(context.Background(), gen.CreateCaseFileRequestObject{
		Body: &gen.CreateCaseFileRequest{
			ProjectId: httpx.ParseUUIDOrZero(p.ID().String()),
			CommitSha: "abc123",
			Branch:    "main",
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.CreateCaseFile201JSONResponse)
	require.True(t, ok, "expected 201 response, got %T", resp)
	assert.NotEqual(t, httpx.ParseUUIDOrZero("00000000-0000-0000-0000-000000000000"), jsonResp.CaseFileId)
}

func TestCreateCaseFile_ProjectNotFoundReturns404(t *testing.T) {
	handler, _, _ := newMutationHandler(t)

	resp, err := handler.CreateCaseFile(context.Background(), gen.CreateCaseFileRequestObject{
		Body: &gen.CreateCaseFileRequest{
			ProjectId: httpx.ParseUUIDOrZero("99999999-9999-9999-9999-999999999999"),
			CommitSha: "abc123",
			Branch:    "main",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateCaseFile404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestCreateCaseFile_InvalidBodyReturns400(t *testing.T) {
	handler, _, _ := newMutationHandler(t)

	resp, err := handler.CreateCaseFile(context.Background(), gen.CreateCaseFileRequestObject{
		Body: &gen.CreateCaseFileRequest{
			ProjectId: httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111"),
			CommitSha: "",
			Branch:    "main",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateCaseFile400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestCreateCaseFile_WithStartedAtAndFreshEvaluation(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	project := seedProject(t, projRepo)

	started := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	fresh := true
	resp, err := handler.CreateCaseFile(context.Background(), gen.CreateCaseFileRequestObject{
		Body: &gen.CreateCaseFileRequest{
			ProjectId:       httpx.ParseUUIDOrZero(project.ID().String()),
			CommitSha:       "def456",
			Branch:          "feature",
			StartedAt:       &started,
			FreshEvaluation: &fresh,
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.CreateCaseFile201JSONResponse)
	assert.True(t, ok, "expected 201 response, got %T", resp)
}

func createCaseFileViaHandler(t *testing.T, handler *casefile.Handler, projectID string) string {
	t.Helper()
	resp, err := handler.CreateCaseFile(context.Background(), gen.CreateCaseFileRequestObject{
		Body: &gen.CreateCaseFileRequest{
			ProjectId: httpx.ParseUUIDOrZero(projectID),
			CommitSha: "abc123",
			Branch:    "main",
		},
	})
	require.NoError(t, err)
	created, ok := resp.(gen.CreateCaseFile201JSONResponse)
	require.True(t, ok, "expected 201, got %T", resp)
	return created.CaseFileId.String()
}

func TestIngestCaseFileEvidence_FindingsSuccess(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	findings := []gen.IngestFinding{
		{
			Tool:        "golangci-lint",
			RuleId:      "errcheck",
			Severity:    "warning",
			FilePath:    "main.go",
			Line:        42,
			Message:     "unchecked error",
			Fingerprint: "fp-001",
		},
	}
	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.IngestEvidenceRequest{
			Subtype:     gen.CodeQuality,
			Source:      "golangci-lint",
			CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Findings:    &findings,
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.IngestCaseFileEvidence201JSONResponse)
	require.True(t, ok, "expected 201 response, got %T", resp)
	assert.NotEqual(t, httpx.ParseUUIDOrZero("00000000-0000-0000-0000-000000000000"), jsonResp.EvidenceId)
}

func TestIngestCaseFileEvidence_CoverageWithByFileAndByLanguage(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	byLanguage := []gen.IngestLanguageStats{
		{Language: "go", TotalLines: 1000, CoveredLines: 800},
	}
	byFile := []gen.IngestFileCoverage{
		{FilePath: "main.go", CoveredLines: []int32{1, 2, 3}, UncoveredLines: []int32{4, 5}},
	}
	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.IngestEvidenceRequest{
			Subtype:     gen.Coverage,
			Source:      "lcov",
			CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Coverage: &gen.IngestCoverage{
				TotalLines:   1000,
				CoveredLines: 800,
				ByLanguage:   &byLanguage,
				ByFile:       &byFile,
			},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.IngestCaseFileEvidence201JSONResponse)
	assert.True(t, ok, "expected 201 response, got %T", resp)
}

func TestIngestCaseFileEvidence_ArchitectureViolations(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.IngestEvidenceRequest{
			Subtype:     gen.Architecture,
			Source:      "archtest",
			CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Architecture: &gen.IngestArchitecture{
				Violations: []gen.IngestViolation{
					{Rule: "domain-imports", SourcePkg: "domain/casefile", TargetPkg: "infrastructure/db", Message: "domain must not import infrastructure"},
				},
			},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.IngestCaseFileEvidence201JSONResponse)
	assert.True(t, ok, "expected 201 response, got %T", resp)
}

func TestIngestCaseFileEvidence_LicenseDependencies(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.IngestEvidenceRequest{
			Subtype:     gen.License,
			Source:      "license-checker",
			CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			License: &gen.IngestLicense{
				Dependencies: []gen.IngestDependency{
					{Name: "testify", Version: "1.9.0", License: "MIT"},
					{Name: "chi", Version: "5.0.0", License: "MIT"},
				},
			},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.IngestCaseFileEvidence201JSONResponse)
	assert.True(t, ok, "expected 201 response, got %T", resp)
}

func TestIngestCaseFileEvidence_NewCodeCoverage(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.IngestEvidenceRequest{
			Subtype:     gen.NewCodeCoverage,
			Source:      "diff-cover",
			CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			NewCodeCoverage: &gen.IngestNewCodeCoverage{
				CoveredLines:   45,
				CoverableLines: 50,
			},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.IngestCaseFileEvidence201JSONResponse)
	assert.True(t, ok, "expected 201 response, got %T", resp)
}

func TestIngestCaseFileEvidence_EmptyContentReturns400(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.IngestEvidenceRequest{
			Subtype:     gen.CodeQuality,
			Source:      "golangci-lint",
			CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.IngestCaseFileEvidence400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestIngestCaseFileEvidence_CaseFileNotFoundReturns404(t *testing.T) {
	handler, _, _ := newMutationHandler(t)

	findings := []gen.IngestFinding{
		{Tool: "lint", RuleId: "r1", Severity: "error", FilePath: "a.go", Line: 1, Message: "msg", Fingerprint: "fp1"},
	}
	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id: httpx.ParseUUIDOrZero("99999999-9999-9999-9999-999999999999"),
		Body: &gen.IngestEvidenceRequest{
			Subtype:     gen.CodeQuality,
			Source:      "lint",
			CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Findings:    &findings,
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.IngestCaseFileEvidence404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestFinalizeCaseFile_PrecomputedVerdictSuccess(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	evaluatedAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	rulings := []gen.Ruling{
		{Subtype: "code_quality", Passed: true, Detail: "0 findings"},
	}
	resp, err := handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.FinalizeCaseFileRequest{
			Verdict: gen.Verdict{
				Outcome:     gen.Pass,
				Rulings:     &rulings,
				EvaluatedAt: evaluatedAt,
			},
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.FinalizeCaseFile200JSONResponse)
	require.True(t, ok, "expected 200 response, got %T", resp)
	assert.Equal(t, httpx.ParseUUIDOrZero(cfID), jsonResp.CaseFileId)
	assert.Equal(t, gen.Pass, jsonResp.Verdict.Outcome)
	require.NotNil(t, jsonResp.Verdict.Rulings)
	require.Len(t, *jsonResp.Verdict.Rulings, 1)
	assert.Equal(t, "code_quality", (*jsonResp.Verdict.Rulings)[0].Subtype)
	assert.True(t, (*jsonResp.Verdict.Rulings)[0].Passed)
}

func TestFinalizeCaseFile_EmptyOutcomeReturns400(t *testing.T) {
	handler, _, _ := newMutationHandler(t)

	resp, err := handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222"),
		Body: &gen.FinalizeCaseFileRequest{
			Verdict: gen.Verdict{Outcome: ""},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.FinalizeCaseFile400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestFinalizeCaseFile_CaseFileNotFoundReturns404(t *testing.T) {
	handler, _, _ := newMutationHandler(t)

	evaluatedAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	resp, err := handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero("99999999-9999-9999-9999-999999999999"),
		Body: &gen.FinalizeCaseFileRequest{
			Verdict: gen.Verdict{
				Outcome:     gen.Pass,
				EvaluatedAt: evaluatedAt,
			},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.FinalizeCaseFile404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestFinalizeCaseFile_AlreadyJudgedReturns409(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	evaluatedAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	body := &gen.FinalizeCaseFileRequest{
		Verdict: gen.Verdict{
			Outcome:     gen.Pass,
			EvaluatedAt: evaluatedAt,
		},
	}

	resp, err := handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID), Body: body,
	})
	require.NoError(t, err)
	_, found := resp.(gen.FinalizeCaseFile200JSONResponse)
	require.True(t, found, "first finalize should succeed, got %T", resp)

	resp, err = handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID), Body: body,
	})
	require.NoError(t, err)
	_, found = resp.(gen.FinalizeCaseFile409JSONResponse)
	assert.True(t, found, "expected 409 response for already judged, got %T", resp)
}

func TestMapFinalizeError_NotFound(t *testing.T) {
	handler, _, _ := newMutationHandler(t)

	evaluatedAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	resp, err := handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero("99999999-9999-9999-9999-999999999999"),
		Body: &gen.FinalizeCaseFileRequest{
			Verdict: gen.Verdict{Outcome: gen.Pass, EvaluatedAt: evaluatedAt},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.FinalizeCaseFile404JSONResponse)
	assert.True(t, ok, "expected 404 from mapFinalizeError, got %T", resp)
}

func TestMapFinalizeError_Validation(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	resp, err := handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.FinalizeCaseFileRequest{
			Verdict: gen.Verdict{
				Outcome:     gen.VerdictOutcome("invalid-outcome"),
				EvaluatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	})

	require.NoError(t, err)
	_, isBadRequest := resp.(gen.FinalizeCaseFile400JSONResponse)
	_, isServerErr := resp.(gen.FinalizeCaseFile200JSONResponse)
	assert.True(t, isBadRequest || !isServerErr, "expected error response for invalid outcome, got %T", resp)
}

func TestIngestCaseFileEvidence_WithClientSuppliedID(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	clientID := "44444444-4444-4444-4444-444444444444"
	findings := []gen.IngestFinding{
		{Tool: "lint", RuleId: "r1", Severity: "error", FilePath: "a.go", Line: 1, Message: "msg", Fingerprint: "fp1"},
	}
	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.IngestEvidenceRequest{
			Subtype:     gen.CodeQuality,
			Source:      "lint",
			CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Id:          &clientID,
			Findings:    &findings,
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.IngestCaseFileEvidence201JSONResponse)
	require.True(t, ok, "expected 201 response, got %T", resp)
	assert.Equal(t, httpx.ParseUUIDOrZero(clientID), jsonResp.EvidenceId)
}

func TestIngestCaseFileEvidence_FileCoverageSaverCalled(t *testing.T) {
	cfRepo := memcf.NewCaseFileRepository()
	projRepo := memproject.NewProjectRepository()

	createH := createcasefile.NewHandler(cfRepo, projRepo)
	ingestH := ingestevidence.NewHandler(cfRepo)
	saver := &fakeFileCoverageSaver{}

	handler := casefile.New(casefile.Deps{
		CreateCaseFile: createH,
		IngestEvidence: ingestH,
		FileCoverage:   saver,
		Now:            func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})

	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	byFile := []gen.IngestFileCoverage{
		{FilePath: "main.go", CoveredLines: []int32{1, 2}, UncoveredLines: []int32{3}},
	}
	resp, err := handler.IngestCaseFileEvidence(context.Background(), gen.IngestCaseFileEvidenceRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.IngestEvidenceRequest{
			Subtype:     gen.Coverage,
			Source:      "lcov",
			CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Coverage: &gen.IngestCoverage{
				TotalLines:   100,
				CoveredLines: 80,
				ByFile:       &byFile,
			},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.IngestCaseFileEvidence201JSONResponse)
	require.True(t, ok, "expected 201 response, got %T", resp)
	assert.Equal(t, cfID, saver.savedCaseFileID)
	require.Len(t, saver.savedEntries, 1)
	assert.Equal(t, "main.go", saver.savedEntries[0].FilePath)
}

func TestFinalizeCaseFile_CountersPopulated(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	evaluatedAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	rulings := []gen.Ruling{
		{Subtype: "code_quality", Passed: true, Detail: "0 findings"},
		{Subtype: "coverage", Passed: true, Detail: "90%"},
	}
	resp, err := handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.FinalizeCaseFileRequest{
			Verdict: gen.Verdict{
				Outcome:     gen.Pass,
				Rulings:     &rulings,
				EvaluatedAt: evaluatedAt,
			},
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.FinalizeCaseFile200JSONResponse)
	require.True(t, ok, "expected 200 response, got %T", resp)
	require.NotNil(t, jsonResp.Verdict.Rulings)
	assert.Len(t, *jsonResp.Verdict.Rulings, 2)
	assert.Equal(t, int32(0), jsonResp.Counters.FindingsCount)
}

func TestFinalizeCaseFile_NoRulingsOmitsRulings(t *testing.T) {
	handler, _, projRepo := newMutationHandler(t)
	p := seedProject(t, projRepo)
	cfID := createCaseFileViaHandler(t, handler, p.ID().String())

	evaluatedAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	resp, err := handler.FinalizeCaseFile(context.Background(), gen.FinalizeCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero(cfID),
		Body: &gen.FinalizeCaseFileRequest{
			Verdict: gen.Verdict{
				Outcome:     gen.Pass,
				EvaluatedAt: evaluatedAt,
			},
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.FinalizeCaseFile200JSONResponse)
	require.True(t, ok, "expected 200 response, got %T", resp)
	assert.Equal(t, gen.Pass, jsonResp.Verdict.Outcome)
}

func TestListCaseFiles_WithGavelspaceParam(t *testing.T) {
	summary := testCaseFileSummary()
	handler := newQueryHandler(
		&fakeListFinder{items: []casefilelist.CaseFileSummary{summary}, total: 1},
		&fakeGetFinder{},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{},
	)

	gavelspace := "gavel"
	resp, err := handler.ListCaseFiles(context.Background(), gen.ListCaseFilesRequestObject{
		Params: gen.ListCaseFilesParams{Gavelspace: &gavelspace},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListCaseFiles200JSONResponse)
	require.True(t, ok, "expected 200, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
}

func TestGetCaseFile_FinderErrorReturnsError(t *testing.T) {
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{err: errFake},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{},
	)

	resp, err := handler.GetCaseFile(context.Background(), gen.GetCaseFileRequestObject{
		Id: httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222"),
	})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestListFindings_WithAllFilterParams(t *testing.T) {
	finding := testFindingView()
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{},
		&fakeFindingFinder{items: []findinglist.FindingView{finding}, total: 1},
		&fakeGetByKeyFinder{},
	)

	projectUUID := httpx.ParseUUIDOrZero("11111111-1111-1111-1111-111111111111")
	cfUUID := httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222")
	tool := "golangci-lint"
	severity := "warning"
	status := "new"
	filePath := "main.go"
	gavelspace := "gavel"
	resp, err := handler.ListFindings(context.Background(), gen.ListFindingsRequestObject{
		Params: gen.ListFindingsParams{
			ProjectId:  &projectUUID,
			CasefileId: &cfUUID,
			Tool:       &tool,
			Severity:   &severity,
			Status:     &status,
			FilePath:   &filePath,
			Gavelspace: &gavelspace,
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.ListFindings200JSONResponse)
	require.True(t, ok, "expected 200, got %T", resp)
	require.Len(t, jsonResp.Items, 1)
}

func TestListFindings_FinderErrorReturnsError(t *testing.T) {
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{},
		&fakeFindingFinder{err: errFake},
		&fakeGetByKeyFinder{},
	)

	cfUUID := httpx.ParseUUIDOrZero("22222222-2222-2222-2222-222222222222")
	resp, err := handler.ListFindings(context.Background(), gen.ListFindingsRequestObject{
		Params: gen.ListFindingsParams{CasefileId: &cfUUID},
	})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestListProjectCaseFiles_FinderErrorReturnsError(t *testing.T) {
	handler := newQueryHandler(
		&fakeListFinder{},
		&fakeGetFinder{},
		&fakeFindingFinder{},
		&fakeGetByKeyFinder{err: errFake},
	)

	resp, err := handler.ListProjectCaseFiles(context.Background(), gen.ListProjectCaseFilesRequestObject{
		Key: "core",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetProjectBaseline_Success(t *testing.T) {
	projRepo := memproject.NewProjectRepository()
	p := seedProject(t, projRepo)
	byKeyFinder := &fakeBaselineFinder{detail: &projectview.ProjectDetail{
		ID:            p.ID().String(),
		Key:           "core",
		DefaultBranch: "main",
	}}
	handler := casefile.New(casefile.Deps{
		GetBaseline:         getbaseline.NewHandler(byKeyFinder, projRepo),
		ResolveProjectByKey: projectgetbykey.NewHandler(byKeyFinder),
		Now:                 func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})

	resp, err := handler.GetProjectBaseline(context.Background(), gen.GetProjectBaselineRequestObject{
		Key: "core",
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetProjectBaseline200JSONResponse)
	require.True(t, ok, "expected 200, got %T", resp)
	assert.NotNil(t, jsonResp.Fingerprints)
	assert.NotNil(t, jsonResp.ArchitectureIds)
}

func TestGetProjectBaseline_WithBranchParam(t *testing.T) {
	projRepo := memproject.NewProjectRepository()
	p := seedProject(t, projRepo)
	byKeyFinder := &fakeBaselineFinder{detail: &projectview.ProjectDetail{
		ID:            p.ID().String(),
		Key:           "core",
		DefaultBranch: "main",
	}}
	handler := casefile.New(casefile.Deps{
		GetBaseline:         getbaseline.NewHandler(byKeyFinder, projRepo),
		ResolveProjectByKey: projectgetbykey.NewHandler(byKeyFinder),
		Now:                 func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})

	branch := "develop"
	resp, err := handler.GetProjectBaseline(context.Background(), gen.GetProjectBaselineRequestObject{
		Key:    "core",
		Params: gen.GetProjectBaselineParams{Branch: &branch},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetProjectBaseline200JSONResponse)
	require.True(t, ok, "expected 200, got %T", resp)
	assert.False(t, jsonResp.HasPrevious)
}

func TestGetProjectBaseline_ProjectNotFoundReturns404(t *testing.T) {
	projRepo := memproject.NewProjectRepository()
	byKeyFinder := &fakeBaselineFinder{err: errNotFound}
	handler := casefile.New(casefile.Deps{
		GetBaseline:         getbaseline.NewHandler(byKeyFinder, projRepo),
		ResolveProjectByKey: projectgetbykey.NewHandler(byKeyFinder),
		Now:                 func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})

	resp, err := handler.GetProjectBaseline(context.Background(), gen.GetProjectBaselineRequestObject{
		Key: "missing",
	})

	require.NoError(t, err)
	_, ok := resp.(gen.GetProjectBaseline404JSONResponse)
	assert.True(t, ok, "expected 404, got %T", resp)
}
