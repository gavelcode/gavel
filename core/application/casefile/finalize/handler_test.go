package finalize_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/application/casefile/judge"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

var testTenant = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

func newHandler(cfRepo *fakeCaseFileRepo, projRepo *fakeProjectRepo) *finalize.Handler {
	return finalize.NewHandler(cfRepo, projRepo, classify.NewHandler(cfRepo), judge.NewHandler(cfRepo, projRepo), nil)
}

func newHandlerWithCounterWriter(cfRepo *fakeCaseFileRepo, projRepo *fakeProjectRepo, cw finalize.CounterWriter) *finalize.Handler {
	return finalize.NewHandler(cfRepo, projRepo, classify.NewHandler(cfRepo), judge.NewHandler(cfRepo, projRepo), cw)
}

func TestExecuteDeltaComputedAgainstBaseline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, nil, nil, nil)
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-new", "fp-existing"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-new", "fp-existing"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Delta.NewCount)
	assert.Equal(t, 1, result.Delta.FixedCount)
	assert.Equal(t, 1, result.Delta.ExistingCount)
	assert.True(t, result.Delta.HasPrevious)
	assert.True(t, result.Delta.NewFingerprints["fp-new"])
}

func TestExecuteSurfacesPreviousCoverageFromBaseline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	prevPct := 80.0
	entry, err := projectmodel.NewFileCoverageEntry("a.go", []int{1, 2}, []int{3, 4})
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1"}, nil, &prevPct, []projectmodel.FileCoverageEntry{entry})
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-1"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	require.NotNil(t, result.Delta.PreviousCoveragePercent)
	assert.InDelta(t, 80.0, *result.Delta.PreviousCoveragePercent, 0.001)
	require.Len(t, result.Delta.PreviousFileCoverage, 1)
	assert.Equal(t, "a.go", result.Delta.PreviousFileCoverage[0].FilePath)
	assert.Equal(t, []int{1, 2}, result.Delta.PreviousFileCoverage[0].Covered)
	assert.Equal(t, []int{3, 4}, result.Delta.PreviousFileCoverage[0].Uncovered)
}

func TestExecuteDeltaEmptyWhenNoBaseline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-1"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.False(t, result.Delta.HasPrevious)
	assert.Equal(t, 1, result.Delta.NewCount, "first run: all fingerprints are new")
	assert.Equal(t, 0, result.Delta.FixedCount)
}

func TestExecuteUpdatesBaselineWhenVerdictPasses(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateQualityGate(qualitygate.Gate{}, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithCoverage(t, project.ID(), "main", []string{"fp-1"}, 85.0)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1"}),
		finalize.WithArchIDs([]string{"arch-1"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict.Outcome)

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	assert.True(t, baseline.HasPrevious())
	assert.Equal(t, []string{"fp-1"}, baseline.Fingerprints())
	assert.Equal(t, []string{"arch-1"}, baseline.ArchIDs())
	require.NotNil(t, baseline.CoveragePercent())
	assert.InDelta(t, 85.0, *baseline.CoveragePercent(), 0.001)
}

func TestExecuteSeedsBaselineOnFirstFailure(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	gate := mustFailingGate()
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithCoverage(t, project.ID(), "main", []string{"fp-1", "fp-2"}, 85.0)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1", "fp-2"}),
		finalize.WithArchIDs([]string{"arch-1"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Verdict.Outcome)

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	assert.True(t, baseline.HasPrevious(), "baseline must be seeded on first failure")
	assert.Equal(t, []string{"fp-1", "fp-2"}, baseline.Fingerprints())
	assert.Equal(t, []string{"arch-1"}, baseline.ArchIDs())
	require.NotNil(t, baseline.CoveragePercent())
	assert.InDelta(t, 85.0, *baseline.CoveragePercent(), 0.001)
}

func TestExecuteRatchetsBaselineWhenVerdictFails(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	cov := 75.0
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, []string{"a1", "a-resolved"}, &cov, nil)
	gate := mustFailingGate()
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-existing", "fp-new"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-existing", "fp-new"}),
		finalize.WithArchIDs([]string{"a1", "a-new"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Verdict.Outcome)

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	assert.Equal(t, []string{"fp-existing"}, baseline.Fingerprints(), "resolved items removed, new items NOT added")
	assert.Equal(t, []string{"a1"}, baseline.ArchIDs(), "resolved arch IDs removed, new ones NOT added")
	require.NotNil(t, baseline.CoveragePercent(), "coverage preserved from previous baseline")
	assert.InDelta(t, 75.0, *baseline.CoveragePercent(), 0.001)
}

func TestExecuteArchDeltaIncludedInResult(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithArchDelta(finalize.ArchDeltaInput{
			NewCount:   2,
			FixedCount: 1,
			NewIDs:     map[string]bool{"v1": true, "v2": true},
		}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 2, result.Delta.NewViolationsCount)
	assert.Equal(t, 1, result.Delta.FixedViolationsCount)
	assert.True(t, result.Delta.HasArchPrevious)
}

func TestExecuteQuickSkipsArchBaseline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", nil, []string{"old-arch"}, nil, nil)
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1"}),
		finalize.WithArchIDs([]string{"new-arch"}),
		finalize.WithQuick(true),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict.Outcome)

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	assert.Equal(t, []string{"old-arch"}, baseline.ArchIDs())
	assert.Equal(t, []string{"fp-1"}, baseline.Fingerprints())
}

func TestExecuteQuickPreservesCoverageBaseline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	prevPct := 94.6
	entry, err := projectmodel.NewFileCoverageEntry("a.go", []int{1, 2, 3}, []int{4})
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-existing"}, nil, &prevPct, []projectmodel.FileCoverageEntry{entry})
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1"}),
		finalize.WithQuick(true),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict.Outcome)

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	require.NotNil(t, baseline.CoveragePercent())
	assert.InDelta(t, prevPct, *baseline.CoveragePercent(), 0.001)
	assert.Len(t, baseline.FileCoverage(), 1)
}

func TestExecuteAbsoluteSkipsDeltaAndBaseline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-existing"}, nil, nil, nil)
	gate := mustFailingGate()
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-existing"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-existing"}),
		finalize.WithAbsolute(true),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, "fail", result.Verdict.Outcome, "absolute mode evaluates ALL findings, not just new")
	assert.Equal(t, finalize.Delta{}, result.Delta, "absolute mode should not compute delta")
	assert.Empty(t, projRepo.saved, "absolute mode should not save project (no baseline update)")
}

func TestExecuteAbsoluteDoesNotUpdateBaseline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-1"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1"}),
		finalize.WithAbsolute(true),
	)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	found, _ := projRepo.FindByID(context.Background(), testTenant, project.ID())
	baseline := found.Baseline("main")
	assert.False(t, baseline.HasPrevious(), "absolute mode must not update baseline")
}

func mustProject(t *testing.T, key, name, target string) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject(testTenant, key, name, target)
	require.NoError(t, err)
	return p
}

func mustCaseFileWithFindings(t *testing.T, projectID projectmodel.ProjectID, branch string, fingerprints []string) casefile.CaseFile {
	t.Helper()
	now := time.Now().UTC()
	caseF, err := casefile.NewCaseFile(testTenant, projectID, "abc123", branch, now, now)
	require.NoError(t, err)

	if len(fingerprints) > 0 {
		var findings []finding.Finding
		for _, fp := range fingerprints {
			fpID, err := finding.NewFingerprintID(fp)
			require.NoError(t, err)
			f, err := finding.NewFinding("tool", "rule1", finding.SeverityWarning, "file.go", 1, "msg", fpID)
			require.NoError(t, err)
			findings = append(findings, f)
		}
		content, err := finding.NewContent(evidence.SubtypeCodeQuality, findings)
		require.NoError(t, err)
		ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "test", content, now)
		require.NoError(t, err)
		require.NoError(t, caseF.AddEvidence(ev, now))
	}

	return caseF
}

func mustCaseFileWithCoverage(t *testing.T, projectID projectmodel.ProjectID, branch string, fingerprints []string, coveragePct float64) casefile.CaseFile {
	t.Helper()
	now := time.Now().UTC()
	caseF, err := casefile.NewCaseFile(testTenant, projectID, "abc123", branch, now, now)
	require.NoError(t, err)

	if len(fingerprints) > 0 {
		var findings []finding.Finding
		for _, fp := range fingerprints {
			fpID, err := finding.NewFingerprintID(fp)
			require.NoError(t, err)
			found, err := finding.NewFinding("tool", "rule1", finding.SeverityWarning, "file.go", 1, "msg", fpID)
			require.NoError(t, err)
			findings = append(findings, found)
		}
		content, err := finding.NewContent(evidence.SubtypeCodeQuality, findings)
		require.NoError(t, err)
		ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "test", content, now)
		require.NoError(t, err)
		require.NoError(t, caseF.AddEvidence(ev, now))
	}

	covContent, err := coverage.NewContent(100, int(coveragePct), nil)
	require.NoError(t, err)
	covEv, err := evidence.NewEvidence(evidence.SubtypeCoverage, "test", covContent, now)
	require.NoError(t, err)
	require.NoError(t, caseF.AddEvidence(covEv, now))

	return caseF
}

func TestExecuteMinResolvedFailsWhenBelowThreshold(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
		qualitygate.WithMinResolved(2),
	)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, nil, nil, nil)
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-existing"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Verdict.Outcome)
	require.NotEmpty(t, result.Verdict.Rulings)
	assert.Contains(t, result.Verdict.Rulings[0].Detail, "resolved 1")
	assert.Contains(t, result.Verdict.Rulings[0].Detail, "min 2")
}

func TestExecuteMinResolvedPassesWhenAboveThreshold(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
		qualitygate.WithMinResolved(1),
	)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, nil, nil, nil)
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-existing"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict.Outcome)
}

func TestExecuteCoverageMinDeltaFailsWhenCoverageDrops(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	pct, err := qualitygate.NewMinPercentage(50)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCoverage,
		pct,
		qualitygate.WithMinDelta(0),
	)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	prevCov := 80.0
	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", nil, nil, &prevCov, nil)
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithCoverage(t, project.ID(), "main", nil, 75.0)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String())
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Verdict.Outcome)
	require.NotEmpty(t, result.Verdict.Rulings)
	assert.Contains(t, result.Verdict.Rulings[0].Detail, "coverage delta")
}

func TestExecuteCoverageMinDeltaPassesWhenCoverageSame(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	pct, err := qualitygate.NewMinPercentage(50)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCoverage,
		pct,
		qualitygate.WithMinDelta(0),
	)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	prevCov := 80.0
	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", nil, nil, &prevCov, nil)
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithCoverage(t, project.ID(), "main", nil, 80.0)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String())
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict.Outcome)
}

func TestExecuteDerivesFingerprointsFromEvidenceWhenNotProvided(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, nil, nil, nil)
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-new", "fp-existing"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String())
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Delta.NewCount)
	assert.Equal(t, 1, result.Delta.FixedCount)
	assert.Equal(t, 1, result.Delta.ExistingCount)
	assert.True(t, result.Delta.HasPrevious)

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	assert.True(t, baseline.HasPrevious())
	assert.Equal(t, []string{"fp-existing", "fp-new"}, baseline.Fingerprints())
}

func TestExecuteWritesCountersAfterEvaluation(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()
	counterWr := &fakeCounterWriter{}

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, nil, nil, nil)
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-new", "fp-existing"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandlerWithCounterWriter(cfRepo, projRepo, counterWr)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-new", "fp-existing"}),
	)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	require.NotNil(t, counterWr.lastWritten, "CounterWriter must be called after finalize")
	assert.Equal(t, caseF.ID().String(), counterWr.lastCaseFileID)
	assert.Equal(t, 2, counterWr.lastWritten.FindingsCount)
	assert.True(t, counterWr.lastWritten.HasTracking)
}

func TestExecuteSeedsDefaultBranchBaselineWhenAbsent(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	gate := mustFailingGate()
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithCoverage(t, project.ID(), "feature-x", []string{"fp-1", "fp-2"}, 85.0)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1", "fp-2"}),
		finalize.WithArchIDs([]string{"arch-1"}),
	)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	saved := projRepo.lastSaved()

	featureBL := saved.Baseline("feature-x")
	assert.True(t, featureBL.HasPrevious(), "feature branch baseline must be seeded")

	mainBL := saved.Baseline("main")
	assert.True(t, mainBL.HasPrevious(), "default branch baseline must be seeded from feature branch analysis")
	assert.Equal(t, []string{"fp-1", "fp-2"}, mainBL.Fingerprints())
	assert.Equal(t, []string{"arch-1"}, mainBL.ArchIDs())
	require.NotNil(t, mainBL.CoveragePercent())
	assert.InDelta(t, 85.0, *mainBL.CoveragePercent(), 0.001)
}

func TestExecuteDoesNotSeedDefaultBranchWhenAlreadyPresent(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"original-fp"}, []string{"original-arch"}, nil, nil)
	gate := mustFailingGate()
	project.UpdateQualityGate(gate, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "feature-y", []string{"fp-new"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-new"}),
		finalize.WithArchIDs([]string{"arch-new"}),
	)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	saved := projRepo.lastSaved()
	mainBL := saved.Baseline("main")
	assert.Equal(t, []string{"original-fp"}, mainBL.Fingerprints(), "existing main baseline must not be overwritten")
	assert.Equal(t, []string{"original-arch"}, mainBL.ArchIDs(), "existing main arch IDs must not be overwritten")
}

func TestExecuteWithPrecomputedVerdictUsesProvidedOutcome(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateQualityGate(qualitygate.Gate{}, time.Now())
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-1"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "fail",
			EvaluatedAt: time.Now().UTC(),
		}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Verdict.Outcome, "must use precomputed outcome, not re-evaluate")
}

func TestExecuteWithPrecomputedVerdictComputesDelta(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, nil, nil, nil)
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-new", "fp-existing"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "fail",
			EvaluatedAt: time.Now().UTC(),
		}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.True(t, result.Delta.HasPrevious)
	assert.Equal(t, 1, result.Delta.NewCount)
	assert.Equal(t, 1, result.Delta.FixedCount)
	assert.Equal(t, 1, result.Delta.ExistingCount)
}

func TestExecuteWithPrecomputedVerdictSeedsBaseline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithCoverage(t, project.ID(), "main", []string{"fp-1", "fp-2"}, 85.0)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "fail",
			EvaluatedAt: time.Now().UTC(),
		}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Verdict.Outcome)

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	assert.True(t, baseline.HasPrevious(), "baseline must be seeded on first precomputed failure")
	assert.Equal(t, []string{"fp-1", "fp-2"}, baseline.Fingerprints())
}

func TestExecuteWithPrecomputedVerdictRatchetsBaseline(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	cov := 75.0
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, nil, &cov, nil)
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-existing", "fp-new"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-existing", "fp-new"}),
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "fail",
			EvaluatedAt: time.Now().UTC(),
		}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Verdict.Outcome)

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	assert.Equal(t, []string{"fp-existing"}, baseline.Fingerprints(), "resolved items removed, new items NOT added")
}

func TestExecuteWithPrecomputedVerdictWritesCounters(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()
	counterWr := &fakeCounterWriter{}

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-1", "fp-2"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandlerWithCounterWriter(cfRepo, projRepo, counterWr)

	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "pass",
			EvaluatedAt: time.Now().UTC(),
		}),
	)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	require.NotNil(t, counterWr.lastWritten, "counters must be written for precomputed verdict")
	assert.Equal(t, caseF.ID().String(), counterWr.lastCaseFileID)
	assert.Equal(t, 2, counterWr.lastWritten.FindingsCount)
}

func TestNewCommandRejectsEmptyPrecomputedOutcome(t *testing.T) {
	_, err := finalize.NewCommand(testTenant.String(), "some-id",
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "",
			EvaluatedAt: time.Now().UTC(),
		}),
	)
	assert.Error(t, err)
}

func TestExecuteDeduplicatesFingerprintsAcrossEvidence(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()
	counterWr := &fakeCounterWriter{}

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, nil, nil, nil)
	projRepo.seed(project)

	caseFile := mustCaseFileWithDuplicateFindings(t, project.ID(), "main",
		[]string{"fp-existing", "fp-new"},
		[]string{"fp-existing", "fp-new"},
	)
	require.NoError(t, cfRepo.Save(context.Background(), caseFile))

	handler := newHandlerWithCounterWriter(cfRepo, projRepo, counterWr)

	cmd, err := finalize.NewCommand(testTenant.String(), caseFile.ID().String())
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Delta.NewCount, "fp-new counted once despite appearing in two evidence objects")
	assert.Equal(t, 1, result.Delta.ExistingCount, "fp-existing counted once")
	assert.Equal(t, 1, result.Delta.FixedCount, "fp-resolved no longer present")

	require.NotNil(t, counterWr.lastWritten)
	assert.Equal(t, 2, counterWr.lastWritten.FindingsCount, "deduplicated finding count")
}

func mustCaseFileWithDuplicateFindings(t *testing.T, projectID projectmodel.ProjectID, branch string, fps1, fps2 []string) casefile.CaseFile {
	t.Helper()
	now := time.Now().UTC()
	caseF, err := casefile.NewCaseFile(testTenant, projectID, "abc123", branch, now, now)
	require.NoError(t, err)

	for i, fps := range [][]string{fps1, fps2} {
		var findings []finding.Finding
		for _, fp := range fps {
			fpID, err := finding.NewFingerprintID(fp)
			require.NoError(t, err)
			found, err := finding.NewFinding("tool", "rule1", finding.SeverityWarning, "file.go", i+1, "msg", fpID)
			require.NoError(t, err)
			findings = append(findings, found)
		}
		content, err := finding.NewContent(evidence.SubtypeCodeQuality, findings)
		require.NoError(t, err)
		ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "test", content, now)
		require.NoError(t, err)
		require.NoError(t, caseF.AddEvidence(ev, now))
	}

	return caseF
}

func mustFailingGate() qualitygate.Gate {
	rule, _ := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
	)
	gate, _ := qualitygate.NewGate([]qualitygate.Rule{rule})
	return gate
}

func TestNewHandlerPanicsOnNilDeps(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()
	classifyH := classify.NewHandler(cfRepo)
	judgeH := judge.NewHandler(cfRepo, projRepo)

	assert.Panics(t, func() { finalize.NewHandler(nil, projRepo, classifyH, judgeH, nil) })
	assert.Panics(t, func() { finalize.NewHandler(cfRepo, nil, classifyH, judgeH, nil) })
	assert.Panics(t, func() { finalize.NewHandler(cfRepo, projRepo, nil, judgeH, nil) })
	assert.Panics(t, func() { finalize.NewHandler(cfRepo, projRepo, classifyH, nil, nil) })
}

func TestNewHandlerWithLoggerOption(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()
	classifyH := classify.NewHandler(cfRepo)
	judgeH := judge.NewHandler(cfRepo, projRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)
	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := finalize.NewHandler(cfRepo, projRepo, classifyH, judgeH, nil, finalize.WithLogger(logger))
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String())
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
}

func TestExecuteInvalidCaseFileID(t *testing.T) {
	handler := newHandler(newFakeCaseFileRepo(), newFakeProjectRepo())
	cmd, err := finalize.NewCommand(testTenant.String(), "not-a-uuid")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "case file id")
}

func TestExecuteCaseFileNotFound(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	cfRepo.findErr = errors.New("db error")
	handler := newHandler(cfRepo, newFakeProjectRepo())

	cmd, err := finalize.NewCommand(testTenant.String(), "00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load case file")
}

func TestExecuteProjectNotFound(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	projRepo.store = make(map[string]projectmodel.Project)

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String())
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load project")
}

func TestExecuteClassifyError(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-old"}, nil, nil, nil)
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-1"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	cfRepo.fpErr = errors.New("db timeout")

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String())
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "classify")
}

func TestExecuteCounterWriterError(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()
	counterWr := &fakeCounterWriter{err: errors.New("write failed")}

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandlerWithCounterWriter(cfRepo, projRepo, counterWr)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String())
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err, "counter writer error is logged, not propagated")
	assert.Equal(t, "pass", result.Verdict.Outcome)
}

func TestExecuteBaselineSaveError(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-1"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	projRepo.saveErr = errors.New("disk full")

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err, "baseline save error is logged, not propagated")
	assert.Equal(t, "pass", result.Verdict.Outcome)
}

func TestExecuteBaselineSaveErrorOnUpdate(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	project.UpdateBaseline("main", []string{"fp-old"}, nil, nil, nil)
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", []string{"fp-1"})
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	projRepo.saveErr = errors.New("disk full")

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithFingerprints([]string{"fp-1"}),
	)
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err, "baseline save error on update is logged, not propagated")
	assert.Equal(t, "pass", result.Verdict.Outcome)
}

func TestExecutePrecomputedVerdictInvalidSubtype(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "pass",
			EvaluatedAt: time.Now().UTC(),
			Rulings:     []finalize.RulingInput{{Subtype: "INVALID", Passed: true}},
		}),
	)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ruling subtype")
}

func TestExecutePrecomputedVerdictSaveError(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	caseF := mustCaseFileWithFindings(t, project.ID(), "main", nil)
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	cfRepo.saveErr = errors.New("disk full")

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String(),
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "pass",
			EvaluatedAt: time.Now().UTC(),
		}),
	)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save case file")
}

func TestExecuteExtractsArchIDsFromEvidence(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t, "test", "test", "//test/...")
	projRepo.seed(project)

	now := time.Now().UTC()
	caseF, err := casefile.NewCaseFile(testTenant, project.ID(), "abc", "main", now, now)
	require.NoError(t, err)

	archContent, err := architecture.NewContent([]architecture.Violation{
		mustViolation(t, "layer", "api", "domain", "forbidden import"),
	})
	require.NoError(t, err)
	archEv, err := evidence.NewEvidence(evidence.SubtypeArchitecture, "archtest", archContent, now)
	require.NoError(t, err)
	require.NoError(t, caseF.AddEvidence(archEv, now))
	require.NoError(t, cfRepo.Save(context.Background(), caseF))

	handler := newHandler(cfRepo, projRepo)
	cmd, err := finalize.NewCommand(testTenant.String(), caseF.ID().String())
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict.Outcome)
}

func mustViolation(t *testing.T, rule, sourcePkg, targetPkg, message string) architecture.Violation {
	t.Helper()
	v, err := architecture.NewViolation(rule, sourcePkg, targetPkg, message)
	require.NoError(t, err)
	return v
}
