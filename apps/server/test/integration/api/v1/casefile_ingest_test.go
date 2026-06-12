package v1integration

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func seedProjectInRepo(t *testing.T, f *testFixture) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject("ingest-test", "Ingest Test", "//...")
	require.NoError(t, err)
	p.ClearEvents()
	require.NoError(t, f.projRepo.Save(context.Background(), p))
	return p
}

func TestCreateCaseFile_Success(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)
	project := seedProjectInRepo(t, f)

	res := f.do(t, http.MethodPost, "/casefiles", map[string]any{
		"project_id": project.ID().String(),
		"commit_sha": "abc123",
		"branch":     "main",
	}, cookie)
	require.Equal(t, http.StatusCreated, res.Code, res.Body.String())

	var body struct {
		CaseFileID string `json:"case_file_id"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.NotEmpty(t, body.CaseFileID)
	_, err := uuid.Parse(body.CaseFileID)
	require.NoError(t, err)
}

func TestCreateCaseFile_ProjectNotFound(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPost, "/casefiles", map[string]any{
		"project_id": uuid.NewString(),
		"commit_sha": "abc123",
		"branch":     "main",
	}, cookie)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestCreateCaseFile_MissingBody(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPost, "/casefiles", map[string]any{}, cookie)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestIngestCaseFileEvidence_Success(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)
	project := seedProjectInRepo(t, fixture)

	createRes := fixture.do(t, http.MethodPost, "/casefiles", map[string]any{
		"project_id": project.ID().String(),
		"commit_sha": "abc123",
		"branch":     "main",
	}, cookie)
	require.Equal(t, http.StatusCreated, createRes.Code)
	var created struct {
		CaseFileID string `json:"case_file_id"`
	}
	mustDecode(t, createRes.Body.Bytes(), &created)

	ingestRes := fixture.do(t, http.MethodPost, "/casefiles/"+created.CaseFileID+"/evidence", map[string]any{
		"subtype":      "code_quality",
		"source":       "golangci-lint",
		"collected_at": time.Now().UTC().Format(time.RFC3339),
		"findings": []map[string]any{
			{
				"tool":        "golangci-lint",
				"rule_id":     "errcheck",
				"severity":    "error",
				"file_path":   "main.go",
				"line":        1,
				"message":     "x",
				"fingerprint": uuid.NewString(),
			},
		},
	}, cookie)
	require.Equal(t, http.StatusCreated, ingestRes.Code, ingestRes.Body.String())

	var ingested struct {
		EvidenceID string `json:"evidence_id"`
	}
	mustDecode(t, ingestRes.Body.Bytes(), &ingested)
	require.NotEmpty(t, ingested.EvidenceID)
}

func TestIngestCaseFileEvidence_CaseFileNotFound(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPost, "/casefiles/"+uuid.NewString()+"/evidence", map[string]any{
		"subtype":      "code_quality",
		"source":       "golangci-lint",
		"collected_at": time.Now().UTC().Format(time.RFC3339),
		"findings": []map[string]any{
			{
				"tool":        "golangci-lint",
				"rule_id":     "errcheck",
				"severity":    "error",
				"file_path":   "main.go",
				"line":        1,
				"message":     "x",
				"fingerprint": uuid.NewString(),
			},
		},
	}, cookie)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestIngestCaseFileEvidence_MissingContent(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPost, "/casefiles/"+uuid.NewString()+"/evidence", map[string]any{
		"subtype":      "code_quality",
		"source":       "golangci-lint",
		"collected_at": time.Now().UTC().Format(time.RFC3339),
	}, cookie)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestIngestCaseFileEvidence_BodyTooLarge(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)
	project := seedProjectInRepo(t, fixture)

	createRes := fixture.do(t, http.MethodPost, "/casefiles", map[string]any{
		"project_id": project.ID().String(),
		"commit_sha": "abc123",
		"branch":     "main",
	}, cookie)
	require.Equal(t, http.StatusCreated, createRes.Code)
	var created struct {
		CaseFileID string `json:"case_file_id"`
	}
	mustDecode(t, createRes.Body.Bytes(), &created)

	var oversized bytes.Buffer
	oversized.WriteString(`{"subtype":"`)
	oversized.Write(bytes.Repeat([]byte("A"), 11<<20))
	oversized.WriteString(`"}`)

	req := httptest.NewRequest(http.MethodPost, "/casefiles/"+created.CaseFileID+"/evidence", &oversized)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	fixture.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}
