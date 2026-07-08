package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
)

type Client struct {
	api   *gen.ClientWithResponses
	token string
}

func New(baseURL, token string) (*Client, error) {
	serverURL := strings.TrimRight(baseURL, "/") + "/api/v1"
	api, err := gen.NewClientWithResponses(serverURL, gen.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return nil
	}))
	if err != nil {
		return nil, fmt.Errorf("create api client: %w", err)
	}
	return &Client{api: api, token: token}, nil
}

type ProjectDetail struct {
	ID            string
	Key           string
	Name          string
	DefaultBranch string
}

func (c *Client) FetchProject(ctx context.Context, projectKey string) (*ProjectDetail, error) {
	resp, err := c.api.GetProjectWithResponse(ctx, gen.ProjectKey(projectKey))
	if err != nil {
		return nil, fmt.Errorf("fetch project: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("fetch project: status %d", resp.StatusCode())
	}
	p := resp.JSON200
	return &ProjectDetail{
		ID:            p.Id.String(),
		Key:           string(p.Key),
		Name:          p.Name,
		DefaultBranch: p.DefaultBranch,
	}, nil
}

func (c *Client) CreateProject(ctx context.Context, key, name, targetPattern string) (*ProjectDetail, error) {
	tp := targetPattern
	body := gen.CreateProjectJSONRequestBody{
		Key:           key,
		Name:          name,
		TargetPattern: &tp,
	}
	resp, err := c.api.CreateProjectWithResponse(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("create project: status %d", resp.StatusCode())
	}
	return c.FetchProject(ctx, key)
}

func (c *Client) OpenCaseFile(ctx context.Context, projectID, commitSHA, branch string, freshEvaluation bool) (string, error) {
	fresh := freshEvaluation
	body := gen.CreateCaseFileJSONRequestBody{
		ProjectId:       mustUUID(projectID),
		CommitSha:       commitSHA,
		Branch:          branch,
		FreshEvaluation: &fresh,
	}
	resp, err := c.api.CreateCaseFileWithResponse(ctx, nil, body)
	if err != nil {
		return "", fmt.Errorf("open case file: %w", err)
	}
	if resp.JSON201 == nil {
		return "", fmt.Errorf("open case file: status %d", resp.StatusCode())
	}
	return resp.JSON201.CaseFileId.String(), nil
}

func (c *Client) IngestCaseFileEvidence(ctx context.Context, caseFileID string, ev gen.IngestEvidenceRequest) (string, error) {
	body := gen.IngestCaseFileEvidenceJSONRequestBody(ev)
	resp, err := c.api.IngestCaseFileEvidenceWithResponse(ctx, mustUUID(caseFileID), body)
	if err != nil {
		return "", fmt.Errorf("ingest evidence: %w", err)
	}
	if resp.JSON201 == nil {
		return "", fmt.Errorf("ingest evidence: status %d", resp.StatusCode())
	}
	return resp.JSON201.EvidenceId.String(), nil
}

type SubmitResult struct {
	CaseFileID string
	Verdict    VerdictResult
	Counters   CountersResult
}

type VerdictResult struct {
	Outcome     string
	Rulings     []RulingResult
	EvaluatedAt time.Time
}

type RulingResult struct {
	Subtype string
	Passed  bool
	Detail  string
}

type CountersResult struct {
	FindingsCount   int
	CoveragePercent float64
	NewCount        int
	ExistingCount   int
	ResolvedCount   int
	HasTracking     bool
}

func (c *Client) FinalizeCaseFileWithVerdict(ctx context.Context, caseFileID string, verdict VerdictResult, counters CountersResult) (*SubmitResult, error) {
	outcome := gen.VerdictOutcome(verdict.Outcome)
	rulings := make([]gen.Ruling, 0, len(verdict.Rulings))
	for _, raw := range verdict.Rulings {
		rulings = append(rulings, gen.Ruling{
			Subtype: raw.Subtype,
			Passed:  raw.Passed,
			Detail:  raw.Detail,
		})
	}
	body := gen.FinalizeCaseFileJSONRequestBody{
		Verdict: gen.Verdict{
			Outcome:     outcome,
			Rulings:     &rulings,
			EvaluatedAt: verdict.EvaluatedAt,
		},
		Counters: gen.AnalysisCounters{
			FindingsCount:   int32(counters.FindingsCount),
			CoveragePercent: counters.CoveragePercent,
			NewCount:        int32(counters.NewCount),
			ExistingCount:   int32(counters.ExistingCount),
			ResolvedCount:   int32(counters.ResolvedCount),
			HasTracking:     counters.HasTracking,
		},
	}
	return c.finalizeCaseFileWithBody(ctx, caseFileID, body)
}

func (c *Client) finalizeCaseFileWithBody(ctx context.Context, caseFileID string, body gen.FinalizeCaseFileJSONRequestBody) (*SubmitResult, error) {
	resp, err := c.api.FinalizeCaseFileWithResponse(ctx, mustUUID(caseFileID), body)
	if err != nil {
		return nil, fmt.Errorf("finalize: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("finalize: status %d", resp.StatusCode())
	}
	raw := resp.JSON200
	result := &SubmitResult{
		CaseFileID: raw.CaseFileId.String(),
		Verdict: VerdictResult{
			Outcome:     string(raw.Verdict.Outcome),
			EvaluatedAt: raw.Verdict.EvaluatedAt,
		},
		Counters: CountersResult{
			FindingsCount:   int(raw.Counters.FindingsCount),
			CoveragePercent: raw.Counters.CoveragePercent,
			NewCount:        int(raw.Counters.NewCount),
			ExistingCount:   int(raw.Counters.ExistingCount),
			ResolvedCount:   int(raw.Counters.ResolvedCount),
			HasTracking:     raw.Counters.HasTracking,
		},
	}
	if raw.Verdict.Rulings != nil {
		for _, rl := range *raw.Verdict.Rulings {
			result.Verdict.Rulings = append(result.Verdict.Rulings, RulingResult{
				Subtype: rl.Subtype,
				Passed:  rl.Passed,
				Detail:  rl.Detail,
			})
		}
	}
	return result, nil
}

type Baseline struct {
	Fingerprints     []string
	ArchViolationIDs []string
	HasPrevious      bool
}

func (c *Client) FetchBaseline(ctx context.Context, projectKey, branch string) (*Baseline, error) {
	params := &gen.GetProjectBaselineParams{Branch: &branch}
	resp, err := c.api.GetProjectBaselineWithResponse(ctx, gen.ProjectKey(projectKey), params)
	if err != nil {
		return nil, fmt.Errorf("fetch baseline: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("fetch baseline: status %d", resp.StatusCode())
	}
	b := resp.JSON200
	return &Baseline{
		Fingerprints:     b.Fingerprints,
		ArchViolationIDs: b.ArchitectureIds,
		HasPrevious:      b.HasPrevious,
	}, nil
}

func (c *Client) FilePleading(ctx context.Context, projectKey string, number int, title, petitioner, sourceBranch, targetBranch, commitSHA string) (string, error) {
	body := gen.FileProjectPleadingJSONRequestBody{
		Number:       int32(number),
		Title:        title,
		Petitioner:   petitioner,
		SourceBranch: sourceBranch,
		TargetBranch: targetBranch,
		CommitSha:    commitSHA,
	}
	resp, err := c.api.FileProjectPleadingWithResponse(ctx, gen.ProjectKey(projectKey), nil, body)
	if err != nil {
		return "", fmt.Errorf("file pleading: %w", err)
	}
	if resp.JSON201 == nil {
		return "", fmt.Errorf("file pleading: status %d", resp.HTTPResponse.StatusCode)
	}
	return resp.JSON201.PleadingId.String(), nil
}

func EvidenceToWire(input evidencedto.Evidence) gen.IngestEvidenceRequest {
	subtype := gen.IngestEvidenceRequestSubtype(input.Subtype)
	out := gen.IngestEvidenceRequest{
		Subtype:     subtype,
		Source:      input.Source,
		CollectedAt: input.CollectedAt,
	}
	if input.ID != "" {
		out.Id = &input.ID
	}
	if input.Findings != nil {
		findings := make([]gen.IngestFinding, 0, len(input.Findings))
		for _, finding := range input.Findings {
			findings = append(findings, gen.IngestFinding{
				Tool:        finding.Tool,
				RuleId:      finding.RuleID,
				Severity:    finding.Severity,
				FilePath:    finding.FilePath,
				Line:        int32(finding.Line),
				Message:     finding.Message,
				Fingerprint: finding.FingerprintID,
			})
		}
		out.Findings = &findings
	}
	if input.Coverage != nil {
		byLang := make([]gen.IngestLanguageStats, 0, len(input.Coverage.ByLanguage))
		for _, l := range input.Coverage.ByLanguage {
			byLang = append(byLang, gen.IngestLanguageStats{
				Language:     l.Language,
				TotalLines:   int32(l.TotalLines),
				CoveredLines: int32(l.CoveredLines),
			})
		}
		cov := gen.IngestCoverage{
			TotalLines:   int32(input.Coverage.TotalLines),
			CoveredLines: int32(input.Coverage.CoveredLines),
			ByLanguage:   &byLang,
		}
		if len(input.Coverage.ByFile) > 0 {
			byFile := make([]gen.IngestFileCoverage, 0, len(input.Coverage.ByFile))
			for _, fc := range input.Coverage.ByFile {
				byFile = append(byFile, gen.IngestFileCoverage{
					FilePath:       fc.FilePath,
					CoveredLines:   toInt32Slice(fc.Covered),
					UncoveredLines: toInt32Slice(fc.Uncovered),
				})
			}
			cov.ByFile = &byFile
		}
		out.Coverage = &cov
	}
	if input.NewCodeCoverage != nil {
		out.NewCodeCoverage = &gen.IngestNewCodeCoverage{
			CoveredLines:   int32(input.NewCodeCoverage.CoveredLines),
			CoverableLines: int32(input.NewCodeCoverage.CoverableLines),
		}
	}
	if input.Architecture != nil {
		violations := make([]gen.IngestViolation, 0, len(input.Architecture.Violations))
		for _, v := range input.Architecture.Violations {
			violations = append(violations, gen.IngestViolation{
				Rule:      v.Rule,
				SourcePkg: v.SourcePkg,
				TargetPkg: v.TargetPkg,
				Message:   v.Message,
			})
		}
		out.Architecture = &gen.IngestArchitecture{Violations: violations}
	}
	return out
}

type TrendEntry struct {
	CommitSHA        string    `json:"commit_sha"`
	Branch           string    `json:"branch"`
	CoveragePercent  *float64  `json:"coverage_percent,omitempty"`
	TotalFindings    int       `json:"total_findings"`
	NewFindings      int       `json:"new_findings"`
	ResolvedFindings int       `json:"resolved_findings"`
	VerdictOutcome   string    `json:"verdict_outcome"`
	CreatedAt        time.Time `json:"created_at"`
}

func (c *Client) ListProjectCaseFiles(ctx context.Context, projectKey string, limit int) ([]TrendEntry, error) {
	lim := gen.Limit(limit)
	params := &gen.ListProjectCaseFilesParams{Limit: &lim}
	resp, err := c.api.ListProjectCaseFilesWithResponse(ctx, gen.ProjectKey(projectKey), params)
	if err != nil {
		return nil, fmt.Errorf("list project casefiles: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("list project casefiles: status %d", resp.StatusCode())
	}
	entries := make([]TrendEntry, 0, len(resp.JSON200.Items))
	for _, caseF := range resp.JSON200.Items {
		entries = append(entries, TrendEntry{
			CommitSHA:        caseF.CommitSha,
			Branch:           caseF.Branch,
			CoveragePercent:  caseF.CoveragePercent,
			TotalFindings:    int(caseF.TotalFindings),
			NewFindings:      int(caseF.NewFindings),
			ResolvedFindings: int(caseF.ResolvedFindings),
			VerdictOutcome:   caseF.VerdictOutcome,
			CreatedAt:        caseF.CreatedAt,
		})
	}
	return entries, nil
}

type SourceFile struct {
	Path    string
	Content string
}

func (c *Client) UploadSource(ctx context.Context, projectKey, commitSHA string, files []SourceFile) error {
	if len(files) == 0 {
		return nil
	}
	genFiles := make([]gen.SourceFile, 0, len(files))
	for _, finding := range files {
		genFiles = append(genFiles, gen.SourceFile{
			Path:    finding.Path,
			Content: finding.Content,
		})
	}
	body := gen.UploadProjectSourceJSONRequestBody{
		Commit: commitSHA,
		Files:  genFiles,
	}
	resp, err := c.api.UploadProjectSourceWithResponse(ctx, gen.ProjectKey(projectKey), body)
	if err != nil {
		return fmt.Errorf("upload source: %w", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("upload source: status %d", resp.StatusCode())
	}
	return nil
}

func toInt32Slice(ints []int) []int32 {
	out := make([]int32, len(ints))
	for i, v := range ints {
		out[i] = int32(v)
	}
	return out
}

func mustUUID(s string) gen.CaseFileID {
	var id gen.CaseFileID
	_ = id.UnmarshalText([]byte(s))
	return id
}
