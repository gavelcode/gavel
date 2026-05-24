package pipeline

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	corejudge "github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/application/casefile/submit"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	casefilememory "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	projectmemory "github.com/usegavel/gavel/core/infrastructure/project/memory"
)

type stubFindingsParser struct {
	results []ingestfind.Parsed
}

func (p stubFindingsParser) Parse(_ context.Context, _ []byte) ([]ingestfind.Parsed, error) {
	return p.results, nil
}

type stubCoverageParser struct {
	result ingestcov.Parsed
}

func (p stubCoverageParser) Parse(_ context.Context, _ []byte) (ingestcov.Parsed, error) {
	return p.result, nil
}

func newTestDeps(t *testing.T) (Deps, *projectmemory.ProjectRepository) {
	t.Helper()
	projectRepo := projectmemory.NewProjectRepository()
	caseFileRepo := casefilememory.NewCaseFileRepository()

	findingsParser := stubFindingsParser{}
	coverageParser := stubCoverageParser{}

	findingsHandler := ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": findingsParser})
	coverageHandler := ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": coverageParser})

	classifyHandler := classify.NewHandler(caseFileRepo)
	judgeHandler := corejudge.NewHandler(caseFileRepo, projectRepo)
	createCFHandler := createcasefile.NewHandler(caseFileRepo, projectRepo)
	ingestEvHandler := ingestevidence.NewHandler(caseFileRepo)
	finalizeHandler := finalize.NewHandler(caseFileRepo, projectRepo, classifyHandler, judgeHandler, nil)
	submitHandler := submit.NewHandler(createCFHandler, ingestEvHandler, finalizeHandler)

	deps := Deps{
		Log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		Submit:   submitHandler,
		Findings: findingsHandler,
		Coverage: coverageHandler,
	}
	return deps, projectRepo
}

func minimalEvidence() evidencedto.Evidence {
	return evidencedto.Evidence{
		Subtype:     "code_quality",
		Source:      "test.sarif",
		CollectedAt: time.Date(2025, 6, 20, 10, 0, 0, 0, time.UTC),
	}
}

func TestRunLocal_PassVerdict(t *testing.T) {
	deps, projectRepo := newTestDeps(t)

	project, err := projectmodel.NewProject("backend", "backend", "//backend/...")
	require.NoError(t, err)
	require.NoError(t, projectRepo.Save(context.Background(), project))

	collected := collectevidence.Result{
		Evidences: []evidencedto.Evidence{minimalEvidence()},
	}
	startedAt := time.Date(2025, 6, 20, 10, 0, 0, 0, time.UTC)

	result, err := RunLocal(
		context.Background(), deps, "/workspace",
		collected, project.ID().String(), "backend", "abc123", "main",
		startedAt, Options{Quick: true},
	)

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict)
	assert.Equal(t, "backend", result.Name)
	assert.Equal(t, "abc123", result.CommitSHA)
	assert.Equal(t, "main", result.Branch)
	assert.Equal(t, startedAt, result.StartedAt)
	assert.True(t, result.FirstRun)
}

func TestRunLocal_WithFindings(t *testing.T) {
	deps, projectRepo := newTestDeps(t)

	project, err := projectmodel.NewProject("backend", "backend", "//backend/...")
	require.NoError(t, err)
	require.NoError(t, projectRepo.Save(context.Background(), project))

	fingerprint, err := finding.NewFingerprintID("fp-001")
	require.NoError(t, err)

	collected := collectevidence.Result{
		Evidences:     []evidencedto.Evidence{minimalEvidence()},
		FindingsCount: 1,
		Findings: []evidencedto.Finding{
			{Tool: "PMD", RuleID: "R1", Severity: "error", FilePath: "a.go", Line: 1, Message: "m", FingerprintID: fingerprint.Value()},
		},
		Fingerprints: []string{fingerprint.Value()},
	}

	result, err := RunLocal(
		context.Background(), deps, "/workspace",
		collected, project.ID().String(), "backend", "abc123", "main",
		time.Now(), Options{Quick: true},
	)

	require.NoError(t, err)
	assert.Equal(t, 1, result.FindingsCount)
	assert.Len(t, result.Findings, 1)
}

func TestRunLocal_WithCoverage(t *testing.T) {
	deps, projectRepo := newTestDeps(t)

	project, err := projectmodel.NewProject("backend", "backend", "//backend/...")
	require.NoError(t, err)
	require.NoError(t, projectRepo.Save(context.Background(), project))

	collected := collectevidence.Result{
		Evidences:  []evidencedto.Evidence{minimalEvidence()},
		CovPercent: 85.5,
		CoverageByFile: []evidencedto.FileCoverage{
			{FilePath: "pkg/a.go", Covered: []int{1, 2, 3}, Uncovered: []int{4}},
		},
	}

	result, err := RunLocal(
		context.Background(), deps, "/workspace",
		collected, project.ID().String(), "backend", "abc123", "main",
		time.Now(), Options{Quick: false},
	)

	require.NoError(t, err)
	assert.Equal(t, 85.5, result.CoveragePercent)
	assert.Len(t, result.CoverageByFile, 1)
}

func TestCollectEvidenceFiles_WithSARIF(t *testing.T) {
	collected := collectevidence.Result{
		RawSARIF: []collectevidence.RawFile{
			{Format: "sarif", Source: "tool.sarif", Data: []byte(`{}`)},
		},
	}

	got := collectEvidenceFiles(collected, "backend")

	require.Len(t, got, 1)
	assert.Equal(t, "sarif", got[0].Format)
	assert.Equal(t, "tool.sarif", got[0].Source)
}

func TestCollectEvidenceFiles_WithSARIFAndLCOV(t *testing.T) {
	collected := collectevidence.Result{
		RawSARIF: []collectevidence.RawFile{
			{Format: "sarif", Source: "tool.sarif", Data: []byte(`{}`)},
		},
		RawLCOV: []byte("SF:a.go\nend_of_record"),
	}

	got := collectEvidenceFiles(collected, "backend")

	require.Len(t, got, 2)
	assert.Equal(t, "sarif", got[0].Format)
	assert.Equal(t, "lcov", got[1].Format)
}

func TestCollectEvidenceFiles_EmptyReturnsEmptyFindingsEvidence(t *testing.T) {
	collected := collectevidence.Result{}

	got := collectEvidenceFiles(collected, "backend")

	require.Len(t, got, 1)
	assert.Equal(t, "sarif", got[0].Format)
	assert.Contains(t, got[0].Source, "backend.empty.sarif")
}

func TestEmptyFindingsEvidence_ProducesValidSARIF(t *testing.T) {
	got := emptyFindingsEvidence("myproject")

	assert.Equal(t, "sarif", got.Format)
	assert.Equal(t, "myproject.empty.sarif", got.Source)
	assert.Contains(t, string(got.Data), `"version":"2.1.0"`)
}

func TestParseEvidence_SARIF(t *testing.T) {
	deps, _ := newTestDeps(t)
	file := evidenceFile{
		Format: "sarif",
		Source: "tool.sarif",
		Data:   []byte(`{"runs":[{"tool":{"driver":{"name":"test"}},"results":[]}]}`),
	}

	got, err := parseEvidence(context.Background(), deps, file)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "code_quality", got.Subtype)
}

func TestParseEvidence_LCOV(t *testing.T) {
	deps, _ := newTestDeps(t)
	file := evidenceFile{
		Format: "lcov",
		Source: "coverage.lcov",
		Data:   []byte("SF:a.go\nDA:1,1\nDA:2,0\nend_of_record\n"),
	}

	got, err := parseEvidence(context.Background(), deps, file)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "coverage", got.Subtype)
}

func TestParseEvidence_SourceReturnsNil(t *testing.T) {
	deps, _ := newTestDeps(t)
	file := evidenceFile{Format: "source", Source: "source.go", Data: nil}

	got, err := parseEvidence(context.Background(), deps, file)

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestParseEvidence_UnknownFormatReturnsError(t *testing.T) {
	deps, _ := newTestDeps(t)
	file := evidenceFile{Format: "xml", Source: "report.xml", Data: []byte("<xml/>")}

	_, err := parseEvidence(context.Background(), deps, file)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown evidence format")
}

func TestRunLocal_SubmitNewCommandError(t *testing.T) {
	deps, _ := newTestDeps(t)

	collected := collectevidence.Result{
		Evidences: []evidencedto.Evidence{minimalEvidence()},
	}

	_, err := RunLocal(
		context.Background(), deps, "/workspace",
		collected, "not-a-uuid", "backend", "abc123", "main",
		time.Now(), Options{Quick: true},
	)

	require.Error(t, err)
}

func TestRunLocal_SubmitExecuteError(t *testing.T) {
	deps, _ := newTestDeps(t)

	collected := collectevidence.Result{
		Evidences: []evidencedto.Evidence{minimalEvidence()},
	}

	_, err := RunLocal(
		context.Background(), deps, "/workspace",
		collected, "550e8400-e29b-41d4-a716-446655440000", "backend", "abc123", "main",
		time.Now(), Options{Quick: true},
	)

	require.Error(t, err)
}

func TestRunProject_NoServerDelegatesToLocal(t *testing.T) {
	deps, projectRepo := newTestDeps(t)

	project, err := projectmodel.NewProject("backend", "backend", "//backend/...")
	require.NoError(t, err)
	require.NoError(t, projectRepo.Save(context.Background(), project))

	result, err := RunProject(
		context.Background(), deps, "/workspace",
		collectevidence.Result{Evidences: []evidencedto.Evidence{minimalEvidence()}},
		project.ID().String(), "backend", "abc123", "main",
		time.Now(), Options{Quick: true},
	)

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict)
	assert.False(t, result.ServerFailed)
}
