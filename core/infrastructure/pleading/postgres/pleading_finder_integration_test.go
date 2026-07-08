package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	"github.com/usegavel/gavel/core/domain/pleading/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	casefilepostgres "github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
	pleadingpostgres "github.com/usegavel/gavel/core/infrastructure/pleading/postgres"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"
)

func insertProjectWithKey(t *testing.T, testDB *database.DB, key, name string) projectmodel.Project {
	t.Helper()
	project, err := projectmodel.NewProject(testTenantID, key, name, "//"+key+"/...")
	require.NoError(t, err)
	repo := projectpostgres.NewRepository(testDB)
	require.NoError(t, repo.Save(context.Background(), project))
	return project
}

func insertPleading(t *testing.T, testDB *database.DB, projectID projectmodel.ProjectID, number int, commitSHA string) model.Pleading {
	t.Helper()
	pleading, err := model.FilePleading(testTenantID, projectID, number, fmt.Sprintf("PR #%d", number), "alice", "feature", "main", commitSHA)
	require.NoError(t, err)
	repo := pleadingpostgres.NewRepository(testDB)
	require.NoError(t, repo.Save(context.Background(), pleading))
	return pleading
}

func insertCaseFileWithVerdict(t *testing.T, testDB *database.DB, projectID projectmodel.ProjectID, commitSHA string, passed bool) casefile.CaseFile {
	t.Helper()
	startedAt := time.Now().UTC()
	caseFile, err := casefile.NewCaseFile(testTenantID, projectID, commitSHA, "main", startedAt, startedAt)
	require.NoError(t, err)

	fp, err := finding.NewFingerprintID("fp-" + commitSHA)
	require.NoError(t, err)
	f, err := finding.NewFinding("pmd", "Rule1", finding.SeverityError, "Foo.java", 1, "msg", fp)
	require.NoError(t, err)
	fc, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc, startedAt)
	require.NoError(t, err)
	require.NoError(t, caseFile.AddEvidence(ev, startedAt))

	rulingCodeQuality := verdict.NewRuling(evidence.SubtypeCodeQuality, passed, "1 error found (max 0)")
	rulingCoverage := verdict.NewRuling(evidence.SubtypeCoverage, true, "85.3% coverage (min 80.0%)")
	verdictResult, err := verdict.Compose([]verdict.Ruling{rulingCodeQuality, rulingCoverage}, startedAt)
	require.NoError(t, err)
	caseFile, err = casefile.ReconstituteCaseFile(
		caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(),
		caseFile.StartedAt(), caseFile.Evidences(), &verdictResult, false,
	)
	require.NoError(t, err)

	repo := casefilepostgres.NewRepository(testDB)
	require.NoError(t, repo.Save(context.Background(), caseFile))
	return caseFile
}

func insertGavelspaceWithProject(t *testing.T, testDB *database.DB, gavelspaceName string, projectID projectmodel.ProjectID) {
	t.Helper()
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx,
		"INSERT INTO iam_tenants (id, slug, display_name, status, created_at) VALUES (?, 'local', 'Local', 'active', NOW()) ON CONFLICT DO NOTHING",
		"11111111-1111-1111-1111-111111111111")
	require.NoError(t, err)
	_, err = testDB.ExecContext(ctx,
		"INSERT INTO gavelspaces (name, tenant_id) VALUES (?, ?) ON CONFLICT DO NOTHING", gavelspaceName, "11111111-1111-1111-1111-111111111111")
	require.NoError(t, err)
	_, err = testDB.ExecContext(ctx,
		"INSERT INTO gavelspace_projects (gavelspace_name, project_id, tenant_id) VALUES (?, ?, ?)",
		gavelspaceName, projectID.UUID(), "11111111-1111-1111-1111-111111111111")
	require.NoError(t, err)
}

func TestFinderListByProjectReturnsAll(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	insertPleading(t, testDB, project.ID(), 1, "sha-1")
	insertPleading(t, testDB, project.ID(), 2, "sha-2")

	items, total, err := finder.ListByProject(ctx, testTenantID, project.ID().String(), "", "", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	titles := []string{items[0].Title, items[1].Title}
	assert.Contains(t, titles, "PR #1")
	assert.Contains(t, titles, "PR #2")
}

func TestFinderListByProjectFiltersByStatus(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	open := insertPleading(t, testDB, project.ID(), 1, "sha-1")
	merged := insertPleading(t, testDB, project.ID(), 2, "sha-2")
	require.NoError(t, merged.MarkMerged(time.Now().UTC()))
	repo := pleadingpostgres.NewRepository(testDB)
	require.NoError(t, repo.Save(ctx, merged))

	items, total, err := finder.ListByProject(ctx, testTenantID, project.ID().String(), "open", "", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, open.ID().String(), items[0].ID)
}

func TestFinderListByProjectFiltersByGavelspace(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	ctx := context.Background()

	projectAlpha := insertProjectWithKey(t, testDB, "alpha", "Alpha")
	projectBeta := insertProjectWithKey(t, testDB, "beta", "Beta")

	insertGavelspaceWithProject(t, testDB, "monorepo", projectAlpha.ID())

	insertPleading(t, testDB, projectAlpha.ID(), 1, "sha-a")
	insertPleading(t, testDB, projectBeta.ID(), 1, "sha-b")

	items, total, err := finder.ListByProject(ctx, testTenantID, "", "", "monorepo", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, projectAlpha.ID().String(), items[0].ProjectID)
}

func TestFinderListByProjectWithGateResult(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	insertPleading(t, testDB, project.ID(), 1, "sha-verdict")
	insertCaseFileWithVerdict(t, testDB, project.ID(), "sha-verdict", true)

	items, total, err := finder.ListByProject(ctx, testTenantID, project.ID().String(), "", "", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	require.NotNil(t, items[0].GateResult)
	assert.True(t, items[0].GateResult.Passed)
}

func TestFinderListByProjectWithoutFilters(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	insertPleading(t, testDB, project.ID(), 1, "sha-all")

	items, total, err := finder.ListByProject(ctx, testTenantID, "", "", "", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.GreaterOrEqual(t, len(items), 1)
}

func TestFinderListByProjectPagination(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		insertPleading(t, testDB, project.ID(), i, fmt.Sprintf("sha-page-%d", i))
	}

	items, total, err := finder.ListByProject(ctx, testTenantID, project.ID().String(), "", "", 2, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, items, 2)

	items2, total2, err := finder.ListByProject(ctx, testTenantID, project.ID().String(), "", "", 2, 2)
	require.NoError(t, err)
	assert.Equal(t, total, total2)
	assert.Len(t, items2, 1)
}

func TestFinderListByProjectReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := finder.ListByProject(ctx, testTenantID, "", "", "", 10, 0)
	assert.Error(t, err)
}

func TestFinderGetByIDReturnsDetail(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	pleading := insertPleading(t, testDB, project.ID(), 10, "sha-detail")

	detail, err := finder.GetByID(ctx, testTenantID, pleading.ID().String())
	require.NoError(t, err)
	assert.Equal(t, pleading.ID().String(), detail.ID)
	assert.Equal(t, project.ID().String(), detail.ProjectID)
	assert.Equal(t, 10, detail.Number)
	assert.Equal(t, "PR #10", detail.Title)
	assert.Equal(t, "alice", detail.Petitioner)
	assert.Equal(t, "feature", detail.SourceBranch)
	assert.Equal(t, "main", detail.TargetBranch)
	assert.Equal(t, "sha-detail", detail.CommitSHA)
	assert.Equal(t, "open", detail.Status)
	assert.False(t, detail.CreatedAt.IsZero())
	assert.False(t, detail.UpdatedAt.IsZero())
	assert.Nil(t, detail.GateResult)
}

func TestFinderGetByIDWithGateResultAndConditions(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	pleading := insertPleading(t, testDB, project.ID(), 1, "sha-gate")
	insertCaseFileWithVerdict(t, testDB, project.ID(), "sha-gate", false)

	detail, err := finder.GetByID(ctx, testTenantID, pleading.ID().String())
	require.NoError(t, err)

	require.NotNil(t, detail.GateResult)
	assert.False(t, detail.GateResult.Passed)
	require.Len(t, detail.GateResult.Conditions, 2)

	codeQuality := detail.GateResult.Conditions[0]
	assert.Equal(t, "Code quality", codeQuality.Label)
	assert.Equal(t, "<=", codeQuality.Operator)
	assert.Equal(t, "1", codeQuality.Value)
	assert.Equal(t, "0", codeQuality.Threshold)
	assert.False(t, codeQuality.Passed)

	coverageCond := detail.GateResult.Conditions[1]
	assert.Equal(t, "Coverage", coverageCond.Label)
	assert.Equal(t, ">=", coverageCond.Operator)
	assert.Equal(t, "85.3%", coverageCond.Value)
	assert.Equal(t, "80.0%", coverageCond.Threshold)
	assert.True(t, coverageCond.Passed)
}

func TestFinderGetByIDNotFound(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)

	_, err := finder.GetByID(context.Background(), testTenantID, "00000000-0000-0000-0000-000000000000")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pleading not found")
}

func TestFinderGetByIDReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)

	pleading := insertPleading(t, testDB, project.ID(), 1, "sha-ctx")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := finder.GetByID(ctx, testTenantID, pleading.ID().String())
	assert.Error(t, err)
}

func TestFinderListByProjectReturnsErrorOnCorruptedCreatedAt(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	pleading := insertPleading(t, testDB, project.ID(), 1, "sha-bad-created")

	_, err := testDB.ExecContext(ctx,
		"UPDATE pleadings SET created_at = 'not-a-date' WHERE id = ?",
		pleading.ID().UUID())
	require.NoError(t, err)

	_, _, err = finder.ListByProject(ctx, testTenantID, project.ID().String(), "", "", 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created_at")
}

func TestFinderListByProjectReturnsErrorOnCorruptedUpdatedAt(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	pleading := insertPleading(t, testDB, project.ID(), 1, "sha-bad-updated")

	_, err := testDB.ExecContext(ctx,
		"UPDATE pleadings SET updated_at = 'not-a-date' WHERE id = ?",
		pleading.ID().UUID())
	require.NoError(t, err)

	_, _, err = finder.ListByProject(ctx, testTenantID, project.ID().String(), "", "", 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updated_at")
}

func TestFinderGetByIDReturnsErrorOnCorruptedCreatedAt(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	pleading := insertPleading(t, testDB, project.ID(), 1, "sha-get-bad-created")

	_, err := testDB.ExecContext(ctx,
		"UPDATE pleadings SET created_at = 'not-a-date' WHERE id = ?",
		pleading.ID().UUID())
	require.NoError(t, err)

	_, err = finder.GetByID(ctx, testTenantID, pleading.ID().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created_at")
}

func TestFinderGetByIDReturnsErrorOnCorruptedUpdatedAt(t *testing.T) {
	testDB := setupDB(t)
	finder := pleadingpostgres.NewPleadingFinder(testDB)
	project := insertTestProject(t, testDB)
	ctx := context.Background()

	pleading := insertPleading(t, testDB, project.ID(), 1, "sha-get-bad-updated")

	_, err := testDB.ExecContext(ctx,
		"UPDATE pleadings SET updated_at = 'not-a-date' WHERE id = ?",
		pleading.ID().UUID())
	require.NoError(t, err)

	_, err = finder.GetByID(ctx, testTenantID, pleading.ID().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updated_at")
}
