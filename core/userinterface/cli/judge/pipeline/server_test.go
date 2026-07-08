package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	corejudge "github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	projectmemory "github.com/usegavel/gavel/core/infrastructure/project/memory"
	apiclient "github.com/usegavel/gavel/core/userinterface/api/v1/client"
)

const (
	testProjectID  = "550e8400-e29b-41d4-a716-446655440000"
	testCaseFileID = "660e8400-e29b-41d4-a716-446655440000"
	testPleadingID = "770e8400-e29b-41d4-a716-446655440000"
)

func newMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/projects/{key}", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": testProjectID, "key": "backend", "name": "backend",
			"target_pattern": "//backend/...", "default_branch": "main",
			"languages": []string{}, "created_at": "2025-01-01T00:00:00Z",
		})
	})

	mux.HandleFunc("POST /api/v1/projects", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{
			"id": testProjectID, "key": "backend", "name": "backend",
		})
	})

	mux.HandleFunc("POST /api/v1/casefiles", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{
			"case_file_id": testCaseFileID,
		})
	})

	mux.HandleFunc("POST /api/v1/casefiles/{id}/evidence", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{
			"evidence_id": "880e8400-e29b-41d4-a716-446655440000",
		})
	})

	mux.HandleFunc("POST /api/v1/casefiles/{id}/finalize", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"case_file_id": testCaseFileID,
			"verdict":      map[string]any{"outcome": "pass", "evaluated_at": "2025-06-20T10:00:00Z"},
			"counters": map[string]any{
				"findings_count": 0, "coverage_percent": 85.0,
				"new_count": 0, "existing_count": 0, "resolved_count": 0, "has_tracking": false,
			},
		})
	})

	mux.HandleFunc("POST /api/v1/projects/{key}/pleadings", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{
			"pleading_id": testPleadingID,
		})
	})

	mux.HandleFunc("POST /api/v1/projects/{key}/source", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return httptest.NewServer(mux)
}

func newFailServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

func serverDeps(t *testing.T, serverURL string) Deps {
	t.Helper()
	deps, _ := newTestDeps(t)
	client, err := apiclient.New(serverURL, "test-token")
	require.NoError(t, err)
	deps.ServerClient = client
	return deps
}

func TestRunServer_SubmitsToServer(t *testing.T) {
	srv := newMockServer(t)
	t.Cleanup(srv.Close)

	deps, projectRepo := newTestDeps(t)
	client, err := apiclient.New(srv.URL, "test-token")
	require.NoError(t, err)
	deps.ServerClient = client

	project, pErr := newTestProject(t, projectRepo)
	require.NoError(t, pErr)

	result, err := RunServer(
		context.Background(), deps, "/workspace",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		localTenantID, project.ID().String(), "backend", "abc123", "main",
		time.Now(), Options{Quick: true},
	)

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict)
	assert.False(t, result.ServerFailed)
}

func TestRunServer_ServerFailureNonRequired(t *testing.T) {
	srv := newFailServer(t)
	t.Cleanup(srv.Close)

	deps, projectRepo := newTestDeps(t)
	client, err := apiclient.New(srv.URL, "test-token")
	require.NoError(t, err)
	deps.ServerClient = client

	project, pErr := newTestProject(t, projectRepo)
	require.NoError(t, pErr)

	result, err := RunServer(
		context.Background(), deps, "/workspace",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		localTenantID, project.ID().String(), "backend", "abc123", "main",
		time.Now(), Options{Quick: true, RequireSubmit: false},
	)

	require.NoError(t, err)
	assert.True(t, result.ServerFailed)
}

func TestRunServer_ServerFailureRequired(t *testing.T) {
	srv := newFailServer(t)
	t.Cleanup(srv.Close)

	deps, projectRepo := newTestDeps(t)
	client, err := apiclient.New(srv.URL, "test-token")
	require.NoError(t, err)
	deps.ServerClient = client

	project, pErr := newTestProject(t, projectRepo)
	require.NoError(t, pErr)

	_, err = RunServer(
		context.Background(), deps, "/workspace",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		localTenantID, project.ID().String(), "backend", "abc123", "main",
		time.Now(), Options{Quick: true, RequireSubmit: true},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "submit to server")
}

func TestRunServer_FilesPleading(t *testing.T) {
	pleadingCalled := false
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/projects/{key}", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": testProjectID, "key": "backend", "name": "backend",
			"target_pattern": "//...", "default_branch": "main",
			"languages": []string{}, "created_at": "2025-01-01T00:00:00Z",
		})
	})
	mux.HandleFunc("POST /api/v1/casefiles", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{"case_file_id": testCaseFileID})
	})
	mux.HandleFunc("POST /api/v1/casefiles/{id}/evidence", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{"evidence_id": "880e8400-e29b-41d4-a716-446655440000"})
	})
	mux.HandleFunc("POST /api/v1/casefiles/{id}/finalize", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"case_file_id": testCaseFileID,
			"verdict":      map[string]any{"outcome": "pass", "evaluated_at": "2025-06-20T10:00:00Z"},
			"counters":     map[string]any{"findings_count": 0, "coverage_percent": 0, "new_count": 0, "existing_count": 0, "resolved_count": 0, "has_tracking": false},
		})
	})
	mux.HandleFunc("POST /api/v1/projects/{key}/pleadings", func(w http.ResponseWriter, _ *http.Request) {
		pleadingCalled = true
		writeJSON(t, w, http.StatusCreated, map[string]any{"pleading_id": testPleadingID})
	})
	mux.HandleFunc("POST /api/v1/projects/{key}/source", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	deps, projectRepo := newTestDeps(t)
	client, err := apiclient.New(srv.URL, "test-token")
	require.NoError(t, err)
	deps.ServerClient = client

	project, pErr := newTestProject(t, projectRepo)
	require.NoError(t, pErr)

	_, err = RunServer(
		context.Background(), deps, "/workspace",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		localTenantID, project.ID().String(), "backend", "abc123", "main",
		time.Now(), Options{Quick: true, PRNumber: 42, PRTitle: "fix", PRAuthor: "user", PRBranch: "feat/fix"},
	)

	require.NoError(t, err)
	assert.True(t, pleadingCalled)
}

func TestRunProject_WithServerDelegatesToServer(t *testing.T) {
	srv := newMockServer(t)
	t.Cleanup(srv.Close)

	deps, projectRepo := newTestDeps(t)
	client, err := apiclient.New(srv.URL, "test-token")
	require.NoError(t, err)
	deps.ServerClient = client

	project, pErr := newTestProject(t, projectRepo)
	require.NoError(t, pErr)

	result, err := RunProject(
		context.Background(), deps, "/workspace",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		localTenantID, project.ID().String(), "backend", "abc123", "main",
		time.Now(), Options{Quick: true},
	)

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict)
}

func TestTryCreateProject_DefaultsTargetPattern(t *testing.T) {
	var receivedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/v1/projects":
			if err := json.NewDecoder(request.Body).Decode(&receivedBody); err != nil {
				t.Fatalf("decode: %v", err)
			}
			writeJSON(t, writer, http.StatusCreated, map[string]any{"id": testProjectID, "key": "new", "name": "new"})
		case request.Method == http.MethodGet:
			writeJSON(t, writer, http.StatusOK, map[string]any{
				"id": testProjectID, "key": "new", "name": "new",
				"target_pattern": "//...", "default_branch": "main",
				"languages": []string{}, "created_at": "2025-01-01T00:00:00Z",
			})
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)

	detail, err := tryCreateProject(context.Background(), deps, "new", "")

	require.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Equal(t, "//...", receivedBody["target_pattern"])
}

func TestUploadSourceFiles_UploadsReferencedFiles(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "a.go"), []byte("package a"), 0o644))

	var uploadCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			uploadCalled = true
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		writer.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	findings := []evidencedto.Finding{
		{FilePath: "a.go", Tool: "PMD", RuleID: "R1", Severity: "error", Line: 1, Message: "m", FingerprintID: "fp1"},
	}

	uploadSourceFiles(context.Background(), deps, "backend", "abc123", findings, workspace)

	assert.True(t, uploadCalled)
}

func TestUploadSourceFiles_SkipsWhenNoWorkspace(t *testing.T) {
	deps, _ := newTestDeps(t)
	findings := []evidencedto.Finding{
		{FilePath: "a.go", Tool: "PMD", RuleID: "R1", Severity: "error", Line: 1, Message: "m", FingerprintID: "fp1"},
	}

	uploadSourceFiles(context.Background(), deps, "backend", "abc123", findings, "")
}

func TestUploadSourceFiles_SkipsWhenNoFindings(t *testing.T) {
	deps, _ := newTestDeps(t)

	uploadSourceFiles(context.Background(), deps, "backend", "abc123", nil, "/workspace")
}

func TestUploadSourceFiles_DeduplicatesFiles(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "a.go"), []byte("package a"), 0o644))

	var uploadCount int
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err == nil {
				if files, ok := body["files"].([]any); ok {
					uploadCount = len(files)
				}
			}
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		writer.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	findings := []evidencedto.Finding{
		{FilePath: "a.go", Tool: "PMD", RuleID: "R1", Severity: "error", Line: 1, Message: "m", FingerprintID: "fp1"},
		{FilePath: "a.go", Tool: "PMD", RuleID: "R2", Severity: "warning", Line: 2, Message: "m", FingerprintID: "fp2"},
	}

	uploadSourceFiles(context.Background(), deps, "backend", "abc123", findings, workspace)

	assert.Equal(t, 1, uploadCount)
}

func TestUploadSourceFiles_SkipsMissingFiles(t *testing.T) {
	workspace := t.TempDir()

	var uploadCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			uploadCalled = true
			w.WriteHeader(http.StatusNoContent)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	findings := []evidencedto.Finding{
		{FilePath: "nonexistent.go", Tool: "PMD", RuleID: "R1", Severity: "error", Line: 1, Message: "m", FingerprintID: "fp1"},
	}

	uploadSourceFiles(context.Background(), deps, "backend", "abc123", findings, workspace)

	assert.False(t, uploadCalled)
}

func TestSubmitToServer_FetchProjectFallsBackToCreate(t *testing.T) {
	fetchCalled := false
	createCalled := false
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/projects/{key}", func(writer http.ResponseWriter, _ *http.Request) {
		fetchCalled = true
		if !createCalled {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writeJSON(t, writer, http.StatusOK, map[string]any{
			"id": testProjectID, "key": "newproj", "name": "newproj",
			"target_pattern": "//...", "default_branch": "main",
			"languages": []string{}, "created_at": "2025-01-01T00:00:00Z",
		})
	})
	mux.HandleFunc("POST /api/v1/projects", func(writer http.ResponseWriter, _ *http.Request) {
		createCalled = true
		writeJSON(t, writer, http.StatusCreated, map[string]any{"id": testProjectID, "key": "newproj", "name": "newproj"})
	})
	mux.HandleFunc("POST /api/v1/casefiles", func(writer http.ResponseWriter, _ *http.Request) {
		writeJSON(t, writer, http.StatusCreated, map[string]any{"case_file_id": testCaseFileID})
	})
	mux.HandleFunc("POST /api/v1/casefiles/{id}/evidence", func(writer http.ResponseWriter, _ *http.Request) {
		writeJSON(t, writer, http.StatusCreated, map[string]any{"evidence_id": "880e8400-e29b-41d4-a716-446655440000"})
	})
	mux.HandleFunc("POST /api/v1/casefiles/{id}/finalize", func(writer http.ResponseWriter, _ *http.Request) {
		writeJSON(t, writer, http.StatusOK, map[string]any{
			"case_file_id": testCaseFileID,
			"verdict":      map[string]any{"outcome": "pass", "evaluated_at": "2025-06-20T10:00:00Z"},
			"counters":     map[string]any{"findings_count": 0, "coverage_percent": 0, "new_count": 0, "existing_count": 0, "resolved_count": 0, "has_tracking": false},
		})
	})
	mux.HandleFunc("POST /api/v1/projects/{key}/source", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	collected := collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}}
	localResult := Result{Verdict: "pass"}

	err := submitToServer(
		context.Background(), deps,
		"newproj", "abc123", "main",
		collected, localResult, Options{TargetPattern: "//newproj/..."},
	)

	require.NoError(t, err)
	assert.True(t, fetchCalled)
	assert.True(t, createCalled)
}

func TestRunServer_RunLocalError(t *testing.T) {
	srv := newMockServer(t)
	t.Cleanup(srv.Close)

	deps, _ := newTestDeps(t)
	client, err := apiclient.New(srv.URL, "test-token")
	require.NoError(t, err)
	deps.ServerClient = client

	_, err = RunServer(
		context.Background(), deps, "/workspace",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		localTenantID, "not-a-uuid", "backend", "abc123", "main",
		time.Now(), Options{Quick: true},
	)

	require.Error(t, err)
}

func TestRunServer_FilePleadingWarning(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/projects/{key}", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": testProjectID, "key": "backend", "name": "backend",
			"target_pattern": "//...", "default_branch": "main",
			"languages": []string{}, "created_at": "2025-01-01T00:00:00Z",
		})
	})
	mux.HandleFunc("POST /api/v1/casefiles", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{"case_file_id": testCaseFileID})
	})
	mux.HandleFunc("POST /api/v1/casefiles/{id}/evidence", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{"evidence_id": "880e8400-e29b-41d4-a716-446655440000"})
	})
	mux.HandleFunc("POST /api/v1/casefiles/{id}/finalize", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"case_file_id": testCaseFileID,
			"verdict":      map[string]any{"outcome": "pass", "evaluated_at": "2025-06-20T10:00:00Z"},
			"counters":     map[string]any{"findings_count": 0, "coverage_percent": 0, "new_count": 0, "existing_count": 0, "resolved_count": 0, "has_tracking": false},
		})
	})
	mux.HandleFunc("POST /api/v1/projects/{key}/pleadings", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("POST /api/v1/projects/{key}/source", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	deps, projectRepo := newTestDeps(t)
	client, err := apiclient.New(srv.URL, "test-token")
	require.NoError(t, err)
	deps.ServerClient = client

	project, pErr := newTestProject(t, projectRepo)
	require.NoError(t, pErr)

	result, err := RunServer(
		context.Background(), deps, "/workspace",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		localTenantID, project.ID().String(), "backend", "abc123", "main",
		time.Now(), Options{Quick: true, PRNumber: 42, PRTitle: "fix", PRAuthor: "user", PRBranch: "feat/fix"},
	)

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict)
}

func TestSubmitToServer_OpenCaseFileError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/projects/{key}", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": testProjectID, "key": "backend", "name": "backend",
			"target_pattern": "//...", "default_branch": "main",
			"languages": []string{}, "created_at": "2025-01-01T00:00:00Z",
		})
	})
	mux.HandleFunc("POST /api/v1/casefiles", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	err := submitToServer(context.Background(), deps, "backend", "abc123", "main",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		Result{Verdict: "pass"}, Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "open case file")
}

func TestSubmitToServer_IngestEvidenceError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/projects/{key}", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": testProjectID, "key": "backend", "name": "backend",
			"target_pattern": "//...", "default_branch": "main",
			"languages": []string{}, "created_at": "2025-01-01T00:00:00Z",
		})
	})
	mux.HandleFunc("POST /api/v1/casefiles", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{"case_file_id": testCaseFileID})
	})
	mux.HandleFunc("POST /api/v1/casefiles/{id}/evidence", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	err := submitToServer(context.Background(), deps, "backend", "abc123", "main",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		Result{Verdict: "pass"}, Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ingest sarif evidence")
}

func TestSubmitToServer_FinalizeError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/projects/{key}", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": testProjectID, "key": "backend", "name": "backend",
			"target_pattern": "//...", "default_branch": "main",
			"languages": []string{}, "created_at": "2025-01-01T00:00:00Z",
		})
	})
	mux.HandleFunc("POST /api/v1/casefiles", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{"case_file_id": testCaseFileID})
	})
	mux.HandleFunc("POST /api/v1/casefiles/{id}/evidence", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{"evidence_id": "880e8400-e29b-41d4-a716-446655440000"})
	})
	mux.HandleFunc("POST /api/v1/casefiles/{id}/finalize", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	err := submitToServer(context.Background(), deps, "backend", "abc123", "main",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		Result{Verdict: "pass"}, Options{})

	require.Error(t, err)
}

func TestSubmitToServer_WithRulings(t *testing.T) {
	srv := newMockServer(t)
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	localResult := Result{
		Verdict: "pass",
		Rulings: []corejudge.RulingView{
			{Subtype: "code_quality", Passed: true, Detail: "0 findings"},
			{Subtype: "coverage", Passed: true, Detail: "90%"},
		},
	}

	err := submitToServer(context.Background(), deps, "backend", "abc123", "main",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		localResult, Options{})

	require.NoError(t, err)
}

func TestSubmitToServer_FetchAndCreateBothFail(t *testing.T) {
	srv := newFailServer(t)
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	err := submitToServer(context.Background(), deps, "backend", "abc123", "main",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		Result{Verdict: "pass"}, Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetch project")
}

func TestUploadSourceFiles_EmptyFilePath(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "a.go"), []byte("package a"), 0o644))

	var uploadCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			uploadCalled = true
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	findings := []evidencedto.Finding{
		{FilePath: "", Tool: "PMD", RuleID: "R1", Severity: "error", Line: 1, Message: "m", FingerprintID: "fp1"},
		{FilePath: "a.go", Tool: "PMD", RuleID: "R2", Severity: "error", Line: 1, Message: "m", FingerprintID: "fp2"},
	}

	uploadSourceFiles(context.Background(), deps, "backend", "abc123", findings, workspace)

	assert.True(t, uploadCalled)
}

func TestUploadSourceFiles_UploadServerError(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "a.go"), []byte("package a"), 0o644))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	findings := []evidencedto.Finding{
		{FilePath: "a.go", Tool: "PMD", RuleID: "R1", Severity: "error", Line: 1, Message: "m", FingerprintID: "fp1"},
	}

	uploadSourceFiles(context.Background(), deps, "backend", "abc123", findings, workspace)
}

type failingFindingsParser struct{}

func (p failingFindingsParser) Parse(_ context.Context, _ []byte) ([]ingestfind.Parsed, error) {
	return nil, fmt.Errorf("parser failed")
}

type failingCoverageParser struct{}

func (p failingCoverageParser) Parse(_ context.Context, _ []byte) (ingestcov.Parsed, error) {
	return ingestcov.Parsed{}, fmt.Errorf("coverage parser failed")
}

func TestSubmitToServer_ParseEvidenceError(t *testing.T) {
	srv := newMockServer(t)
	t.Cleanup(srv.Close)

	deps := serverDeps(t, srv.URL)
	deps.Findings = ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": failingFindingsParser{}})

	collected := collectevidence.Result{
		RawSARIF: []collectevidence.RawFile{
			{Format: "sarif", Source: "tool.sarif", Data: []byte(`{}`)},
		},
	}

	err := submitToServer(context.Background(), deps, "backend", "abc123", "main",
		collected, Result{Verdict: "pass"}, Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse sarif evidence")
}

func TestParseEvidence_SARIFNewCommandError(t *testing.T) {
	deps, _ := newTestDeps(t)
	file := evidenceFile{Format: "sarif", Source: "tool.sarif", Data: nil}

	_, err := parseEvidence(context.Background(), deps, file)

	require.Error(t, err)
}

func TestParseEvidence_SARIFExecuteError(t *testing.T) {
	deps, _ := newTestDeps(t)
	deps.Findings = ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": failingFindingsParser{}})

	file := evidenceFile{Format: "sarif", Source: "tool.sarif", Data: []byte(`{}`)}

	_, err := parseEvidence(context.Background(), deps, file)

	require.Error(t, err)
}

func TestParseEvidence_LCOVNewCommandError(t *testing.T) {
	deps, _ := newTestDeps(t)
	file := evidenceFile{Format: "lcov", Source: "coverage.lcov", Data: nil}

	_, err := parseEvidence(context.Background(), deps, file)

	require.Error(t, err)
}

func TestParseEvidence_LCOVExecuteError(t *testing.T) {
	deps, _ := newTestDeps(t)
	deps.Coverage = ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": failingCoverageParser{}})

	file := evidenceFile{Format: "lcov", Source: "coverage.lcov", Data: []byte("SF:a.go\nDA:1,1\nend_of_record\n")}

	_, err := parseEvidence(context.Background(), deps, file)

	require.Error(t, err)
}

func newTestProject(t *testing.T, repo *projectmemory.ProjectRepository) (projectmodel.Project, error) {
	t.Helper()
	project, err := projectmodel.NewProject(tenant.LocalTenantID, "backend", "backend", "//backend/...")
	if err != nil {
		return projectmodel.Project{}, fmt.Errorf("new project: %w", err)
	}
	if err := repo.Save(context.Background(), project); err != nil {
		return projectmodel.Project{}, fmt.Errorf("save project: %w", err)
	}
	return project, nil
}
