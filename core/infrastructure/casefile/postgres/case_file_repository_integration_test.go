package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/finalize"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
	casefilepostgres "github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
)

func TestCaseFileRepoSaveAndFindByID(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	startedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	caseFile := newCaseFileAt(t, project.ID(), "abc123", "main", startedAt)
	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	assert.Equal(t, caseFile.ID(), found.ID())
	assert.Equal(t, project.ID(), found.ProjectID())
	assert.Equal(t, "abc123", found.CommitSHA())
	assert.Equal(t, "main", found.Branch())
	assert.Equal(t, startedAt, found.StartedAt())
}

func TestCaseFileRepoRoundTripPreservesFreshEvaluation(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newCaseFileAt(t, project.ID(), "fresh123", "main", time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC))
	caseFile.MarkFreshEvaluation()
	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	assert.True(t, found.IsFreshEvaluation())

	cf2 := newCaseFileAt(t, project.ID(), "normal456", "main", time.Date(2024, 6, 15, 11, 30, 0, 0, time.UTC))
	require.NoError(t, repo.Save(ctx, cf2))

	found2, err := repo.FindByID(ctx, testTenantID, cf2.ID())
	require.NoError(t, err)
	assert.False(t, found2.IsFreshEvaluation())
}

func TestCaseFileRepoFindByIDNotFound(t *testing.T) {
	db := setupDB(t)
	repo := casefilepostgres.NewRepository(db)

	id := casefile.NewCaseFileID(uuid.New())

	_, err := repo.FindByID(context.Background(), testTenantID, id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCaseFileRepoSaveWithFindingsEvidence(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, found.Evidences(), 1)

	loadedEv := found.Evidences()[0]
	assert.Equal(t, evidence.SubtypeCodeQuality, loadedEv.Subtype())
	assert.Equal(t, "pmd", loadedEv.Source())

	fc, ok := loadedEv.Content().(finding.Content)
	require.True(t, ok)
	require.Len(t, fc.Findings(), 2)

	findings := fc.Findings()
	toolSet := map[string]bool{findings[0].Tool(): true, findings[1].Tool(): true}
	assert.True(t, toolSet["pmd"])
	assert.True(t, toolSet["spotbugs"])
}

func TestCaseFileRepoSaveWithCoverageEvidence(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newCoverageEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, found.Evidences(), 1)

	loadedEv := found.Evidences()[0]
	assert.Equal(t, evidence.SubtypeCoverage, loadedEv.Subtype())

	covContent, ok := loadedEv.Content().(coverage.Content)
	require.True(t, ok)
	assert.Equal(t, 100, covContent.TotalLines())
	assert.Equal(t, 80, covContent.CoveredLines())
	require.Len(t, covContent.ByLanguage(), 1)
	assert.Equal(t, "java", covContent.ByLanguage()[0].Language().String())
	assert.Equal(t, 100, covContent.ByLanguage()[0].TotalLines())
	assert.Equal(t, 80, covContent.ByLanguage()[0].CoveredLines())
}

func TestCaseFileRepoSaveWithArchitectureEvidence(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newArchitectureEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, found.Evidences(), 1)

	loadedEv := found.Evidences()[0]
	assert.Equal(t, evidence.SubtypeArchitecture, loadedEv.Subtype())

	ac, ok := loadedEv.Content().(architecture.Content)
	require.True(t, ok)
	require.Len(t, ac.Violations(), 2)
	assert.Equal(t, "no-domain-to-infra", ac.Violations()[0].Rule())
	assert.Equal(t, "domain.user", ac.Violations()[0].SourcePkg())
	assert.Equal(t, "infra.db", ac.Violations()[0].TargetPkg())
}

func TestCaseFileRepoSaveWithToolExecutionEvidence(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newToolExecutionEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, found.Evidences(), 1)

	loadedEv := found.Evidences()[0]
	assert.Equal(t, evidence.SubtypeToolExecution, loadedEv.Subtype())

	tec, ok := loadedEv.Content().(toolexecution.Content)
	require.True(t, ok)
	require.Len(t, tec.Failures(), 2)

	byTool := map[string]string{}
	for _, failed := range tec.Failures() {
		byTool[failed.Tool()] = failed.Reason()
	}
	assert.Equal(t, "exit code 1: analyzer crashed", byTool["pmd"])
	assert.Equal(t, "timed out after 300s", byTool["spotbugs"])
}

func TestCaseFileRepoSaveWithNewCodeCoverageEvidence(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newNewCodeCoverageEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, found.Evidences(), 1)

	loadedEv := found.Evidences()[0]
	assert.Equal(t, evidence.SubtypeNewCodeCoverage, loadedEv.Subtype())

	ncc, ok := loadedEv.Content().(coverage.PatchContent)
	require.True(t, ok)
	assert.Equal(t, 45, ncc.CoveredLines())
	assert.Equal(t, 50, ncc.CoverableLines())
}

func TestCaseFileRepoSaveWithVerdict(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	evaluatedAt := time.Date(2024, 6, 15, 11, 0, 0, 0, time.UTC)
	verdict := newTestVerdict(t)

	caseFile := newCaseFileAt(t, project.ID(), "abc123", "main", time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC))

	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	reconstituted, err := casefile.ReconstituteCaseFile(caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(), caseFile.StartedAt(), caseFile.Evidences(), &verdict, false)
	require.NoError(t, err)
	_ = evaluatedAt

	require.NoError(t, repo.Save(ctx, reconstituted))

	found, err := repo.FindByID(ctx, testTenantID, reconstituted.ID())
	require.NoError(t, err)

	loadedVerdict, ok := found.Verdict()
	require.True(t, ok)
	assert.Equal(t, "fail", loadedVerdict.Outcome().String())
	require.Len(t, loadedVerdict.Rulings(), 2)
	assert.True(t, loadedVerdict.Rulings()[0].Passed())
	assert.False(t, loadedVerdict.Rulings()[1].Passed())
	assert.Equal(t, "code_quality", loadedVerdict.Rulings()[0].Subtype().String())
	assert.Equal(t, "coverage", loadedVerdict.Rulings()[1].Subtype().String())
}

func TestCaseFileRepoSaveUpdatesExisting(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	require.NoError(t, repo.Save(ctx, caseFile))

	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	assert.Len(t, found.Evidences(), 1)
}

func TestCaseFileRepoFindByProject(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	cf1 := newTestCaseFile(t, project.ID(), "aaa", "main")
	cf2 := newTestCaseFile(t, project.ID(), "bbb", "develop")
	require.NoError(t, repo.Save(ctx, cf1))
	require.NoError(t, repo.Save(ctx, cf2))

	results, err := repo.FindByProject(ctx, project.ID())
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestCaseFileRepoFindLatestByBranch(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	earlier := newCaseFileAt(t, project.ID(), "aaa", "main", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	later := newCaseFileAt(t, project.ID(), "bbb", "main", time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, repo.Save(ctx, earlier))
	require.NoError(t, repo.Save(ctx, later))

	found, err := repo.FindLatestByBranch(ctx, project.ID(), "main")
	require.NoError(t, err)
	assert.Equal(t, later.ID(), found.ID())
}

func TestCaseFileRepoFindLatestByBranchNotFound(t *testing.T) {
	db := setupDB(t)
	repo := casefilepostgres.NewRepository(db)

	_, err := repo.FindLatestByBranch(context.Background(), mustGenerateProjectID(t), "main")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCaseFileRepoFindFingerprintsByBranch(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	fps, err := repo.FindFingerprintIDsByBranch(ctx, project.ID(), "main")
	require.NoError(t, err)
	assert.Len(t, fps, 2)

	values := map[string]bool{}
	for _, fp := range fps {
		values[fp.Value()] = true
	}
	assert.True(t, values["fp-aaa"])
	assert.True(t, values["fp-bbb"])
}

func TestCaseFileRepoSaveIsIdempotentForFindings(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, found.Evidences(), 1)
	fc, matched := found.Evidences()[0].Content().(finding.Content)
	require.True(t, matched)
	require.Len(t, fc.Findings(), 2)

	require.NoError(t, repo.Save(ctx, found))

	reloaded, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, reloaded.Evidences(), 1)
	fc2, ok := reloaded.Evidences()[0].Content().(finding.Content)
	require.True(t, ok)
	assert.Len(t, fc2.Findings(), 2, "findings must not duplicate on re-save")
}

func TestCaseFileRepoSaveIsIdempotentForArchViolations(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newArchitectureEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	require.NoError(t, repo.Save(ctx, caseFile))
	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, found))

	reloaded, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, reloaded.Evidences(), 1)
	ac, ok := reloaded.Evidences()[0].Content().(architecture.Content)
	require.True(t, ok)
	assert.Len(t, ac.Violations(), 2, "arch violations must not duplicate on re-save")
}

func TestCaseFileRepoSaveIsIdempotentForCoverageByLanguage(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newCoverageEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	require.NoError(t, repo.Save(ctx, caseFile))
	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, found))

	reloaded, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, reloaded.Evidences(), 1)
	cc, ok := reloaded.Evidences()[0].Content().(coverage.Content)
	require.True(t, ok)
	assert.Len(t, cc.ByLanguage(), 1, "coverage by language must not duplicate on re-save")
}

func TestCaseFileRepoSaveWithNewEvidencePreservesExistingFindings(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	loaded, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	covEv := newCoverageEvidence(t)
	require.NoError(t, loaded.AddEvidence(covEv, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, loaded))

	reloaded, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	require.Len(t, reloaded.Evidences(), 2)

	for _, e := range reloaded.Evidences() {
		if fc, ok := e.Content().(finding.Content); ok {
			assert.Len(t, fc.Findings(), 2, "findings must not duplicate when adding new evidence")
		}
	}
}

func TestCaseFileRepoSaveReplacesExistingForSameCommit(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	first := newTestCaseFile(t, project.ID(), "same-commit", "main")
	require.NoError(t, repo.Save(ctx, first))

	second := newTestCaseFile(t, project.ID(), "same-commit", "main")
	require.NotEqual(t, first.ID(), second.ID(), "precondition: different UUIDs")
	require.NoError(t, repo.Save(ctx, second), "saving a new casefile for the same project+commit must not fail")

	_, err := repo.FindByID(ctx, testTenantID, first.ID())
	assert.Error(t, err, "old casefile should no longer exist")

	found, err := repo.FindByID(ctx, testTenantID, second.ID())
	require.NoError(t, err)
	assert.Equal(t, "same-commit", found.CommitSHA())
}

func TestCaseFileRepoSaveReplacesExistingWithFindings(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	ctx := context.Background()

	first := newTestCaseFile(t, project.ID(), "same-commit", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, first.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, first))

	second := newTestCaseFile(t, project.ID(), "same-commit", "main")
	require.NoError(t, repo.Save(ctx, second), "replacing a casefile that has findings must not fail")

	found, err := repo.FindByID(ctx, testTenantID, second.ID())
	require.NoError(t, err)
	assert.Equal(t, "same-commit", found.CommitSHA())
}

func TestCaseFileRepoFindFingerprintsByBranchEmpty(t *testing.T) {
	db := setupDB(t)
	repo := casefilepostgres.NewRepository(db)

	fps, err := repo.FindFingerprintIDsByBranch(context.Background(), mustGenerateProjectID(t), "main")
	require.NoError(t, err)
	assert.Nil(t, fps)
}

func TestCaseFileRepoWriteCounters(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	repo := casefilepostgres.NewRepository(db)
	cfFinder := casefilepostgres.NewCaseFileFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	require.NoError(t, repo.Save(ctx, caseFile))

	counters := finalize.Counters{
		NewCount:      5,
		ExistingCount: 3,
		ResolvedCount: 2,
	}
	require.NoError(t, repo.WriteCounters(ctx, caseFile.ID().String(), counters))

	detail, err := cfFinder.GetByID(ctx, testTenantID, caseFile.ID().String())
	require.NoError(t, err)
	assert.Equal(t, 5, detail.NewFindings)
	assert.Equal(t, 3, detail.ExistingFindings)
	assert.Equal(t, 2, detail.ResolvedFindings)
}
