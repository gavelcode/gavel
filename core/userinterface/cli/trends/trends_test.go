package trends_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/trends"
)

func TestNewCommand_ReturnsCobraCommand(t *testing.T) {
	cmd := trends.NewCommand()

	assert.Equal(t, "trends", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewCommand_HasProjectFlag(t *testing.T) {
	cmd := trends.NewCommand()

	f := cmd.Flags().Lookup("project")
	require.NotNil(t, f)
	assert.Equal(t, "", f.DefValue)
}

func TestNewCommand_HasLimitFlag(t *testing.T) {
	cmd := trends.NewCommand()

	f := cmd.Flags().Lookup("limit")
	require.NotNil(t, f)
	assert.Equal(t, "10", f.DefValue)
}

func TestNewCommand_RequiresProject(t *testing.T) {
	cmd := trends.NewCommand()
	cmd.SetArgs([]string{"--server", "http://localhost:8080", "--token", "tok"})

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project")
}

func TestNewCommand_RequiresServer(t *testing.T) {
	cmd := trends.NewCommand()
	cmd.SetArgs([]string{"--project", "core"})

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server")
}

type caseFileResponse struct {
	Items []caseFileItem `json:"items"`
}

type caseFileItem struct {
	ID               string   `json:"id"`
	ProjectID        string   `json:"project_id"`
	CommitSHA        string   `json:"commit_sha"`
	Branch           string   `json:"branch"`
	CoveragePercent  *float64 `json:"coverage_percent,omitempty"`
	TotalFindings    int      `json:"total_findings"`
	NewFindings      int      `json:"new_findings"`
	ResolvedFindings int      `json:"resolved_findings"`
	ExistingFindings int      `json:"existing_findings"`
	VerdictOutcome   string   `json:"verdict_outcome"`
	CreatedAt        string   `json:"created_at"`
	StartedAt        string   `json:"started_at"`
}

func trendServer(t *testing.T, items []caseFileItem) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		resp := caseFileResponse{Items: items}
		if err := json.NewEncoder(writer).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
}

func sampleItems() []caseFileItem {
	cov1 := 92.5
	cov2 := 90.0
	now := time.Now()
	return []caseFileItem{
		{
			ID: "11111111-1111-1111-1111-111111111111", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "abc1234567890", Branch: "main", CoveragePercent: &cov1,
			TotalFindings: 5, NewFindings: 1, ResolvedFindings: 2, ExistingFindings: 2,
			VerdictOutcome: "pass", CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
			StartedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			ID: "33333333-3333-3333-3333-333333333333", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "def5678901234", Branch: "main", CoveragePercent: &cov2,
			TotalFindings: 7, NewFindings: 0, ResolvedFindings: 0, ExistingFindings: 7,
			VerdictOutcome: "fail", CreatedAt: now.Add(-48 * time.Hour).Format(time.RFC3339),
			StartedAt: now.Add(-48 * time.Hour).Format(time.RFC3339),
		},
	}
}

func TestTrends_TableOutput(t *testing.T) {
	srv := trendServer(t, sampleItems())
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "test-token"})

	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Trends — core")
	assert.Contains(t, output, "abc1234")
	assert.Contains(t, output, "def5678")
	assert.Contains(t, output, "92.5%")
	assert.Contains(t, output, "pass")
	assert.Contains(t, output, "fail")
	assert.Contains(t, output, "Coverage:")
	assert.Contains(t, output, "Findings:")
	assert.Contains(t, output, "Pass rate: 1/2")
}

func TestTrends_JSONOutput(t *testing.T) {
	srv := trendServer(t, sampleItems())
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "test-token", "--json"})

	err := cmd.Execute()

	require.NoError(t, err)
	var entries []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entries))
	assert.Len(t, entries, 2)
	assert.Equal(t, "abc1234567890", entries[0]["commit_sha"])
}

func TestTrends_EmptyHistory(t *testing.T) {
	srv := trendServer(t, nil)
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "test-token"})

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No analysis history")
}

func TestTrends_BranchFilter(t *testing.T) {
	cov := 85.0
	now := time.Now()
	items := []caseFileItem{
		{
			ID: "11111111-1111-1111-1111-111111111111", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "aaa1111", Branch: "main", CoveragePercent: &cov,
			TotalFindings: 0, VerdictOutcome: "pass",
			CreatedAt: now.Format(time.RFC3339), StartedAt: now.Format(time.RFC3339),
		},
		{
			ID: "33333333-3333-3333-3333-333333333333", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "bbb2222", Branch: "feat/x", CoveragePercent: &cov,
			TotalFindings: 3, VerdictOutcome: "fail",
			CreatedAt: now.Add(-time.Hour).Format(time.RFC3339), StartedAt: now.Add(-time.Hour).Format(time.RFC3339),
		},
	}
	srv := trendServer(t, items)
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok", "--branch", "main"})

	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "aaa1111")
	assert.NotContains(t, output, "bbb2222")
	assert.Contains(t, output, "branch: main")
}

func TestTrends_EmptyAfterBranchFilter(t *testing.T) {
	cov := 85.0
	now := time.Now()
	items := []caseFileItem{
		{
			ID: "11111111-1111-1111-1111-111111111111", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "aaa", Branch: "main", CoveragePercent: &cov, VerdictOutcome: "pass",
			CreatedAt: now.Format(time.RFC3339), StartedAt: now.Format(time.RFC3339),
		},
	}
	srv := trendServer(t, items)
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok", "--branch", "nonexistent"})

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No analysis history")
	assert.Contains(t, buf.String(), "nonexistent")
}

func TestTrends_CoverageDirectionArrows(t *testing.T) {
	now := time.Now()
	cov1 := 95.0
	cov2 := 85.0
	items := []caseFileItem{
		{
			ID: "11111111-1111-1111-1111-111111111111", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "newer", Branch: "main", CoveragePercent: &cov1,
			TotalFindings: 3, VerdictOutcome: "pass",
			CreatedAt: now.Format(time.RFC3339), StartedAt: now.Format(time.RFC3339),
		},
		{
			ID: "33333333-3333-3333-3333-333333333333", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "older", Branch: "main", CoveragePercent: &cov2,
			TotalFindings: 8, VerdictOutcome: "fail",
			CreatedAt: now.Add(-24 * time.Hour).Format(time.RFC3339), StartedAt: now.Add(-24 * time.Hour).Format(time.RFC3339),
		},
	}
	srv := trendServer(t, items)
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok"})

	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "95.0%")
	assert.Contains(t, output, fmt.Sprintf("%+.1f%%", 10.0))
	assert.Contains(t, output, "Pass rate: 1/2")
}

func TestTrends_NoCoverage(t *testing.T) {
	now := time.Now()
	items := []caseFileItem{
		{
			ID: "11111111-1111-1111-1111-111111111111", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "aaa", Branch: "main", TotalFindings: 0, VerdictOutcome: "pass",
			CreatedAt: now.Format(time.RFC3339), StartedAt: now.Format(time.RFC3339),
		},
	}
	srv := trendServer(t, items)
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok"})

	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Findings:")
	assert.NotContains(t, output, "Coverage:")
}

func TestTrends_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok"})

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fetch trends")
}

func TestTrends_TimeAgoDisplay(t *testing.T) {
	now := time.Now()
	cov := 80.0
	items := []caseFileItem{
		{
			ID: "11111111-1111-1111-1111-111111111111", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "recent", Branch: "main", CoveragePercent: &cov,
			TotalFindings: 0, VerdictOutcome: "pass",
			CreatedAt: now.Add(-30 * time.Second).Format(time.RFC3339), StartedAt: now.Format(time.RFC3339),
		},
		{
			ID: "33333333-3333-3333-3333-333333333333", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "hoursag", Branch: "main", CoveragePercent: &cov,
			TotalFindings: 0, VerdictOutcome: "pass",
			CreatedAt: now.Add(-5 * time.Hour).Format(time.RFC3339), StartedAt: now.Format(time.RFC3339),
		},
	}
	srv := trendServer(t, items)
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok"})

	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "just now")
	assert.Contains(t, output, "5h ago")
}

func TestTrends_CustomLimit(t *testing.T) {
	srv := trendServer(t, nil)
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok", "--limit", "5"})

	err := cmd.Execute()

	require.NoError(t, err)
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

type failOnNthWriter struct {
	n       int
	written int
}

func (w *failOnNthWriter) Write(p []byte) (int, error) {
	w.written++
	if w.written >= w.n {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}

func singleItemWithCoverage() []caseFileItem {
	cov := 85.0
	now := time.Now()
	return []caseFileItem{
		{
			ID: "11111111-1111-1111-1111-111111111111", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "abc1234", Branch: "main", CoveragePercent: &cov,
			TotalFindings: 3, NewFindings: 1, ResolvedFindings: 0, ExistingFindings: 2,
			VerdictOutcome: "pass", CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
			StartedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
		},
	}
}

func TestTrendsDownwardArrow(t *testing.T) {
	covNew := 80.0
	covOld := 90.0
	now := time.Now()
	items := []caseFileItem{
		{
			ID: "11111111-1111-1111-1111-111111111111", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "newer", Branch: "main", CoveragePercent: &covNew,
			TotalFindings: 10, VerdictOutcome: "fail",
			CreatedAt: now.Format(time.RFC3339), StartedAt: now.Format(time.RFC3339),
		},
		{
			ID: "33333333-3333-3333-3333-333333333333", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "older", Branch: "main", CoveragePercent: &covOld,
			TotalFindings: 3, VerdictOutcome: "pass",
			CreatedAt: now.Add(-24 * time.Hour).Format(time.RFC3339), StartedAt: now.Add(-24 * time.Hour).Format(time.RFC3339),
		},
	}
	srv := trendServer(t, items)
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok"})

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "▼")
}

func TestTrendsTimeAgoMinutes(t *testing.T) {
	cov := 80.0
	now := time.Now()
	items := []caseFileItem{
		{
			ID: "11111111-1111-1111-1111-111111111111", ProjectID: "22222222-2222-2222-2222-222222222222",
			CommitSHA: "minutesago", Branch: "main", CoveragePercent: &cov,
			TotalFindings: 0, VerdictOutcome: "pass",
			CreatedAt: now.Add(-5 * time.Minute).Format(time.RFC3339), StartedAt: now.Format(time.RFC3339),
		},
	}
	srv := trendServer(t, items)
	defer srv.Close()

	cmd := trends.NewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok"})

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "5m ago")
}

func TestTrendsWriteErrorOnEmptyHistory(t *testing.T) {
	srv := trendServer(t, nil)
	defer srv.Close()

	cmd := trends.NewCommand()
	cmd.SetOut(failWriter{})
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestTrendsWriteErrorOnEmptyHistoryBranch(t *testing.T) {
	srv := trendServer(t, nil)
	defer srv.Close()

	cmd := trends.NewCommand()
	cmd.SetOut(&failOnNthWriter{n: 2})
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok", "--branch", "main"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestTrendsWriteErrorOnEmptyHistoryRunHint(t *testing.T) {
	srv := trendServer(t, nil)
	defer srv.Close()

	cmd := trends.NewCommand()
	cmd.SetOut(&failOnNthWriter{n: 2})
	cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestTrendsWriteErrorOnTableOutput(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{"header", 1},
		{"column_headers", 2},
		{"separator", 3},
		{"entry_row", 4},
		{"blank_line", 5},
		{"coverage_summary", 6},
		{"findings_summary", 7},
		{"pass_rate", 8},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			srv := trendServer(t, singleItemWithCoverage())
			defer srv.Close()

			cmd := trends.NewCommand()
			cmd.SetOut(&failOnNthWriter{n: testCase.n})
			cmd.SetArgs([]string{"--server", srv.URL, "--project", "core", "--token", "tok"})

			err := cmd.Execute()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "write failed")
		})
	}
}

func TestEnvOrDefaultUsesEnvVar(t *testing.T) {
	t.Setenv("GAVEL_SERVER_URL", "http://env.example.com")

	cmd := trends.NewCommand()

	f := cmd.Flags().Lookup("server")
	require.NotNil(t, f)
	assert.Equal(t, "http://env.example.com", f.DefValue)
}
