package appintegration_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	"github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/application/casefile/submit"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestCaseFileSubmitChainFullPipeline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateQualityGate(qualitygate.Gate{}, time.Now())
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	evidences := []evidencedto.Evidence{
		buildFindingEvidence("fp-1", "fp-2"),
		buildCoverageEvidence(100, 85),
	}
	cmd := mustSubmitCommand(t, project.ID().String(), "abc123", "main", evidences,
		[]string{"fp-1", "fp-2"}, nil, finalize.ArchDeltaInput{}, false, false)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, "pass", result.Verdict.Outcome)
	assert.Equal(t, 2, result.Counters.FindingsCount)
	assert.InDelta(t, 85.0, result.Counters.CoveragePercent, 0.001)
}

func TestCaseFileSubmitChainFailingGate(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	gate := mustZeroToleranceGate(t)
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	evidences := []evidencedto.Evidence{
		buildFindingEvidence("fp-1", "fp-2"),
	}
	cmd := mustSubmitCommand(t, project.ID().String(), "abc123", "main", evidences,
		[]string{"fp-1", "fp-2"}, nil, finalize.ArchDeltaInput{}, false, false)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, "fail", result.Verdict.Outcome)
	require.NotEmpty(t, result.Verdict.Rulings)
	assert.False(t, result.Verdict.Rulings[0].Passed)
	assert.Contains(t, result.Verdict.Rulings[0].Detail, "2")
}

func TestCaseFileSubmitChainBaselineDelta(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateQualityGate(qualitygate.Gate{}, time.Now())
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	firstEvidences := []evidencedto.Evidence{
		buildFindingEvidence("fp-1", "fp-2"),
	}
	firstCmd := mustSubmitCommand(t, project.ID().String(), "abc111", "main", firstEvidences,
		[]string{"fp-1", "fp-2"}, nil, finalize.ArchDeltaInput{}, false, false)

	firstResult, err := handler.Execute(context.Background(), firstCmd)
	require.NoError(t, err)
	assert.Equal(t, "pass", firstResult.Verdict.Outcome)

	secondEvidences := []evidencedto.Evidence{
		buildFindingEvidence("fp-2", "fp-3"),
	}
	secondCmd := mustSubmitCommand(t, project.ID().String(), "abc222", "main", secondEvidences,
		[]string{"fp-2", "fp-3"}, nil, finalize.ArchDeltaInput{}, false, false)

	secondResult, err := handler.Execute(context.Background(), secondCmd)
	require.NoError(t, err)

	assert.True(t, secondResult.Delta.HasPrevious)
	assert.Equal(t, 1, secondResult.Delta.ExistingCount, "fp-2 existed before")
	assert.Equal(t, 1, secondResult.Delta.NewCount, "fp-3 is new")
	assert.Equal(t, 1, secondResult.Delta.FixedCount, "fp-1 was resolved")
}

func TestCaseFileSubmitChainPrecomputedVerdict(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateQualityGate(qualitygate.Gate{}, time.Now())
	projRepo.seed(project)

	createH := createcasefile.NewHandler(cfRepo, projRepo)
	ingestH := ingestevidence.NewHandler(cfRepo)
	finalizeH := newFinalizeHandler(cfRepo, projRepo)

	now := time.Now().UTC()

	createCmd, err := createcasefile.NewCommand(tenant.LocalTenantID.String(), project.ID().String(), "abc123", "main", now)
	require.NoError(t, err)
	createRes, err := createH.Execute(context.Background(), createCmd)
	require.NoError(t, err)

	evidences := []evidencedto.Evidence{
		buildFindingEvidence("fp-1"),
		buildCoverageEvidence(100, 90),
	}
	ingestCmd, err := ingestevidence.NewCommand(tenant.LocalTenantID.String(), createRes.CaseFileID, evidences)
	require.NoError(t, err)
	_, err = ingestH.Execute(context.Background(), ingestCmd)
	require.NoError(t, err)

	finalizeCmd, err := finalize.NewCommand(tenant.LocalTenantID.String(), createRes.CaseFileID,
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "fail",
			EvaluatedAt: now,
		}),
	)
	require.NoError(t, err)
	result, err := finalizeH.Execute(context.Background(), finalizeCmd)
	require.NoError(t, err)

	assert.Equal(t, "fail", result.Verdict.Outcome, "must use precomputed outcome")

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	assert.True(t, baseline.HasPrevious(), "baseline must be seeded from evidence")
	assert.Equal(t, []string{"fp-1"}, baseline.Fingerprints())
}

func TestCaseFileSubmitChainAbsoluteSkipsDelta(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-existing"}, nil, nil, nil)
	project.UpdateQualityGate(qualitygate.Gate{}, time.Now())
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	evidences := []evidencedto.Evidence{
		buildFindingEvidence("fp-existing", "fp-new"),
	}
	cmd := mustSubmitCommand(t, project.ID().String(), "abc123", "main", evidences,
		[]string{"fp-existing", "fp-new"}, nil, finalize.ArchDeltaInput{}, false, true)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, finalize.Delta{}, result.Delta, "absolute mode produces zero delta")
	assert.Empty(t, projRepo.saved, "absolute mode must not save project baseline")
}

func TestCaseFileSubmitChainDoubleFinalizeRejected(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateQualityGate(qualitygate.Gate{}, time.Now())
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	evidences := []evidencedto.Evidence{
		buildFindingEvidence("fp-1"),
	}
	cmd := mustSubmitCommand(t, project.ID().String(), "abc123", "main", evidences,
		[]string{"fp-1"}, nil, finalize.ArchDeltaInput{}, false, false)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict.Outcome)

	finalizeH := newFinalizeHandler(cfRepo, projRepo)
	finalizeCmd, err := finalize.NewCommand(tenant.LocalTenantID.String(), result.CaseFileID)
	require.NoError(t, err)

	_, err = finalizeH.Execute(context.Background(), finalizeCmd)
	require.Error(t, err)
	assert.True(t, errors.Is(err, casefile.ErrAlreadyJudged), "double finalize must fail with ErrAlreadyJudged")
}

func newSubmitHandler(cfRepo *fakeCaseFileRepo, projRepo *fakeProjectRepo) *submit.Handler {
	createH := createcasefile.NewHandler(cfRepo, projRepo)
	ingestH := ingestevidence.NewHandler(cfRepo)
	finalizeH := newFinalizeHandler(cfRepo, projRepo)
	return submit.NewHandler(createH, ingestH, finalizeH)
}

func newFinalizeHandler(cfRepo *fakeCaseFileRepo, projRepo *fakeProjectRepo) *finalize.Handler {
	classifyH := classify.NewHandler(cfRepo)
	judgeH := judge.NewHandler(cfRepo, projRepo)
	return finalize.NewHandler(cfRepo, projRepo, classifyH, judgeH, nil)
}

func mustProject(t *testing.T, key, name, target string) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject(tenant.LocalTenantID, key, name, target)
	require.NoError(t, err)
	return p
}

func mustSubmitCommand(
	t *testing.T,
	projectID, commitSHA, branch string,
	evidences []evidencedto.Evidence,
	fingerprints, archIDs []string,
	archDelta finalize.ArchDeltaInput,
	quick, absolute bool,
) submit.Command {
	t.Helper()
	cmd, err := submit.NewCommand(
		tenant.LocalTenantID.String(), projectID, commitSHA, branch,
		evidences, fingerprints, archIDs,
		archDelta, nil, quick, absolute,
		time.Now().UTC(),
	)
	require.NoError(t, err)
	return cmd
}

func mustZeroToleranceGate(t *testing.T) qualitygate.Gate {
	t.Helper()
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
	)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)
	return gate
}

func buildFindingEvidence(fingerprints ...string) evidencedto.Evidence {
	findings := make([]evidencedto.Finding, 0, len(fingerprints))
	for _, fingerprint := range fingerprints {
		findings = append(findings, evidencedto.Finding{
			Tool:          "test-tool",
			RuleID:        "test-rule",
			Severity:      "error",
			FilePath:      "file.go",
			Line:          1,
			Message:       "test finding",
			FingerprintID: fingerprint,
		})
	}
	return evidencedto.Evidence{
		Subtype:     "code_quality",
		Source:      "test",
		CollectedAt: time.Now().UTC(),
		Findings:    findings,
	}
}

func buildCoverageEvidence(totalLines, coveredLines int) evidencedto.Evidence {
	return evidencedto.Evidence{
		Subtype:     "coverage",
		Source:      "test",
		CollectedAt: time.Now().UTC(),
		Coverage: &evidencedto.Coverage{
			TotalLines:   totalLines,
			CoveredLines: coveredLines,
		},
	}
}

var errNotFound = errors.New("not found")

type fakeCaseFileRepo struct {
	mu      sync.Mutex
	store   map[string]casefile.CaseFile
	fpStore map[string][]finding.FingerprintID
}

func newFakeCaseFileRepo() *fakeCaseFileRepo {
	return &fakeCaseFileRepo{
		store:   make(map[string]casefile.CaseFile),
		fpStore: make(map[string][]finding.FingerprintID),
	}
}

func (r *fakeCaseFileRepo) Save(_ context.Context, cf casefile.CaseFile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[cf.ID().String()] = cf
	return nil
}

func (r *fakeCaseFileRepo) FindByID(_ context.Context, _ tenant.TenantID, id casefile.CaseFileID) (casefile.CaseFile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cf, ok := r.store[id.String()]
	if !ok {
		return casefile.CaseFile{}, errNotFound
	}
	return cf, nil
}

func (r *fakeCaseFileRepo) FindLatestByBranch(_ context.Context, _ projectmodel.ProjectID, _ string) (casefile.CaseFile, error) {
	return casefile.CaseFile{}, errNotFound
}

func (r *fakeCaseFileRepo) FindByProject(_ context.Context, _ projectmodel.ProjectID) ([]casefile.CaseFile, error) {
	return nil, nil
}

func (r *fakeCaseFileRepo) FindFingerprintIDsByBranch(_ context.Context, _ projectmodel.ProjectID, _ string) ([]finding.FingerprintID, error) {
	return nil, nil
}

type fakeProjectRepo struct {
	mu    sync.Mutex
	store map[string]projectmodel.Project
	saved []projectmodel.Project
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{store: make(map[string]projectmodel.Project)}
}

func (r *fakeProjectRepo) seed(p projectmodel.Project) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[p.ID().String()] = p
}

func (r *fakeProjectRepo) Save(_ context.Context, p projectmodel.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[p.ID().String()] = p
	r.saved = append(r.saved, p)
	return nil
}

func (r *fakeProjectRepo) FindByID(_ context.Context, _ tenant.TenantID, id projectmodel.ProjectID) (projectmodel.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.store[id.String()]
	if !ok {
		return projectmodel.Project{}, errNotFound
	}
	return p, nil
}

func (r *fakeProjectRepo) FindByName(_ context.Context, _ tenant.TenantID, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, errNotFound
}

func (r *fakeProjectRepo) FindByKey(_ context.Context, _ tenant.TenantID, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, errNotFound
}

func (r *fakeProjectRepo) lastSaved() projectmodel.Project {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.saved) == 0 {
		return projectmodel.Project{}
	}
	return r.saved[len(r.saved)-1]
}
