package judge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	"github.com/usegavel/gavel/core/application/project/preparebaseline"
	apiclient "github.com/usegavel/gavel/core/userinterface/api/v1/client"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
	casefilememory "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	projectmemory "github.com/usegavel/gavel/core/infrastructure/project/memory"
)

var errHelper = errors.New("helper error")

type failingSource struct {
	commitErr error
	branchErr error
}

func (f *failingSource) CommitSHA(_ context.Context) (string, error) {
	return "", f.commitErr
}

func (f *failingSource) Branch(_ context.Context) (string, error) {
	return "", f.branchErr
}

func (f *failingSource) ChangedLines(_ context.Context, _, _ string) (map[string][]int, error) {
	return nil, nil
}

type issueVerifier struct {
	issues []string
	err    error
}

func (v *issueVerifier) VerifyStructure(_ string) ([]string, error) {
	return v.issues, v.err
}

func TestResolveGitInfo_CommitSHAError(t *testing.T) {
	src := &failingSource{commitErr: errHelper}

	_, _, err := resolveGitInfo(context.Background(), src, "", "")

	assert.ErrorIs(t, err, errHelper)
}

func TestResolveGitInfo_BranchError(t *testing.T) {
	src := &failingSource{branchErr: errHelper}

	_, _, err := resolveGitInfo(context.Background(), src, "sha", "")

	assert.ErrorIs(t, err, errHelper)
}

func TestValidateStructure_NoIssues(t *testing.T) {
	var buf bytes.Buffer
	verifier := &issueVerifier{issues: nil}

	err := validateStructure(&buf, verifier, "/workspace")

	require.NoError(t, err)
	assert.Empty(t, buf.String())
}

func TestValidateStructure_WithIssues(t *testing.T) {
	var buf bytes.Buffer
	verifier := &issueVerifier{issues: []string{"missing .bazelrc", "bad MODULE.bazel"}}

	err := validateStructure(&buf, verifier, "/workspace")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gavel structure invalid")
	assert.Contains(t, buf.String(), "missing .bazelrc")
	assert.Contains(t, buf.String(), "bad MODULE.bazel")
}

func TestValidateStructure_VerifierError(t *testing.T) {
	var buf bytes.Buffer
	verifier := &issueVerifier{err: errHelper}

	err := validateStructure(&buf, verifier, "/workspace")

	assert.ErrorIs(t, err, errHelper)
}

func TestResolveConfigPath_DefaultsToGavelDirWhenNothingExists(t *testing.T) {
	tmpDir := t.TempDir()

	result := resolveConfigPath("", tmpDir)

	assert.Contains(t, result, ".gavel")
	assert.Contains(t, result, "gavel.yaml")
}

func TestHasFailedVerdict_NoResults(t *testing.T) {
	assert.False(t, hasFailedVerdict(nil))
}

func TestHasFailedVerdict_AllPass(t *testing.T) {
	results := []pipeline.Result{
		{Verdict: "pass"},
		{Verdict: "pass"},
	}
	assert.False(t, hasFailedVerdict(results))
}

func TestHasFailedVerdict_OneFails(t *testing.T) {
	results := []pipeline.Result{
		{Verdict: "pass"},
		{Verdict: "fail"},
	}
	assert.True(t, hasFailedVerdict(results))
}

func TestPrintBaselineStatus_WithPreviousBaseline(t *testing.T) {
	var buf bytes.Buffer
	result := preparebaseline.Result{
		Baselines: []preparebaseline.ProjectBaseline{
			{ProjectName: "core", FingerprintCount: 42, ArchIDCount: 3, HasPrevious: true, Source: "local"},
		},
	}

	err := printBaselineStatus(&buf, result)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "core")
	assert.Contains(t, buf.String(), "baseline")
}

func TestPrintBaselineStatus_NoPreviousBaseline(t *testing.T) {
	var buf bytes.Buffer
	result := preparebaseline.Result{
		Baselines: []preparebaseline.ProjectBaseline{
			{ProjectName: "core", HasPrevious: false},
		},
	}

	err := printBaselineStatus(&buf, result)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "core")
	assert.Contains(t, buf.String(), "no previous baseline")
}

func TestPrintBaselineStatus_EmptyBaselines(t *testing.T) {
	var buf bytes.Buffer
	result := preparebaseline.Result{Baselines: nil}

	err := printBaselineStatus(&buf, result)

	require.NoError(t, err)
	assert.Empty(t, buf.String())
}

type failWriter struct{}

func (f failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestValidateStructure_WriteError(t *testing.T) {
	verifier := &issueVerifier{issues: []string{"missing file"}}

	err := validateStructure(failWriter{}, verifier, "/workspace")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestPrintBaselineStatus_WriteErrorWithPrevious(t *testing.T) {
	result := preparebaseline.Result{
		Baselines: []preparebaseline.ProjectBaseline{
			{ProjectName: "core", HasPrevious: true, Source: "local", FingerprintCount: 5, ArchIDCount: 2},
		},
	}

	err := printBaselineStatus(failWriter{}, result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestPrintBaselineStatus_WriteErrorNoPrevious(t *testing.T) {
	result := preparebaseline.Result{
		Baselines: []preparebaseline.ProjectBaseline{
			{ProjectName: "core", HasPrevious: false},
		},
	}

	err := printBaselineStatus(failWriter{}, result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func helperDeps() deps {
	return deps{
		projectRepo: projectmemory.NewProjectRepository(),
		fpSeeder:    casefilememory.NewCaseFileRepository(),
		log:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestPrepareBaselines_EmptyProjectsReturnsEmpty(t *testing.T) {
	dependencies := helperDeps()

	result := prepareBaselines(context.Background(), dependencies, nil)

	assert.Empty(t, result.Baselines)
}

func TestPrepareBaselines_ReturnsBaselines(t *testing.T) {
	dependencies := helperDeps()
	projects := []loadgavelspace.ProjectView{
		{Name: "core", DefaultBranch: "main"},
	}

	result := prepareBaselines(context.Background(), dependencies, projects)

	require.Len(t, result.Baselines, 1)
	assert.Equal(t, "core", result.Baselines[0].ProjectName)
	assert.False(t, result.Baselines[0].HasPrevious)
}

func TestPrepareBaselines_WithServerClient(t *testing.T) {
	baselineCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			baselineCalled = true
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"has_previous":     true,
				"fingerprints":     []string{"fp1", "fp2"},
				"architecture_ids": []string{},
			})
			return
		}
		writer.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client, err := apiclient.New(srv.URL, "token")
	require.NoError(t, err)

	dependencies := helperDeps()
	dependencies.serverClient = client

	projects := []loadgavelspace.ProjectView{
		{Name: "core", DefaultBranch: "main"},
	}

	result := prepareBaselines(context.Background(), dependencies, projects)

	require.Len(t, result.Baselines, 1)
	assert.True(t, baselineCalled)
}

func TestClientBaselineFetcher_FetchBaseline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"has_previous":     true,
			"fingerprints":     []string{"fp1", "fp2"},
			"architecture_ids": []string{"rule:a:b"},
		})
	}))
	t.Cleanup(srv.Close)

	client, err := apiclient.New(srv.URL, "token")
	require.NoError(t, err)

	fetcher := &clientBaselineFetcher{client: client}
	baseline, err := fetcher.FetchBaseline(context.Background(), "core", "main")

	require.NoError(t, err)
	require.NotNil(t, baseline)
	assert.True(t, baseline.HasPrevious)
	assert.Equal(t, []string{"fp1", "fp2"}, baseline.Fingerprints)
	assert.Equal(t, []string{"rule:a:b"}, baseline.ArchViolationIDs)
}

func TestClientBaselineFetcher_FetchBaselineError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	client, err := apiclient.New(srv.URL, "token")
	require.NoError(t, err)

	fetcher := &clientBaselineFetcher{client: client}
	_, err = fetcher.FetchBaseline(context.Background(), "core", "main")

	require.Error(t, err)
}
