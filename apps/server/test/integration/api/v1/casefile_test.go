package v1integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	casefileget "github.com/usegavel/gavel/core/application/casefile/get"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
)

func seedCaseFile(f *testFixture, projectID string) string {
	caseFileID := uuid.NewString()
	f.casefiles.putDetail(&casefileget.CaseFileDetail{
		ID:               caseFileID,
		ProjectID:        projectID,
		CommitSHA:        "abc123",
		Branch:           "main",
		StartedAt:        time.Date(2026, time.June, 6, 9, 0, 0, 0, time.UTC),
		VerdictOutcome:   "pass",
		TotalFindings:    3,
		NewFindings:      1,
		ExistingFindings: 2,
		ResolvedFindings: 0,
		CreatedAt:        time.Date(2026, time.June, 6, 9, 0, 0, 0, time.UTC),
		Evidences:        []casefileget.EvidenceSummary{},
		Rulings:          []casefileget.RulingView{},
	})
	return caseFileID
}

func TestGetCaseFile_Success(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)
	caseFileID := seedCaseFile(f, uuid.NewString())

	res := f.do(t, http.MethodGet, "/casefiles/"+caseFileID, nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var body struct {
		ID             string `json:"id"`
		VerdictOutcome string `json:"verdict_outcome"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Equal(t, caseFileID, body.ID)
	require.Equal(t, "pass", body.VerdictOutcome)
}

func TestGetCaseFile_NotFound(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/casefiles/"+uuid.NewString(), nil, cookie)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestListFindings_Empty(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/findings", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
}

func TestListFindings_FiltersBySeverity(t *testing.T) {
	fixture := newTestFixture(t)
	cookie := fixture.loginCookie(t, adminEmail, adminPassword)

	fixture.casefiles.putFinding(findinglist.FindingView{
		Tool: "golangci-lint", RuleID: "errcheck", Severity: "error",
		FilePath: "main.go", Line: 1, Message: "x", FingerprintID: "fp1",
		Status: "new", Source: "lint", CommitSHA: "abc", ProjectKey: "core",
	})
	fixture.casefiles.putFinding(findinglist.FindingView{
		Tool: "golangci-lint", RuleID: "ineffassign", Severity: "warning",
		FilePath: "main.go", Line: 2, Message: "y", FingerprintID: "fp2",
		Status: "new", Source: "lint", CommitSHA: "abc", ProjectKey: "core",
	})

	res := fixture.do(t, http.MethodGet, "/findings?severity=error", nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var body struct {
		Items []struct {
			Severity string `json:"severity"`
		} `json:"items"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Len(t, body.Items, 1)
	require.Equal(t, "error", body.Items[0].Severity)
}
