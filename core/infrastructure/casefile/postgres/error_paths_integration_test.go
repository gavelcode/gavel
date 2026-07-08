package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	casefilepostgres "github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
)

func TestFindingFinderListByFile(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewFindingFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, err := finder.ListByFile(ctx, testTenantID.String(), caseFile.ID().String(), "src/Foo.java")
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "pmd", items[0].Tool)
	assert.Equal(t, "UnusedVariable", items[0].RuleID)
	assert.Equal(t, "src/Foo.java", items[0].FilePath)
	assert.Equal(t, 10, items[0].Line)
}

func TestFindingFinderListByFileEmpty(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewFindingFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, err := finder.ListByFile(ctx, testTenantID.String(), caseFile.ID().String(), "nonexistent.java")
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestRepoSaveErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Save(ctx, caseFile)
	assert.Error(t, err)
}

func TestRepoFindByIDErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	require.NoError(t, repo.Save(context.Background(), caseFile))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestFileCoverageSaveErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	require.NoError(t, cfRepo.Save(context.Background(), caseFile))

	store := casefilepostgres.NewFileCoverageStore(testDB)
	entries := []evidencedto.FileCoverage{
		{FilePath: "src/main.go", Covered: []int{1, 3, 5}, Uncovered: []int{2, 4}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := store.Save(ctx, caseFile.ID().String(), entries)
	assert.Error(t, err)
}

func TestFileCoverageFetchErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	store := casefilepostgres.NewFileCoverageStore(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.Fetch(ctx, "00000000-0000-0000-0000-000000000000", "src/main.go")
	assert.Error(t, err)
}

func TestRepoSaveReturnsErrorOnFindingsInsertFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "findings-fail", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE findings ADD COLUMN extra TEXT NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE findings DROP COLUMN extra")
	})

	err = repo.Save(ctx, caseFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save evidence")
}

func TestRepoSaveReturnsErrorOnCoverageInsertFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "cov-fail", "main")
	ev := newCoverageEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE coverage_data ADD COLUMN extra TEXT NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE coverage_data DROP COLUMN extra")
	})

	err = repo.Save(ctx, caseFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save evidence")
}

func TestRepoSaveReturnsErrorOnArchViolationsInsertFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "arch-fail", "main")
	ev := newArchitectureEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE architecture_violations ADD COLUMN extra TEXT NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE architecture_violations DROP COLUMN extra")
	})

	err = repo.Save(ctx, caseFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save evidence")
}

func TestRepoSaveReturnsErrorOnRulingsInsertFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "ruling-fail", "main")
	verdictRes := newTestVerdict(t)
	caseFile, err := casefile.ReconstituteCaseFile(
		caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(),
		caseFile.StartedAt(), nil, &verdictRes, false,
	)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE rulings ADD COLUMN extra TEXT NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE rulings DROP COLUMN extra")
	})

	err = repo.Save(ctx, caseFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save rulings")
}

func TestRepoFindByIDReturnsErrorOnCorruptedEvidenceSubtype(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "subtype-corrupt", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"UPDATE evidences SET subtype = 'invalid_subtype' WHERE casefile_id = ?",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestRepoFindByIDReturnsErrorOnCorruptedEvidenceCollectedAt(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "time-corrupt", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"UPDATE evidences SET collected_at = 'not-a-date' WHERE casefile_id = ?",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestRepoFindByIDReturnsErrorOnCorruptedFindingSeverity(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "sev-corrupt", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"UPDATE findings SET severity = 'invalid_severity' WHERE casefile_id = ?",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestRepoFindByIDReturnsErrorOnCorruptedCoverageLanguage(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "lang-corrupt", "main")
	ev := newCoverageEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"UPDATE coverage_by_language SET language = '' WHERE evidence_id IN (SELECT id FROM evidences WHERE casefile_id = ?)",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestRepoFindByIDReturnsErrorOnEvidenceQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "ev-query-fail", "main")
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE evidences RENAME TO evidences_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE evidences_corrupted RENAME TO evidences")
	})

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestRepoFindByIDReturnsErrorOnCorruptedVerdictTime(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "verdict-corrupt", "main")
	verdictRes := newTestVerdict(t)
	caseFile, err := casefile.ReconstituteCaseFile(
		caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(),
		caseFile.StartedAt(), nil, &verdictRes, false,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err = testDB.ExecContext(ctx,
		"UPDATE casefiles SET verdict_evaluated_at = 'not-a-date' WHERE id = ?",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verdict")
}

func TestRepoFindByIDReturnsErrorOnRulingsQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "rulings-fail", "main")
	verdictRes := newTestVerdict(t)
	caseFile, err := casefile.ReconstituteCaseFile(
		caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(),
		caseFile.StartedAt(), nil, &verdictRes, false,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE rulings RENAME TO rulings_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE rulings_corrupted RENAME TO rulings")
	})

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestFindingFinderListReturnsErrorOnDataQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewFindingFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "query-fail", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE findings RENAME COLUMN severity TO severity_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE findings RENAME COLUMN severity_corrupted TO severity")
	})

	filters := findinglist.Filters{ProjectID: project.ID().String()}
	_, _, err = finder.List(ctx, testTenantID, filters, 10, 0)
	assert.Error(t, err)
}

func TestFindingFinderListByFileReturnsErrorOnQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	finder := casefilepostgres.NewFindingFinder(testDB)
	ctx := context.Background()

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE findings RENAME TO findings_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE findings_corrupted RENAME TO findings")
	})

	_, err = finder.ListByFile(ctx, testTenantID.String(), "00000000-0000-0000-0000-000000000000", "src/Main.java")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list findings by file")
}

func TestCaseFileFinderListByProjectReturnsErrorOnQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "finder-fail", "main")
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE casefiles RENAME COLUMN commit_sha TO commit_sha_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE casefiles RENAME COLUMN commit_sha_corrupted TO commit_sha")
	})

	_, _, err = finder.ListByProject(ctx, testTenantID, project.ID().String(), "", 10, 0)
	assert.Error(t, err)
}

func TestCaseFileFinderGetByIDReturnsErrorOnEvidenceQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "get-ev-fail", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE evidences RENAME COLUMN subtype TO subtype_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE evidences RENAME COLUMN subtype_corrupted TO subtype")
	})

	_, err = finder.GetByID(ctx, testTenantID, caseFile.ID().String())
	assert.Error(t, err)
}

func TestRepoFindByProjectReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)

	caseFile := newTestCaseFile(t, project.ID(), "proj-cancel", "main")
	require.NoError(t, repo.Save(context.Background(), caseFile))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindByProject(ctx, project.ID())
	assert.Error(t, err)
}

func TestRepoFindByProjectReturnsErrorOnCorruptedStartedAt(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "started-corrupt", "main")
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"UPDATE casefiles SET started_at = 'not-a-date' WHERE id = ?",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByProject(ctx, project.ID())
	assert.Error(t, err)
}

func TestRepoFindFingerprintsByBranchReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindFingerprintIDsByBranch(ctx, project.ID(), "main")
	assert.Error(t, err)
}

func TestRepoFindLatestByBranchReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindLatestByBranch(ctx, project.ID(), "main")
	assert.Error(t, err)
}

func TestFindingFinderListReturnsErrorOnCountFailure(t *testing.T) {
	testDB := setupDB(t)
	finder := casefilepostgres.NewFindingFinder(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := finder.List(ctx, testTenantID, findinglist.Filters{}, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count findings")
}

func TestRepoSaveReturnsErrorOnEvidenceInsertFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "ev-insert-fail", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE evidences ADD COLUMN extra TEXT NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE evidences DROP COLUMN extra")
	})

	err = repo.Save(ctx, caseFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save evidence")
}

func TestRepoFindByIDReturnsErrorOnFindingsQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "findings-query", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE findings RENAME TO findings_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE findings_corrupted RENAME TO findings")
	})

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestRepoFindByIDReturnsErrorOnCoverageQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "cov-query", "main")
	ev := newCoverageEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE coverage_by_language RENAME TO coverage_by_language_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE coverage_by_language_corrupted RENAME TO coverage_by_language")
	})

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestRepoFindByIDReturnsErrorOnArchViolationsQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "arch-query", "main")
	ev := newArchitectureEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE architecture_violations RENAME TO architecture_violations_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE architecture_violations_corrupted RENAME TO architecture_violations")
	})

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestRepoFindByIDReturnsErrorOnNewCodeCoverageQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "ncc-query", "main")
	ev := newNewCodeCoverageEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE new_code_coverage_data RENAME TO new_code_coverage_data_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE new_code_coverage_data_corrupted RENAME TO new_code_coverage_data")
	})

	_, err = repo.FindByID(ctx, testTenantID, caseFile.ID())
	assert.Error(t, err)
}

func TestCaseFileFinderListByProjectWithNoFiltersCoversEmptyConditions(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "no-filter", "main")
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := finder.ListByProject(ctx, testTenantID, "", "", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.GreaterOrEqual(t, len(items), 1)
}

func TestCaseFileFinderListByProjectWithGavelspaceFilter(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	insertGavelspaceProject(t, testDB, "test-gs", project.ID().String())
	caseFile := newTestCaseFile(t, project.ID(), "gs-filter", "main")
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := finder.ListByProject(ctx, testTenantID, project.ID().String(), "test-gs", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.GreaterOrEqual(t, len(items), 1)
}

func TestCaseFileFinderListByProjectShowsCoveragePercent(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "cov-pct", "main")
	ev := newCoverageEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, _, err := finder.ListByProject(ctx, testTenantID, project.ID().String(), "", 10, 0)
	require.NoError(t, err)
	require.NotEmpty(t, items)
	assert.NotNil(t, items[0].CoveragePercent)
}

func TestCaseFileFinderListByProjectReturnsErrorOnCorruptedStartedAt(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "scan-corrupt", "main")
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"UPDATE casefiles SET started_at = 'bad-date' WHERE id = ?",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, _, err = finder.ListByProject(ctx, testTenantID, project.ID().String(), "", 10, 0)
	assert.Error(t, err)
}

func TestCaseFileFinderGetByIDReturnsErrorOnCorruptedStartedAt(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "get-started", "main")
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"UPDATE casefiles SET started_at = 'bad-date' WHERE id = ?",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, err = finder.GetByID(ctx, testTenantID, caseFile.ID().String())
	assert.Error(t, err)
}

func TestCaseFileFinderGetByIDReturnsErrorOnCorruptedCreatedAt(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "get-created", "main")
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"UPDATE casefiles SET created_at = 'bad-date' WHERE id = ?",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, err = finder.GetByID(ctx, testTenantID, caseFile.ID().String())
	assert.Error(t, err)
}

func TestCaseFileFinderGetByIDReturnsErrorOnRulingsQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "get-rulings", "main")
	verdictRes := newTestVerdict(t)
	caseFile, err := casefile.ReconstituteCaseFile(
		caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(),
		caseFile.StartedAt(), nil, &verdictRes, false,
	)
	require.NoError(t, err)
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE rulings RENAME TO rulings_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE rulings_corrupted RENAME TO rulings")
	})

	_, err = finder.GetByID(ctx, testTenantID, caseFile.ID().String())
	assert.Error(t, err)
}

func TestCaseFileFinderGetByIDReturnsErrorOnEvidenceCollectedAtCorruption(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "ev-time", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"UPDATE evidences SET collected_at = 'bad-date' WHERE casefile_id = ?",
		caseFile.ID().UUID())
	require.NoError(t, err)

	_, err = finder.GetByID(ctx, testTenantID, caseFile.ID().String())
	assert.Error(t, err)
}

func TestFileCoverageSaveReturnsErrorOnInsertFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	store := casefilepostgres.NewFileCoverageStore(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "fc-insert", "main")
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE casefile_file_coverage ADD COLUMN extra TEXT NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE casefile_file_coverage DROP COLUMN extra")
	})

	entries := []evidencedto.FileCoverage{
		{FilePath: "src/main.go", Covered: []int{1, 2, 3}, Uncovered: []int{4, 5}},
	}
	err = store.Save(ctx, caseFile.ID().String(), entries)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upsert file coverage")
}

func TestCaseFileFinderGetByIDReturnsNotFoundForMissingCaseFile(t *testing.T) {
	testDB := setupDB(t)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	_, err := finder.GetByID(ctx, testTenantID, "00000000-0000-0000-0000-999999999999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCaseFileFinderGetByIDReturnsErrorOnEvidenceSourceCorruption(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "ev-source-fail", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE evidences RENAME COLUMN source TO source_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE evidences RENAME COLUMN source_corrupted TO source")
	})

	_, err = finder.GetByID(ctx, testTenantID, caseFile.ID().String())
	assert.Error(t, err)
}

func TestCaseFileFinderGetByIDReturnsErrorOnRulingsScanCorruption(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "rul-scan", "main")
	verdictRes := newTestVerdict(t)
	caseFile, err := casefile.ReconstituteCaseFile(
		caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(),
		caseFile.StartedAt(), nil, &verdictRes, false,
	)
	require.NoError(t, err)
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE rulings RENAME COLUMN detail TO detail_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE rulings RENAME COLUMN detail_corrupted TO detail")
	})

	_, err = finder.GetByID(ctx, testTenantID, caseFile.ID().String())
	assert.Error(t, err)
}

func TestCaseFileFinderListByProjectReturnsErrorOnScanCorruption(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	cfRepo := casefilepostgres.NewRepository(testDB)
	finder := casefilepostgres.NewCaseFileFinder(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "list-scan", "main")
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE casefiles RENAME COLUMN verdict_outcome TO verdict_outcome_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE casefiles RENAME COLUMN verdict_outcome_corrupted TO verdict_outcome")
	})

	_, _, err = finder.ListByProject(ctx, testTenantID, project.ID().String(), "", 10, 0)
	assert.Error(t, err)
}

func TestCaseFileFinderListByProjectReturnsErrorOnCountFailure(t *testing.T) {
	testDB := setupDB(t)
	finder := casefilepostgres.NewCaseFileFinder(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := finder.ListByProject(ctx, testTenantID, "", "", 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count casefiles")
}

func TestCaseFileFinderGetByIDReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	finder := casefilepostgres.NewCaseFileFinder(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := finder.GetByID(ctx, testTenantID, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

func TestRepoSaveReturnsErrorOnUpsertCaseFileFailure(t *testing.T) {
	testDB := setupDB(t)
	project := insertTestProject(t, testDB)
	repo := casefilepostgres.NewRepository(testDB)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "upsert-fail", "main")

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE casefiles RENAME COLUMN commit_sha TO commit_sha_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE casefiles RENAME COLUMN commit_sha_corrupted TO commit_sha")
	})

	err = repo.Save(ctx, caseFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upsert casefile")
}
