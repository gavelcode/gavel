package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/userinterface/api/v1/client"
)

func TestNew_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		gotAuth = req.Header.Get("Authorization")
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"detail":"not found"}`))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "my-secret-token")
	require.NoError(t, err)

	_, _ = apiClient.FetchProject(context.Background(), "any")
	assert.Equal(t, "Bearer my-secret-token", gotAuth)
}

func TestNew_NoTokenOmitsHeader(t *testing.T) {
	var gotAuth string
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		gotAuth = req.Header.Get("Authorization")
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"detail":"not found"}`))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "")
	require.NoError(t, err)

	_, _ = apiClient.FetchProject(context.Background(), "any")
	assert.Empty(t, gotAuth)
}

func TestFetchProject_ReturnsDetail(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/projects/core", req.URL.Path)

		resp := map[string]any{
			"id":             "00000000-0000-0000-0000-000000000001",
			"key":            "core",
			"name":           "Core Module",
			"default_branch": "main",
			"languages":      []string{"go"},
			"created_at":     "2026-01-01T00:00:00Z",
		}
		writer.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(writer).Encode(resp))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	detail, err := apiClient.FetchProject(context.Background(), "core")

	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", detail.ID)
	assert.Equal(t, "core", detail.Key)
	assert.Equal(t, "Core Module", detail.Name)
	assert.Equal(t, "main", detail.DefaultBranch)
}

func TestFetchProject_NotFound(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"detail":"not found"}`))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	_, err = apiClient.FetchProject(context.Background(), "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestCreateProject_ReturnsDetail(t *testing.T) {
	callCount := 0
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		callCount++
		writer.Header().Set("Content-Type", "application/json")
		if req.Method == http.MethodPost {
			resp := map[string]any{
				"project_id": "00000000-0000-0000-0000-000000000002",
			}
			writer.WriteHeader(http.StatusCreated)
			require.NoError(t, json.NewEncoder(writer).Encode(resp))
			return
		}
		resp := map[string]any{
			"id":             "00000000-0000-0000-0000-000000000002",
			"key":            "new-proj",
			"name":           "New Project",
			"default_branch": "main",
			"languages":      []string{},
			"created_at":     "2026-01-01T00:00:00Z",
		}
		require.NoError(t, json.NewEncoder(writer).Encode(resp))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	detail, err := apiClient.CreateProject(context.Background(), "new-proj", "New Project", "//new/...")

	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, "new-proj", detail.Key)
	assert.Equal(t, 2, callCount)
}

func TestCreateProject_ServerError(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(`{"detail":"internal"}`))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	_, err = apiClient.CreateProject(context.Background(), "x", "X", "//x/...")
	assert.Error(t, err)
}

func TestOpenCaseFile_ReturnsID(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/api/v1/casefiles", req.URL.Path)

		resp := map[string]any{
			"case_file_id": "00000000-0000-0000-0000-000000000003",
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(writer).Encode(resp))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	caseFileID, err := apiClient.OpenCaseFile(context.Background(), "00000000-0000-0000-0000-000000000099", "abc123", "main", false)

	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", caseFileID)
}

func TestOpenCaseFile_ServerRejects(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(`{"detail":"bad"}`))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	_, err = apiClient.OpenCaseFile(context.Background(), "00000000-0000-0000-0000-000000000099", "abc", "main", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestIngestCaseFileEvidence_ReturnsEvidenceID(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Contains(t, req.URL.Path, "/evidence")

		resp := map[string]any{
			"evidence_id": "00000000-0000-0000-0000-000000000004",
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(writer).Encode(resp))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	evidenceID, err := apiClient.IngestCaseFileEvidence(context.Background(),
		"00000000-0000-0000-0000-000000000003",
		client.EvidenceToWire(evidencedto.Evidence{
			Subtype:     "code_quality",
			Source:      "golangci-lint",
			CollectedAt: time.Now().UTC(),
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000004", evidenceID)
}

func TestFinalizeCaseFileWithVerdict_ReturnsResult(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		rulings := []map[string]any{
			{"subtype": "code_quality", "passed": true, "detail": "0 new findings"},
		}
		resp := map[string]any{
			"case_file_id": "00000000-0000-0000-0000-000000000003",
			"verdict": map[string]any{
				"outcome":      "pass",
				"rulings":      rulings,
				"evaluated_at": now.Format(time.RFC3339),
			},
			"counters": map[string]any{
				"findings_count":   10,
				"coverage_percent": 92.5,
				"new_count":        2,
				"existing_count":   7,
				"resolved_count":   1,
				"has_tracking":     true,
			},
		}
		writer.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(writer).Encode(resp))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	result, err := apiClient.FinalizeCaseFileWithVerdict(context.Background(),
		"00000000-0000-0000-0000-000000000003",
		client.VerdictResult{
			Outcome:     "pass",
			EvaluatedAt: now,
			Rulings:     []client.RulingResult{{Subtype: "code_quality", Passed: true, Detail: "0 new findings"}},
		},
		client.CountersResult{FindingsCount: 10, CoveragePercent: 92.5, NewCount: 2, ExistingCount: 7, ResolvedCount: 1, HasTracking: true},
	)

	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", result.CaseFileID)
	assert.Equal(t, "pass", result.Verdict.Outcome)
	require.Len(t, result.Verdict.Rulings, 1)
	assert.True(t, result.Verdict.Rulings[0].Passed)
	assert.Equal(t, 10, result.Counters.FindingsCount)
	assert.InDelta(t, 92.5, result.Counters.CoveragePercent, 0.01)
	assert.Equal(t, 2, result.Counters.NewCount)
	assert.True(t, result.Counters.HasTracking)
}

func TestFinalizeCaseFile_ServerError(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(`{"detail":"fail"}`))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	_, err = apiClient.FinalizeCaseFileWithVerdict(context.Background(),
		"00000000-0000-0000-0000-000000000003",
		client.VerdictResult{Outcome: "pass", EvaluatedAt: time.Now()},
		client.CountersResult{},
	)
	assert.Error(t, err)
}

func TestFetchBaseline_ReturnsData(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		assert.Contains(t, req.URL.Path, "/baseline")
		assert.Equal(t, "main", req.URL.Query().Get("branch"))

		resp := map[string]any{
			"fingerprints":     []string{"fp-1", "fp-2"},
			"architecture_ids": []string{"arch-1"},
			"has_previous":     true,
		}
		writer.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(writer).Encode(resp))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	baseline, err := apiClient.FetchBaseline(context.Background(), "core", "main")

	require.NoError(t, err)
	assert.Equal(t, []string{"fp-1", "fp-2"}, baseline.Fingerprints)
	assert.Equal(t, []string{"arch-1"}, baseline.ArchViolationIDs)
	assert.True(t, baseline.HasPrevious)
}

func TestFetchBaseline_NotFound(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"detail":"not found"}`))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	_, err = apiClient.FetchBaseline(context.Background(), "core", "main")
	assert.Error(t, err)
}

func TestFilePleading_ReturnsID(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Contains(t, req.URL.Path, "/pleadings")

		resp := map[string]any{
			"pleading_id": "00000000-0000-0000-0000-000000000005",
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(writer).Encode(resp))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	pleadingID, err := apiClient.FilePleading(context.Background(), "core", 42, "Fix bug", "jorge", "fix/bug", "main", "abc123")

	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000005", pleadingID)
}

func TestFilePleading_ServerRejects(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(`{"detail":"bad"}`))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	_, err = apiClient.FilePleading(context.Background(), "core", 1, "t", "a", "b", "c", "d")
	assert.Error(t, err)
}

func TestUploadSource_Success(t *testing.T) {
	var gotBody map[string]any
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Contains(t, req.URL.Path, "/source")
		require.NoError(t, json.NewDecoder(req.Body).Decode(&gotBody))
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	err = apiClient.UploadSource(context.Background(), "core", "abc123", []client.SourceFile{
		{Path: "main.go", Content: "package main"},
	})

	require.NoError(t, err)
	assert.Equal(t, "abc123", gotBody["commit"])
}

func TestUploadSource_EmptyFilesSkips(t *testing.T) {
	apiClient, err := client.New("http://unused", "tok")
	require.NoError(t, err)

	err = apiClient.UploadSource(context.Background(), "core", "abc", nil)
	require.NoError(t, err)
}

func TestUploadSource_ServerError(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "tok")
	require.NoError(t, err)

	err = apiClient.UploadSource(context.Background(), "core", "abc", []client.SourceFile{
		{Path: "x.go", Content: "x"},
	})
	assert.Error(t, err)
}

func TestEvidenceToWire_FindingsEvidence(t *testing.T) {
	input := evidencedto.Evidence{
		ID:          "ev-1",
		Subtype:     "code_quality",
		Source:      "golangci-lint",
		CollectedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		Findings: []evidencedto.Finding{
			{
				Tool:          "golangci-lint",
				RuleID:        "varnamelen",
				Severity:      "error",
				FilePath:      "main.go",
				Line:          10,
				Message:       "too short",
				FingerprintID: "fp-1",
			},
		},
	}

	wire := client.EvidenceToWire(input)

	assert.Equal(t, "code_quality", string(wire.Subtype))
	assert.Equal(t, "golangci-lint", wire.Source)
	require.NotNil(t, wire.Id)
	assert.Equal(t, "ev-1", *wire.Id)
	require.NotNil(t, wire.Findings)
	require.Len(t, *wire.Findings, 1)
	assert.Equal(t, "varnamelen", (*wire.Findings)[0].RuleId)
	assert.Equal(t, int32(10), (*wire.Findings)[0].Line)
}

func TestEvidenceToWire_CoverageEvidence(t *testing.T) {
	input := evidencedto.Evidence{
		Subtype:     "coverage",
		Source:      "go-test",
		CollectedAt: time.Now().UTC(),
		Coverage: &evidencedto.Coverage{
			TotalLines:   200,
			CoveredLines: 180,
			ByLanguage: []evidencedto.LanguageStats{
				{Language: "go", TotalLines: 200, CoveredLines: 180},
			},
			ByFile: []evidencedto.FileCoverage{
				{FilePath: "main.go", Covered: []int{1, 2, 3}, Uncovered: []int{4}},
			},
		},
	}

	wire := client.EvidenceToWire(input)

	require.NotNil(t, wire.Coverage)
	assert.Equal(t, int32(200), wire.Coverage.TotalLines)
	assert.Equal(t, int32(180), wire.Coverage.CoveredLines)
	require.NotNil(t, wire.Coverage.ByLanguage)
	require.Len(t, *wire.Coverage.ByLanguage, 1)
	assert.Equal(t, "go", (*wire.Coverage.ByLanguage)[0].Language)
	require.NotNil(t, wire.Coverage.ByFile)
	require.Len(t, *wire.Coverage.ByFile, 1)
	assert.Equal(t, "main.go", (*wire.Coverage.ByFile)[0].FilePath)
	assert.Equal(t, []int32{1, 2, 3}, (*wire.Coverage.ByFile)[0].CoveredLines)
}

func TestEvidenceToWire_ArchitectureEvidence(t *testing.T) {
	input := evidencedto.Evidence{
		Subtype:     "architecture",
		Source:      "archtest",
		CollectedAt: time.Now().UTC(),
		Architecture: &evidencedto.Architecture{
			Violations: []evidencedto.Violation{
				{Rule: "layer-dep", SourcePkg: "domain/a", TargetPkg: "infra/b", Message: "bad import"},
			},
		},
	}

	wire := client.EvidenceToWire(input)

	require.NotNil(t, wire.Architecture)
	require.Len(t, wire.Architecture.Violations, 1)
	assert.Equal(t, "layer-dep", wire.Architecture.Violations[0].Rule)
	assert.Equal(t, "domain/a", wire.Architecture.Violations[0].SourcePkg)
}

func TestEvidenceToWire_NewCodeCoverage(t *testing.T) {
	input := evidencedto.Evidence{
		Subtype:     "coverage",
		Source:      "go-test",
		CollectedAt: time.Now().UTC(),
		NewCodeCoverage: &evidencedto.NewCodeCoverage{
			CoveredLines:   15,
			CoverableLines: 20,
		},
	}

	wire := client.EvidenceToWire(input)

	require.NotNil(t, wire.NewCodeCoverage)
	assert.Equal(t, int32(15), wire.NewCodeCoverage.CoveredLines)
	assert.Equal(t, int32(20), wire.NewCodeCoverage.CoverableLines)
}

func TestEvidenceToWire_MinimalEvidence(t *testing.T) {
	input := evidencedto.Evidence{
		Subtype:     "code_quality",
		Source:      "test",
		CollectedAt: time.Now().UTC(),
	}

	wire := client.EvidenceToWire(input)

	assert.Nil(t, wire.Id)
	assert.Nil(t, wire.Findings)
	assert.Nil(t, wire.Coverage)
	assert.Nil(t, wire.Architecture)
	assert.Nil(t, wire.NewCodeCoverage)
}

func TestListProjectCaseFiles_ReturnsHistory(t *testing.T) {
	cov := 92.3
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/projects/core/casefiles", req.URL.Path)
		assert.Equal(t, "10", req.URL.Query().Get("limit"))

		resp := map[string]any{
			"items": []map[string]any{
				{
					"id":                "00000000-0000-0000-0000-000000000001",
					"project_id":        "00000000-0000-0000-0000-000000000099",
					"commit_sha":        "abc1234",
					"branch":            "main",
					"coverage_percent":  cov,
					"total_findings":    45,
					"new_findings":      2,
					"existing_findings": 40,
					"resolved_findings": 3,
					"verdict_outcome":   "pass",
					"started_at":        "2026-06-14T10:00:00Z",
					"created_at":        "2026-06-14T10:01:00Z",
				},
				{
					"id":                "00000000-0000-0000-0000-000000000002",
					"project_id":        "00000000-0000-0000-0000-000000000099",
					"commit_sha":        "def5678",
					"branch":            "main",
					"total_findings":    47,
					"new_findings":      3,
					"existing_findings": 43,
					"resolved_findings": 1,
					"verdict_outcome":   "fail",
					"started_at":        "2026-06-14T09:00:00Z",
					"created_at":        "2026-06-14T09:01:00Z",
				},
			},
			"next_cursor": nil,
		}
		writer.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(writer).Encode(resp))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "test-token")
	require.NoError(t, err)

	entries, err := apiClient.ListProjectCaseFiles(context.Background(), "core", 10)

	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "abc1234", entries[0].CommitSHA)
	assert.Equal(t, "main", entries[0].Branch)
	assert.InDelta(t, 92.3, *entries[0].CoveragePercent, 0.01)
	assert.Equal(t, 45, entries[0].TotalFindings)
	assert.Equal(t, 2, entries[0].NewFindings)
	assert.Equal(t, 3, entries[0].ResolvedFindings)
	assert.Equal(t, "pass", entries[0].VerdictOutcome)
	assert.Equal(t, 2026, entries[0].CreatedAt.Year())

	assert.Equal(t, "def5678", entries[1].CommitSHA)
	assert.Nil(t, entries[1].CoveragePercent)
	assert.Equal(t, "fail", entries[1].VerdictOutcome)
}

func TestListProjectCaseFiles_EmptyList(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{"items": []any{}, "next_cursor": nil}
		writer.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(writer).Encode(resp))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "test-token")
	require.NoError(t, err)

	entries, err := apiClient.ListProjectCaseFiles(context.Background(), "core", 10)

	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestListProjectCaseFiles_NotFound(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		require.NoError(t, json.NewEncoder(writer).Encode(map[string]string{"detail": "not found"}))
	}))
	defer testSrv.Close()

	apiClient, err := client.New(testSrv.URL, "test-token")
	require.NoError(t, err)

	_, err = apiClient.ListProjectCaseFiles(context.Background(), "nonexistent", 10)

	assert.Error(t, err)
}
