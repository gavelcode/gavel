package source_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/supporting/source"
)

const testProjectID = "11111111-1111-1111-1111-111111111111"
const testCasefileID = "22222222-2222-2222-2222-222222222222"

func TestGetProjectSourceWithCasefileReturnsCoverageAndFindings(t *testing.T) {
	projectID := testProjectID
	casefileID := testCasefileID
	blobs := &fakeBlobs{content: []byte("package main\n"), ct: "text/plain"}
	coverage := &fakeCoverageFetcher{result: &evidencedto.FileCoverage{
		FilePath:  "main.go",
		Covered:   []int{1},
		Uncovered: []int{2},
	}}
	findings := &fakeFindingFetcher{items: []findinglist.FindingView{
		{Tool: "golint", RuleID: "R1", Severity: "warning", FilePath: "main.go", Line: 2, Message: "bad", FingerprintID: "fp1", Status: "new", Source: "golint", CommitSHA: "abc", ProjectKey: "core", CaseFileID: casefileID},
	}}
	handler := source.New(source.Deps{
		Blobs:               blobs,
		ResolveProjectByKey: fakeProjectResolver(projectID),
		FileCoverage:        coverage,
		Findings:            findings,
	})

	cfUUID := httpx.ParseUUIDOrZero(casefileID)
	resp, err := handler.GetProjectSource(authContext(), gen.GetProjectSourceRequestObject{
		Key: "core",
		Params: gen.GetProjectSourceParams{
			Commit:   "abc123",
			Path:     "main.go",
			Casefile: &cfUUID,
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetProjectSource200JSONResponse)
	require.True(t, ok, "expected JSON response, got %T", resp)
	assert.Equal(t, "package main\n", jsonResp.Content)
	require.NotNil(t, jsonResp.Coverage)
	assert.Equal(t, []int32{1}, jsonResp.Coverage.CoveredLines)
	assert.Equal(t, []int32{2}, jsonResp.Coverage.UncoveredLines)
	require.NotNil(t, jsonResp.Findings)
	require.Len(t, *jsonResp.Findings, 1)
	assert.Equal(t, "golint", (*jsonResp.Findings)[0].Tool)
}

func TestGetProjectSourceWithoutCasefileReturnsBinary(t *testing.T) {
	projectID := testProjectID
	blobs := &fakeBlobs{content: []byte("package main\n"), ct: "text/plain"}
	handler := source.New(source.Deps{
		Blobs:               blobs,
		ResolveProjectByKey: fakeProjectResolver(projectID),
	})

	resp, err := handler.GetProjectSource(authContext(), gen.GetProjectSourceRequestObject{
		Key: "core",
		Params: gen.GetProjectSourceParams{
			Commit: "abc123",
			Path:   "main.go",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.GetProjectSource200AsteriskResponse)
	assert.True(t, ok, "expected binary response, got %T", resp)
}

func TestGetProjectSourceWithCasefileNoCoverageReturnsNilCoverage(t *testing.T) {
	projectID := testProjectID
	casefileID := testCasefileID
	blobs := &fakeBlobs{content: []byte("package main\n"), ct: "text/plain"}
	coverage := &fakeCoverageFetcher{result: nil}
	findings := &fakeFindingFetcher{items: nil}
	handler := source.New(source.Deps{
		Blobs:               blobs,
		ResolveProjectByKey: fakeProjectResolver(projectID),
		FileCoverage:        coverage,
		Findings:            findings,
	})

	cfUUID := httpx.ParseUUIDOrZero(casefileID)
	resp, err := handler.GetProjectSource(authContext(), gen.GetProjectSourceRequestObject{
		Key: "core",
		Params: gen.GetProjectSourceParams{
			Commit:   "abc123",
			Path:     "main.go",
			Casefile: &cfUUID,
		},
	})

	require.NoError(t, err)
	jsonResp, ok := resp.(gen.GetProjectSource200JSONResponse)
	require.True(t, ok)
	assert.Nil(t, jsonResp.Coverage)
	assert.Nil(t, jsonResp.Findings)
}

func TestUploadProjectSourceNilBodyReturns400(t *testing.T) {
	handler := source.New(source.Deps{
		Blobs:               &fakeBlobs{},
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
	})

	resp, err := handler.UploadProjectSource(authContext(), gen.UploadProjectSourceRequestObject{
		Key:  "core",
		Body: nil,
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UploadProjectSource400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestUploadProjectSourceEmptyCommitReturns400(t *testing.T) {
	handler := source.New(source.Deps{
		Blobs:               &fakeBlobs{},
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
	})

	resp, err := handler.UploadProjectSource(authContext(), gen.UploadProjectSourceRequestObject{
		Key: "core",
		Body: &gen.UploadProjectSourceJSONRequestBody{
			Commit: "   ",
			Files:  []gen.SourceFile{{Path: "main.go", Content: "pkg"}},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UploadProjectSource400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestUploadProjectSourceEmptyFilesReturns204(t *testing.T) {
	handler := source.New(source.Deps{
		Blobs:               &fakeBlobs{},
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
	})

	resp, err := handler.UploadProjectSource(authContext(), gen.UploadProjectSourceRequestObject{
		Key: "core",
		Body: &gen.UploadProjectSourceJSONRequestBody{
			Commit: "abc123",
			Files:  []gen.SourceFile{},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UploadProjectSource204Response)
	assert.True(t, ok, "expected 204 response, got %T", resp)
}

func TestUploadProjectSourceSavesBlobs(t *testing.T) {
	blobs := &trackingBlobs{}
	handler := source.New(source.Deps{
		Blobs:               blobs,
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
	})

	resp, err := handler.UploadProjectSource(authContext(), gen.UploadProjectSourceRequestObject{
		Key: "core",
		Body: &gen.UploadProjectSourceJSONRequestBody{
			Commit: "abc123",
			Files: []gen.SourceFile{
				{Path: "main.go", Content: "package main"},
				{Path: "util.go", Content: "package util"},
			},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UploadProjectSource204Response)
	require.True(t, ok, "expected 204 response, got %T", resp)
	require.Len(t, blobs.saved, 2)
	assert.Equal(t, testProjectID, blobs.saved[0].projectID)
	assert.Equal(t, "abc123", blobs.saved[0].commitSHA)
	assert.Equal(t, "main.go", blobs.saved[0].path)
	assert.Equal(t, []byte("package main"), blobs.saved[0].content)
	assert.Equal(t, "text/plain; charset=utf-8", blobs.saved[0].contentType)
	assert.Equal(t, "util.go", blobs.saved[1].path)
	assert.Equal(t, []byte("package util"), blobs.saved[1].content)
}

func TestUploadProjectSourceSkipsUnsafePaths(t *testing.T) {
	blobs := &trackingBlobs{}
	handler := source.New(source.Deps{
		Blobs:               blobs,
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
	})

	resp, err := handler.UploadProjectSource(authContext(), gen.UploadProjectSourceRequestObject{
		Key: "core",
		Body: &gen.UploadProjectSourceJSONRequestBody{
			Commit: "abc123",
			Files: []gen.SourceFile{
				{Path: "../etc/passwd", Content: "root:x:0:0"},
				{Path: "safe.go", Content: "package safe"},
			},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UploadProjectSource204Response)
	require.True(t, ok, "expected 204 response, got %T", resp)
	require.Len(t, blobs.saved, 1)
	assert.Equal(t, "safe.go", blobs.saved[0].path)
}

func TestUploadProjectSourceProjectNotFoundReturns404(t *testing.T) {
	handler := source.New(source.Deps{
		Blobs:               &fakeBlobs{},
		ResolveProjectByKey: notFoundProjectResolver(),
	})

	resp, err := handler.UploadProjectSource(authContext(), gen.UploadProjectSourceRequestObject{
		Key: "nonexistent",
		Body: &gen.UploadProjectSourceJSONRequestBody{
			Commit: "abc123",
			Files:  []gen.SourceFile{{Path: "main.go", Content: "pkg"}},
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.UploadProjectSource404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestGetProjectSourceMissingParamsReturns400(t *testing.T) {
	handler := source.New(source.Deps{
		Blobs:               &fakeBlobs{},
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
	})

	resp, err := handler.GetProjectSource(authContext(), gen.GetProjectSourceRequestObject{
		Key: "core",
		Params: gen.GetProjectSourceParams{
			Commit: "",
			Path:   "",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.GetProjectSource400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestGetProjectSourceUnsafePathReturns400(t *testing.T) {
	handler := source.New(source.Deps{
		Blobs:               &fakeBlobs{},
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
	})

	resp, err := handler.GetProjectSource(authContext(), gen.GetProjectSourceRequestObject{
		Key: "core",
		Params: gen.GetProjectSourceParams{
			Commit: "abc",
			Path:   "../secret",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.GetProjectSource400JSONResponse)
	assert.True(t, ok, "expected 400 response, got %T", resp)
}

func TestGetProjectSourceProjectNotFoundReturns404(t *testing.T) {
	handler := source.New(source.Deps{
		Blobs:               &fakeBlobs{},
		ResolveProjectByKey: notFoundProjectResolver(),
	})

	resp, err := handler.GetProjectSource(authContext(), gen.GetProjectSourceRequestObject{
		Key: "nonexistent",
		Params: gen.GetProjectSourceParams{
			Commit: "abc123",
			Path:   "main.go",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.GetProjectSource404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestGetProjectSourceBlobNotFoundReturns404(t *testing.T) {
	handler := source.New(source.Deps{
		Blobs:               &notFoundBlobs{},
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
	})

	resp, err := handler.GetProjectSource(authContext(), gen.GetProjectSourceRequestObject{
		Key: "core",
		Params: gen.GetProjectSourceParams{
			Commit: "abc123",
			Path:   "missing.go",
		},
	})

	require.NoError(t, err)
	_, ok := resp.(gen.GetProjectSource404JSONResponse)
	assert.True(t, ok, "expected 404 response, got %T", resp)
}

func TestGetProjectSourceCoverageFetchErrorReturnsError(t *testing.T) {
	casefileID := testCasefileID
	blobs := &fakeBlobs{content: []byte("package main\n"), ct: "text/plain"}
	coverageErr := errors.New("coverage database unavailable")
	handler := source.New(source.Deps{
		Blobs:               blobs,
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
		FileCoverage:        &fakeCoverageFetcher{err: coverageErr},
		Findings:            &fakeFindingFetcher{},
	})

	cfUUID := httpx.ParseUUIDOrZero(casefileID)
	resp, err := handler.GetProjectSource(authContext(), gen.GetProjectSourceRequestObject{
		Key: "core",
		Params: gen.GetProjectSourceParams{
			Commit:   "abc123",
			Path:     "main.go",
			Casefile: &cfUUID,
		},
	})

	assert.Nil(t, resp)
	require.ErrorIs(t, err, coverageErr)
}

func TestGetProjectSourceFindingsFetchErrorReturnsError(t *testing.T) {
	casefileID := testCasefileID
	blobs := &fakeBlobs{content: []byte("package main\n"), ct: "text/plain"}
	findingsErr := errors.New("findings database unavailable")
	handler := source.New(source.Deps{
		Blobs:               blobs,
		ResolveProjectByKey: fakeProjectResolver(testProjectID),
		FileCoverage:        &fakeCoverageFetcher{},
		Findings:            &fakeFindingFetcher{err: findingsErr},
	})

	cfUUID := httpx.ParseUUIDOrZero(casefileID)
	resp, err := handler.GetProjectSource(authContext(), gen.GetProjectSourceRequestObject{
		Key: "core",
		Params: gen.GetProjectSourceParams{
			Commit:   "abc123",
			Path:     "main.go",
			Casefile: &cfUUID,
		},
	})

	assert.Nil(t, resp)
	require.ErrorIs(t, err, findingsErr)
}
