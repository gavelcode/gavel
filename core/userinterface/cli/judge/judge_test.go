package judge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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
	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	gavelspacemodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	casefilememory "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	projectmemory "github.com/usegavel/gavel/core/infrastructure/project/memory"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
)

const testDefaultBranch = "main"

type stubSourceContext struct{}

func (s stubSourceContext) CommitSHA(_ context.Context) (string, error) { return "abc123", nil }
func (s stubSourceContext) Branch(_ context.Context) (string, error)    { return testDefaultBranch, nil }
func (s stubSourceContext) ChangedLines(_ context.Context, _, _ string) (map[string][]int, error) {
	return nil, nil
}

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

func newTestDeps(findingsParser ingestfind.Parser, coverageParser ingestcov.Parser) deps {
	projectRepo := projectmemory.NewProjectRepository()
	caseFileRepo := casefilememory.NewCaseFileRepository()

	findingsHandler := ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": findingsParser})
	coverageHandler := ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": coverageParser})

	classifyHandler := classify.NewHandler(caseFileRepo)
	judgeHandler := corejudge.NewHandler(caseFileRepo, projectRepo)
	createCFHandler := createcasefile.NewHandler(caseFileRepo, projectRepo)
	ingestEvHandler := ingestevidence.NewHandler(caseFileRepo)
	finalizeHandler := finalize.NewHandler(caseFileRepo, projectRepo, classifyHandler, judgeHandler, nil)
	submitHandler := submit.NewHandler(createCFHandler, ingestEvHandler, finalizeHandler)

	return deps{
		findings:         findingsHandler,
		coverage:         coverageHandler,
		submitH:          submitHandler,
		collectEvH:       nil,
		projectRepo:      projectRepo,
		fpSeeder:         caseFileRepo,
		resolveWorkspace: stubWorkspaceResolver,
		source:           stubSourceContext{},
		validate:         nil,
		log:              slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestResolveConfigPath_ExplicitOverride(t *testing.T) {
	got := resolveConfigPath("/custom/gavel.yaml", "/some/workspace")
	assert.Equal(t, "/custom/gavel.yaml", got)
}

func TestResolveConfigPath_DefaultResolvesRelativeToWorkspace(t *testing.T) {
	workspace := t.TempDir()
	gavelDir := filepath.Join(workspace, ".gavel")
	require.NoError(t, os.MkdirAll(gavelDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(gavelDir, "gavel.yaml"), []byte(""), 0o644))

	got := resolveConfigPath("", workspace)
	assert.Equal(t, filepath.Join(workspace, ".gavel", "gavel.yaml"), got)
}

func TestResolveConfigPath_FallsBackToRootYaml(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "gavel.yaml"), []byte(""), 0o644))

	got := resolveConfigPath("", workspace)
	assert.Equal(t, filepath.Join(workspace, "gavel.yaml"), got)
}

func TestResolveGitInfo_OverridesBoth(t *testing.T) {
	sha, branch, err := resolveGitInfo(context.Background(), stubSourceContext{}, "custom-sha", "custom-branch")
	require.NoError(t, err)
	assert.Equal(t, "custom-sha", sha)
	assert.Equal(t, "custom-branch", branch)
}

func TestResolveGitInfo_FallsBackToSource(t *testing.T) {
	sha, branch, err := resolveGitInfo(context.Background(), stubSourceContext{}, "", "")
	require.NoError(t, err)
	assert.Equal(t, "abc123", sha)
	assert.Equal(t, testDefaultBranch, branch)
}

type stubStructureVerifier struct{}

func (s stubStructureVerifier) VerifyStructure(_ string) ([]string, error) { return nil, nil }

var _ SourceContext = stubSourceContext{}
var _ ingestfind.Parser = stubFindingsParser{}
var _ ingestcov.Parser = stubCoverageParser{}

func stubWorkspaceResolver() (string, error) { return "/tmp/workspace", nil }

var _ StructureVerifier = stubStructureVerifier{}

func TestNewCommand_PanicsOnNilVerifier(t *testing.T) {
	dependencies := newTestDeps(stubFindingsParser{}, stubCoverageParser{})
	assert.Panics(t, func() {
		NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, nil, nil, dependencies.projectRepo, dependencies.fpSeeder, stubWorkspaceResolver, dependencies.source, nil, nil, nil, nil)
	})
}

func TestNewCommand_ReturnsCobraCommand(t *testing.T) {
	dependencies := newTestDeps(stubFindingsParser{}, stubCoverageParser{})
	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, nil, nil, dependencies.projectRepo, dependencies.fpSeeder, stubWorkspaceResolver, dependencies.source, stubStructureVerifier{}, nil, nil, nil)

	assert.Equal(t, "judge", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestNewCommand_HasFindingsSourceFlag(t *testing.T) {
	dependencies := newTestDeps(stubFindingsParser{}, stubCoverageParser{})
	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, nil, nil, dependencies.projectRepo, dependencies.fpSeeder, stubWorkspaceResolver, dependencies.source, stubStructureVerifier{}, nil, nil, nil)

	flag := cmd.Flags().Lookup("findings-source")
	require.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
}

func TestNewCommand_HasQuickFlag(t *testing.T) {
	dependencies := newTestDeps(stubFindingsParser{}, stubCoverageParser{})
	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, nil, nil, dependencies.projectRepo, dependencies.fpSeeder, stubWorkspaceResolver, dependencies.source, stubStructureVerifier{}, nil, nil, nil)

	flag := cmd.Flags().Lookup("quick")
	require.NotNil(t, flag)
	assert.Equal(t, "q", flag.Shorthand)
	assert.Equal(t, "false", flag.DefValue)
}

func TestNewCommand_HasProjectFlag(t *testing.T) {
	dependencies := newTestDeps(stubFindingsParser{}, stubCoverageParser{})
	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, nil, nil, dependencies.projectRepo, dependencies.fpSeeder, stubWorkspaceResolver, dependencies.source, stubStructureVerifier{}, nil, nil, nil)

	flag := cmd.Flags().Lookup("project")
	require.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
}

func TestIngestFindings_ReturnsEvidence(t *testing.T) {
	filePath, err := finding.NewFingerprintID("test-fingerprint-001")
	require.NoError(t, err)
	parser := stubFindingsParser{
		results: []ingestfind.Parsed{
			{
				RuleID:        "unused-var",
				Severity:      finding.SeverityWarning,
				FilePath:      "main.go",
				Line:          10,
				Message:       "unused variable",
				FingerprintID: filePath,
			},
		},
	}
	dependencies := newTestDeps(parser, stubCoverageParser{})

	cmd, err := ingestfind.NewCommand([]byte(`{"runs":[]}`), "sarif", "test-tool", "code_quality")
	require.NoError(t, err)

	result, err := dependencies.findings.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.NotEqual(t, evidencedto.Evidence{}, result.Evidence)
}

func TestNewCommand_HasAbsoluteFlag(t *testing.T) {
	dependencies := newTestDeps(stubFindingsParser{}, stubCoverageParser{})
	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, nil, nil, dependencies.projectRepo, dependencies.fpSeeder, stubWorkspaceResolver, dependencies.source, stubStructureVerifier{}, nil, nil, nil)

	flag := cmd.Flags().Lookup("absolute")
	require.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestNewCommand_HasRequireSubmitFlag(t *testing.T) {
	dependencies := newTestDeps(stubFindingsParser{}, stubCoverageParser{})
	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, nil, nil, dependencies.projectRepo, dependencies.fpSeeder, stubWorkspaceResolver, dependencies.source, stubStructureVerifier{}, nil, nil, nil)

	flag := cmd.Flags().Lookup("require-submit")
	require.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestProjectBaseline_LoadsFromAggregate(t *testing.T) {
	project, err := projectmodel.NewProject("backend", "backend", "//backend/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp1", "fp2"}, []string{"rule:a:b"}, nil, nil)

	bl := project.Baseline("main")

	assert.True(t, bl.HasPrevious())
	assert.Equal(t, []string{"fp1", "fp2"}, bl.Fingerprints())
	assert.Equal(t, []string{"rule:a:b"}, bl.ArchIDs())
}

func TestNewCommand_HasOutputSARIFFlag(t *testing.T) {
	dependencies := newTestDeps(stubFindingsParser{}, stubCoverageParser{})
	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, nil, nil, dependencies.projectRepo, dependencies.fpSeeder, stubWorkspaceResolver, dependencies.source, stubStructureVerifier{}, nil, nil, nil)

	flag := cmd.Flags().Lookup("output-sarif")
	require.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
}

func TestEnvOrDefault_UsesEnvVar(t *testing.T) {
	t.Setenv("GAVEL_SERVER_URL", "http://test-server")

	got := envOrDefault("GAVEL_SERVER_URL", "fallback")

	assert.Equal(t, "http://test-server", got)
}

func TestNewCommand_PanicsOnNilWorkspaceResolver(t *testing.T) {
	dependencies := newTestDeps(stubFindingsParser{}, stubCoverageParser{})
	assert.Panics(t, func() {
		NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, nil, nil, dependencies.projectRepo, dependencies.fpSeeder, nil, dependencies.source, stubStructureVerifier{}, nil, nil, nil)
	})
}

type fakeFinder struct {
	gavelspace gavelspacemodel.Gavelspace
	projects   []projectmodel.Project
	err        error
}

func (f fakeFinder) LoadFromConfig(_ string) (gavelspacemodel.Gavelspace, []projectmodel.Project, error) {
	return f.gavelspace, f.projects, f.err
}

type stubFindingsCollector struct{}

func (s stubFindingsCollector) CollectFindings(_ context.Context, _ string, _ []string, _ map[string][]string) ([]evidencedto.Evidence, []collectevidence.RawFile, string, error) {
	return []evidencedto.Evidence{{
		Subtype:     "code_quality",
		Source:      "empty.sarif",
		CollectedAt: time.Date(2025, 6, 20, 10, 0, 0, 0, time.UTC),
	}}, nil, "", nil
}

type stubTargetResolver struct {
	ownerTarget     string
	ownerErr        error
	affectedTargets []string
	affectedErr     error
}

func (s *stubTargetResolver) FindOwnerTarget(_ context.Context, _, _ string) (string, error) {
	return s.ownerTarget, s.ownerErr
}

func (s *stubTargetResolver) FindAffectedTargets(_ context.Context, _ string, _ []string, _ string) ([]string, error) {
	return s.affectedTargets, s.affectedErr
}

func newGavelspace(t *testing.T, name string) gavelspacemodel.Gavelspace {
	t.Helper()
	gs, err := gavelspacemodel.NewGavelspace(tenant.LocalTenantID, name)
	require.NoError(t, err)
	return gs
}

func newProject(t *testing.T, key, pattern string) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject(key, key, pattern)
	require.NoError(t, err)
	return p
}

func newRunDeps(t *testing.T, finder fakeFinder) deps {
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

	collectEvH := collectevidence.NewHandler(
		stubFindingsCollector{}, nil, nil,
		findingsHandler, coverageHandler, nil, nil,
	)

	loadWsH := loadgavelspace.NewHandler(finder)

	for _, p := range finder.projects {
		require.NoError(t, projectRepo.Save(context.Background(), p))
	}

	return deps{
		findings:         findingsHandler,
		coverage:         coverageHandler,
		submitH:          submitHandler,
		collectEvH:       collectEvH,
		loadWorkspace:    loadWsH,
		projectRepo:      projectRepo,
		fpSeeder:         caseFileRepo,
		resolveWorkspace: stubWorkspaceResolver,
		source:           stubSourceContext{},
		validate:         stubStructureVerifier{},
		log:              slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func writeGavelConfig(t *testing.T, workspace string) {
	t.Helper()
	gavelDir := filepath.Join(workspace, ".gavel")
	require.NoError(t, os.MkdirAll(gavelDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(gavelDir, "gavel.yaml"), []byte("name: test\n"), 0o644))
}

func newJudgeCommand(t *testing.T, finder fakeFinder, args ...string) (*bytes.Buffer, error) {
	t.Helper()
	dependencies := newRunDeps(t, finder)

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(append([]string{"--quick"}, args...))

	err := cmd.Execute()
	return &buf, err
}

func TestRun_JSONOutput_SingleProject(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}}

	buf, err := newJudgeCommand(t, finder, "--json")

	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	projects := got["projects"].([]any)
	require.Len(t, projects, 1)
	first := projects[0].(map[string]any)
	assert.Equal(t, "core", first["name"])
	assert.Equal(t, "pass", first["verdict"])
}

func TestRun_TextOutput_SingleProject(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}}

	buf, err := newJudgeCommand(t, finder)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "core")
}

func TestRun_MultipleProjects(t *testing.T) {
	p1 := newProject(t, "core", "//core/...")
	p2 := newProject(t, "server", "//server/...")
	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{p1, p2}}

	buf, err := newJudgeCommand(t, finder)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "core")
	assert.Contains(t, buf.String(), "server")
}

func TestRun_MultipleProjects_JSON(t *testing.T) {
	p1 := newProject(t, "core", "//core/...")
	p2 := newProject(t, "server", "//server/...")
	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{p1, p2}}

	buf, err := newJudgeCommand(t, finder, "--json")

	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	projects := got["projects"].([]any)
	assert.Len(t, projects, 2)
}

func TestRun_WorkspaceResolverError(t *testing.T) {
	dependencies := newRunDeps(t, fakeFinder{})
	dependencies.resolveWorkspace = func() (string, error) { return "", errors.New("no workspace") }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspace")
}

func TestRun_LoadConfigError(t *testing.T) {
	finder := fakeFinder{err: errors.New("bad config")}
	_, err := newJudgeCommand(t, finder)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestRun_GitInfoError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})
	dependencies.source = &failingSource{commitErr: errors.New("no git")}

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no git")
}

func TestRun_OutputSARIF(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}}

	sarifDir := t.TempDir()
	sarifPath := filepath.Join(sarifDir, "report.sarif")

	_, err := newJudgeCommand(t, finder, "--output-sarif="+sarifPath)

	require.NoError(t, err)
	_, statErr := os.Stat(sarifPath)
	assert.NoError(t, statErr)
}

func TestRun_AbsoluteSkipsBaseline(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}}

	_, err := newJudgeCommand(t, finder, "--absolute")

	require.NoError(t, err)
}

func TestRun_ProjectFilter(t *testing.T) {
	p1 := newProject(t, "core", "//core/...")
	p2 := newProject(t, "server", "//server/...")
	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{p1, p2}}

	buf, err := newJudgeCommand(t, finder, "--json", "--project=core")

	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	projects := got["projects"].([]any)
	assert.Len(t, projects, 1)
}

func TestRun_AffectedSetsQuick(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}}

	_, err := newJudgeCommand(t, finder, "--affected")

	require.NoError(t, err)
}

func TestRun_TargetFileResolvesOwner(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})
	dependencies.targetResolver = &stubTargetResolver{ownerTarget: "//core:lib"}

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, dependencies.targetResolver)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--target-file=main.go"})

	err := cmd.Execute()

	require.NoError(t, err)
}

func TestRun_TargetFileResolverError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})
	dependencies.targetResolver = &stubTargetResolver{ownerErr: errors.New("no target")}

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, dependencies.targetResolver)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--target-file=main.go"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "find owner target")
}

func TestRun_AffectedResolvesTargets(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gavelspace := newGavelspace(t, "test")

	changedSource := &failingSource{}
	changedSource.commitErr = nil
	changedSource.branchErr = nil

	dependencies := newRunDeps(t, fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{proj}})
	dependencies.source = &stubSourceContextWithChanges{
		changes: map[string][]int{"core/main.go": {1, 2, 3}},
	}
	dependencies.targetResolver = &stubTargetResolver{
		affectedTargets: []string{"//core:lib", "//core:test"},
	}

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, dependencies.targetResolver)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--affected"})

	err := cmd.Execute()

	require.NoError(t, err)
}

type stubSourceContextWithChanges struct {
	changes map[string][]int
}

func (s *stubSourceContextWithChanges) CommitSHA(_ context.Context) (string, error) {
	return "abc123", nil
}
func (s *stubSourceContextWithChanges) Branch(_ context.Context) (string, error) {
	return testDefaultBranch, nil
}
func (s *stubSourceContextWithChanges) ChangedLines(_ context.Context, _, _ string) (map[string][]int, error) {
	return s.changes, nil
}

func newProjectWithGate(t *testing.T, key, pattern string) projectmodel.Project {
	t.Helper()
	strategy, err := qualitygate.NewCountBySeverity(0, 0, 0)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, strategy)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	p, err := projectmodel.NewProject(key, key, pattern)
	require.NoError(t, err)
	p.UpdateQualityGate(gate, time.Now())
	return p
}

func TestRun_VerdictFail_ReturnsError(t *testing.T) {
	proj := newProjectWithGate(t, "core", "//core/...")
	gs := newGavelspace(t, "test")

	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	findingsParser := stubFindingsParser{
		results: []ingestfind.Parsed{
			{RuleID: "R1", Severity: finding.SeverityError, FilePath: "a.go", Line: 1, Message: "m",
				FingerprintID: mustFP(t, "fp-err-001")},
		},
	}
	findingsHandler := ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": findingsParser})
	dependencies.findings = findingsHandler

	dependencies.collectEvH = collectevidence.NewHandler(
		&findingsCollectorWithData{findings: findingsParser.results, parser: findingsHandler},
		nil, nil, findingsHandler, dependencies.coverage, nil, nil,
	)

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--quick", "--absolute"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrVerdictFail)
}

type findingsCollectorWithData struct {
	findings []ingestfind.Parsed
	parser   *ingestfind.Handler
}

func (f *findingsCollectorWithData) CollectFindings(_ context.Context, _ string, _ []string, _ map[string][]string) ([]evidencedto.Evidence, []collectevidence.RawFile, string, error) {
	cmd, err := ingestfind.NewCommand([]byte(`{"runs":[]}`), "sarif", "test.sarif", "code_quality")
	if err != nil {
		return nil, nil, "", err
	}
	res, err := f.parser.Execute(context.Background(), cmd)
	if err != nil {
		return nil, nil, "", err
	}
	return []evidencedto.Evidence{res.Evidence}, nil, "", nil
}

func mustFP(t *testing.T, value string) finding.FingerprintID {
	t.Helper()
	fp, err := finding.NewFingerprintID(value)
	require.NoError(t, err)
	return fp
}

type countFailWriter struct {
	n   int
	max int
}

func (w *countFailWriter) Write(p []byte) (int, error) {
	if w.n >= w.max {
		return 0, errors.New("write failed")
	}
	w.n++
	return len(p), nil
}

func TestEmitResults_TextWriteError(t *testing.T) {
	results := []pipeline.Result{
		{Name: "core", Verdict: "pass"},
	}
	opts := Options{JSONOutput: false}

	err := emitResults(&countFailWriter{max: 0}, results, opts, time.Now())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestEmitResults_RulingsWriteError(t *testing.T) {
	results := []pipeline.Result{
		{
			Name:    "core",
			Verdict: "pass",
			Rulings: []corejudge.RulingView{
				{Subtype: "code_quality", Passed: true, Detail: "ok"},
			},
		},
	}
	opts := Options{JSONOutput: false}

	err := emitResults(&countFailWriter{max: 4}, results, opts, time.Now())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestEmitResults_FirstRunHintWriteError(t *testing.T) {
	results := []pipeline.Result{
		{Name: "core", Verdict: "pass", FirstRun: true},
	}
	opts := Options{JSONOutput: false}

	err := emitResults(&countFailWriter{max: 4}, results, opts, time.Now())

	require.Error(t, err)
}

func TestEmitResults_ServerWarningWriteError(t *testing.T) {
	results := []pipeline.Result{
		{Name: "core", Verdict: "pass", ServerFailed: true},
	}
	opts := Options{JSONOutput: false}

	err := emitResults(&countFailWriter{max: 4}, results, opts, time.Now())

	require.Error(t, err)
}

func TestEmitResults_BuildWarningWriteError(t *testing.T) {
	results := []pipeline.Result{
		{Name: "core", Verdict: "pass", BuildWarning: "partial failure"},
	}
	opts := Options{JSONOutput: false}

	err := emitResults(&countFailWriter{max: 4}, results, opts, time.Now())

	require.Error(t, err)
}

func TestEmitSARIF_WriteError(t *testing.T) {
	results := []pipeline.Result{
		{RawSARIFDocs: [][]byte{[]byte(`not valid json`)}},
	}
	opts := Options{OutputSARIF: "/dev/null/bad/report.sarif"}

	err := emitSARIF(io.Discard, results, opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write SARIF output")
}

func TestEmitSARIF_ConfirmationWriteError(t *testing.T) {
	sarifDir := t.TempDir()
	sarifPath := filepath.Join(sarifDir, "report.sarif")

	validSARIF := `{"$schema":"https://json.schemastore.org/sarif-2.1.0.json","version":"2.1.0","runs":[]}`
	results := []pipeline.Result{
		{RawSARIFDocs: [][]byte{[]byte(validSARIF)}},
	}
	opts := Options{OutputSARIF: sarifPath, JSONOutput: false}

	err := emitSARIF(failWriter{}, results, opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestExecuteProjects_WriteError(t *testing.T) {
	projectRepo := projectmemory.NewProjectRepository()
	caseFileRepo := casefilememory.NewCaseFileRepository()

	findingsHandler := ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": stubFindingsParser{}})
	coverageHandler := ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": stubCoverageParser{}})
	classifyHandler := classify.NewHandler(caseFileRepo)
	judgeHandler := corejudge.NewHandler(caseFileRepo, projectRepo)
	createCFHandler := createcasefile.NewHandler(caseFileRepo, projectRepo)
	ingestEvHandler := ingestevidence.NewHandler(caseFileRepo)
	finalizeHandler := finalize.NewHandler(caseFileRepo, projectRepo, classifyHandler, judgeHandler, nil)
	submitHandler := submit.NewHandler(createCFHandler, ingestEvHandler, finalizeHandler)

	proj := newProject(t, "core", "//core/...")
	require.NoError(t, projectRepo.Save(context.Background(), proj))

	collectEvH := collectevidence.NewHandler(stubFindingsCollector{}, nil, nil, findingsHandler, coverageHandler, nil, nil)

	dependencies := deps{
		findings:   findingsHandler,
		coverage:   coverageHandler,
		submitH:    submitHandler,
		collectEvH: collectEvH,
		log:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	projects := []loadgavelspace.ProjectView{
		{ID: proj.ID().String(), Name: "core", TargetPattern: "//core/..."},
	}

	_, err := executeProjects(context.Background(), failWriter{}, dependencies, t.TempDir(),
		projects, "abc123", "main", time.Now(), Options{Quick: true})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestRun_ValidateStructureError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})
	dependencies.validate = &issueVerifier{err: errors.New("structure broken")}

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "structure broken")
}

type failingCollector struct {
	err error
}

func (f *failingCollector) CollectFindings(_ context.Context, _ string, _ []string, _ map[string][]string) ([]evidencedto.Evidence, []collectevidence.RawFile, string, error) {
	return nil, nil, "", f.err
}

func TestRun_CollectEvidenceError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})

	dependencies.collectEvH = collectevidence.NewHandler(
		&failingCollector{err: errors.New("collect failed")},
		nil, nil, dependencies.findings, dependencies.coverage, nil, nil,
	)

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.Error(t, err)
}

type buildWarningCollector struct{}

func (b *buildWarningCollector) CollectFindings(_ context.Context, _ string, _ []string, _ map[string][]string) ([]evidencedto.Evidence, []collectevidence.RawFile, string, error) {
	return []evidencedto.Evidence{{
		Subtype:     "code_quality",
		Source:      "empty.sarif",
		CollectedAt: time.Date(2025, 6, 20, 10, 0, 0, 0, time.UTC),
	}}, nil, "partial build failure", nil
}

func TestRun_BuildWarning(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})

	dependencies.collectEvH = collectevidence.NewHandler(
		&buildWarningCollector{},
		nil, nil, dependencies.findings, dependencies.coverage, nil, nil,
	)

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.NoError(t, err)
}

func TestRun_ServerURLFromConfig(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	gs.SetServerConfig(gavelspacemodel.NewServerConfig("http://gavel.example.com", "tok123"))

	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.NoError(t, err)
}

func TestRun_AffectedNoChangedLines(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")

	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})
	dependencies.targetResolver = &stubTargetResolver{}

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, dependencies.targetResolver)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--affected"})

	err := cmd.Execute()

	require.NoError(t, err)
}

type changedLinesErrorSource struct{}

func (s *changedLinesErrorSource) CommitSHA(_ context.Context) (string, error) { return "abc", nil }
func (s *changedLinesErrorSource) Branch(_ context.Context) (string, error) {
	return testDefaultBranch, nil
}
func (s *changedLinesErrorSource) ChangedLines(_ context.Context, _, _ string) (map[string][]int, error) {
	return nil, errors.New("diff failed")
}

func TestRun_AffectedChangedLinesDiffError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")

	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})
	dependencies.source = &changedLinesErrorSource{}
	dependencies.targetResolver = &stubTargetResolver{}

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, dependencies.targetResolver)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--affected"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "changed lines")
}

func TestRun_AffectedFindAffectedError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gavelspace := newGavelspace(t, "test")

	dependencies := newRunDeps(t, fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{proj}})
	dependencies.source = &stubSourceContextWithChanges{
		changes: map[string][]int{"core/main.go": {1, 2}},
	}
	dependencies.targetResolver = &stubTargetResolver{affectedErr: errors.New("resolve failed")}

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, dependencies.targetResolver)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--affected"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "find affected targets")
}

func TestRun_SummaryTableWriteError(t *testing.T) {
	p1 := newProject(t, "core", "//core/...")
	p2 := newProject(t, "server", "//server/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{p1, p2}})

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)
	cmd.SetOut(&countFailWriter{max: 15})
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.Error(t, err)
}

func TestRun_BaselineStatusWriteError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gavelspace := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{proj}})

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)
	cmd.SetOut(&countFailWriter{max: 0})
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.Error(t, err)
}

func TestEmitResults_CoverageSummaryWriteError(t *testing.T) {
	results := []pipeline.Result{
		{Name: "core", Verdict: "pass"},
	}
	opts := Options{JSONOutput: false}

	err := emitResults(&countFailWriter{max: 1}, results, opts, time.Now())

	require.Error(t, err)
}

func TestEmitResults_JudgeVerdictWriteError(t *testing.T) {
	results := []pipeline.Result{
		{Name: "core", Verdict: "pass"},
	}
	opts := Options{JSONOutput: false}

	err := emitResults(&countFailWriter{max: 2}, results, opts, time.Now())

	require.Error(t, err)
}

func TestEmitResults_DeltaSummaryWriteError(t *testing.T) {
	results := []pipeline.Result{
		{Name: "core", Verdict: "pass"},
	}
	opts := Options{JSONOutput: false}

	err := emitResults(&countFailWriter{max: 3}, results, opts, time.Now())

	require.Error(t, err)
}

func TestRun_AffectedResolvesWithScopedTargets(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")

	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})
	dependencies.source = &stubSourceContextWithChanges{
		changes: map[string][]int{"core/pkg/main.go": {1, 2}},
	}
	dependencies.targetResolver = &stubTargetResolver{
		affectedTargets: []string{"//core/pkg:lib", "//core/pkg:test"},
	}

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, dependencies.targetResolver)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--affected"})

	err := cmd.Execute()

	require.NoError(t, err)
}

func TestRun_EmitResultsError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)
	cmd.SetOut(&countFailWriter{max: 4})
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.Error(t, err)
}

func TestRun_EmitSARIFError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--quick", "--output-sarif=/dev/null/bad/path/report.sarif"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write SARIF output")
}

func TestRun_WriteCacheError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})

	dependencies.resolveWorkspace = func() (string, error) { return "/dev/null/bad", nil }

	writeCmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)

	var buf bytes.Buffer
	writeCmd.SetOut(&buf)
	writeCmd.SetErr(&buf)
	writeCmd.SetArgs([]string{"--quick"})

	_ = writeCmd.Execute()
}

func TestRun_JudgeHeaderWriteError(t *testing.T) {
	proj := newProject(t, "core", "//core/...")
	gs := newGavelspace(t, "test")
	dependencies := newRunDeps(t, fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}})

	workspace := t.TempDir()
	writeGavelConfig(t, workspace)
	dependencies.resolveWorkspace = func() (string, error) { return workspace, nil }

	cmd := NewCommand(dependencies.findings, dependencies.coverage, dependencies.submitH, dependencies.collectEvH, dependencies.loadWorkspace,
		dependencies.projectRepo, dependencies.fpSeeder, dependencies.resolveWorkspace, dependencies.source, dependencies.validate, dependencies.log, nil, nil)
	cmd.SetOut(&countFailWriter{max: 1})
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--quick"})

	err := cmd.Execute()

	require.Error(t, err)
}
