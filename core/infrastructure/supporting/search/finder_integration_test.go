package search_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	casefilepostgres "github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"
	searchinfra "github.com/usegavel/gavel/core/infrastructure/supporting/search"
)

var testTenantID = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

func setupDB(t *testing.T) *database.DB {
	testDB := testkit.TestDB(t)
	seedTenant(t, testDB)
	return testDB
}

func seedTenant(t *testing.T, testDB *database.DB) {
	t.Helper()
	_, err := testDB.ExecContext(context.Background(),
		`INSERT INTO iam_tenants (id, slug, display_name, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		testTenantID.UUID(), "test-tenant", "Test Tenant", "active", database.Now())
	require.NoError(t, err)
}

func TestSearch_MatchesProjectByName(t *testing.T) {
	db := setupDB(t)
	seedProject(t, db, "payments", "Payments Service", "//payments/...")

	finder := searchinfra.NewFinder(db)
	results, err := finder.Search(context.Background(), "Payments", 10)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "project", results[0].Type)
	assert.Equal(t, "Payments Service", results[0].Title)
	assert.Equal(t, "payments", results[0].Subtitle)
	assert.Equal(t, "/projects/payments", results[0].URL)
}

func TestSearch_MatchesProjectByKey(t *testing.T) {
	db := setupDB(t)
	seedProject(t, db, "auth-svc", "Auth Service", "//auth/...")

	finder := searchinfra.NewFinder(db)
	results, err := finder.Search(context.Background(), "auth-svc", 10)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "project", results[0].Type)
	assert.Equal(t, "Auth Service", results[0].Title)
}

func TestSearch_MatchesCaseFileByCommitSHA(t *testing.T) {
	db := setupDB(t)
	projectID := seedProject(t, db, "proj-cf", "Project CF", "//cf/...")
	seedCaseFile(t, db, projectID, "abc123def456", "main")

	finder := searchinfra.NewFinder(db)
	results, err := finder.Search(context.Background(), "abc123", 10)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "casefile", results[0].Type)
	assert.Equal(t, "abc123d", results[0].Title)
	assert.Contains(t, results[0].Subtitle, "main")
}

func TestSearch_MatchesCaseFileByBranch(t *testing.T) {
	db := setupDB(t)
	projectID := seedProject(t, db, "proj-br", "Project BR", "//br/...")
	seedCaseFile(t, db, projectID, "deadbeef", "feat/search-test")

	finder := searchinfra.NewFinder(db)
	results, err := finder.Search(context.Background(), "feat/search", 10)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "casefile", results[0].Type)
	assert.Contains(t, results[0].Subtitle, "feat/search-test")
}

func TestSearch_MatchesFindingByRuleID(t *testing.T) {
	db := setupDB(t)
	projectID := seedProject(t, db, "proj-find", "Project Find", "//find/...")
	seedCaseFileWithFindings(t, db, projectID, "aaa111bbb", "main")

	finder := searchinfra.NewFinder(db)
	results, err := finder.Search(context.Background(), "UnusedVariable", 10)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "finding", results[0].Type)
	assert.Equal(t, "UnusedVariable", results[0].Title)
	assert.Contains(t, results[0].Subtitle, "src/Foo.java")
}

func TestSearch_MatchesFindingByFilePath(t *testing.T) {
	db := setupDB(t)
	projectID := seedProject(t, db, "proj-fp", "Project FP", "//fp/...")
	seedCaseFileWithFindings(t, db, projectID, "bbb222ccc", "main")

	finder := searchinfra.NewFinder(db)
	results, err := finder.Search(context.Background(), "src/Bar.java", 10)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "finding", results[0].Type)
	assert.Equal(t, "NP_NULL_DEREF", results[0].Title)
	assert.Contains(t, results[0].Subtitle, "src/Bar.java:25")
}

func TestSearch_ReturnsMultipleTypes(t *testing.T) {
	db := setupDB(t)
	projectID := seedProject(t, db, "multi-search", "Multi Search Service", "//multi/...")
	seedCaseFile(t, db, projectID, "multi0000", "feat/multi-search")

	finder := searchinfra.NewFinder(db)
	results, err := finder.Search(context.Background(), "multi-search", 10)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)

	types := map[string]bool{}
	for _, r := range results {
		types[r.Type] = true
	}
	assert.True(t, types["project"])
	assert.True(t, types["casefile"])
}

func TestSearch_DefaultLimitAppliedWhenZero(t *testing.T) {
	database := setupDB(t)
	for i := range 12 {
		key := "dlimit-" + string(rune('a'+i))
		seedProject(t, database, key, "DLimit Project "+key, "//"+key+"/...")
	}

	finder := searchinfra.NewFinder(database)
	results, err := finder.Search(context.Background(), "DLimit", 0)

	require.NoError(t, err)
	assert.Len(t, results, 10)
}

func TestSearch_DefaultLimitAppliedWhenNegative(t *testing.T) {
	database := setupDB(t)
	for i := range 12 {
		key := "neglim-" + string(rune('a'+i))
		seedProject(t, database, key, "NegLim Project "+key, "//"+key+"/...")
	}

	finder := searchinfra.NewFinder(database)
	results, err := finder.Search(context.Background(), "NegLim", -1)

	require.NoError(t, err)
	assert.Len(t, results, 10)
}

func TestSearch_ExplicitLimitRespected(t *testing.T) {
	database := setupDB(t)
	for i := range 5 {
		key := "elimit-" + string(rune('a'+i))
		seedProject(t, database, key, "ELimit Project "+key, "//"+key+"/...")
	}

	finder := searchinfra.NewFinder(database)
	results, err := finder.Search(context.Background(), "ELimit", 3)

	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestSearch_NoResults(t *testing.T) {
	db := setupDB(t)

	finder := searchinfra.NewFinder(db)
	results, err := finder.Search(context.Background(), "nonexistent-query-xyz", 10)

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_EscapesLikeSpecialCharacters(t *testing.T) {
	database := setupDB(t)
	seedProject(t, database, "esc-pct", "100% Coverage", "//esc/...")

	finder := searchinfra.NewFinder(database)

	t.Run("percent", func(t *testing.T) {
		results, err := finder.Search(context.Background(), "100%", 10)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "100% Coverage", results[0].Title)
	})

	t.Run("underscore_not_wildcard", func(t *testing.T) {
		seedProject(t, database, "esc-under", "A_B_C", "//under/...")
		results, err := finder.Search(context.Background(), "A_B", 10)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "A_B_C", results[0].Title)
	})
}

func TestSearch_ReturnsErrorOnCancelledContext(t *testing.T) {
	db := setupDB(t)

	finder := searchinfra.NewFinder(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := finder.Search(ctx, "anything", 10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search:")
}

func TestSearch_ReturnsErrorWhenScanFails(t *testing.T) {
	database := setupDB(t)
	ctx := context.Background()

	_, err := database.ExecContext(ctx, "ALTER TABLE projects ALTER COLUMN name DROP NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = database.ExecContext(ctx, "DELETE FROM projects WHERE name IS NULL")
		_, _ = database.ExecContext(ctx, "ALTER TABLE projects ALTER COLUMN name SET NOT NULL")
	})

	_, err = database.ExecContext(ctx,
		"INSERT INTO projects (id, key, name, target_pattern) VALUES ($1, $2, NULL, $3)",
		"00000000-0000-0000-0000-ffffffffffff", "null-name-key", "//null/...")
	require.NoError(t, err)

	finder := searchinfra.NewFinder(database)
	_, err = finder.Search(ctx, "null-name", 10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scan search result:")
}

func seedProject(t *testing.T, db *database.DB, key, name, pattern string) projectmodel.ProjectID {
	t.Helper()
	project, err := projectmodel.NewProject(testTenantID, key, name, pattern)
	require.NoError(t, err)
	repo := projectpostgres.NewRepository(db)
	require.NoError(t, repo.Save(context.Background(), project))
	return project.ID()
}

func seedCaseFile(t *testing.T, db *database.DB, projectID projectmodel.ProjectID, commitSHA, branch string) {
	t.Helper()
	now := time.Now().UTC()
	cf, err := casefile.NewCaseFile(testTenantID, projectID, commitSHA, branch, now, now)
	require.NoError(t, err)
	repo := casefilepostgres.NewRepository(db)
	require.NoError(t, repo.Save(context.Background(), cf))
}

func seedCaseFileWithFindings(t *testing.T, database *database.DB, projectID projectmodel.ProjectID, commitSHA, branch string) {
	t.Helper()
	now := time.Now().UTC()
	caseFile, err := casefile.NewCaseFile(testTenantID, projectID, commitSHA, branch, now, now)
	require.NoError(t, err)

	fp1, err := finding.NewFingerprintID("fp-search-aaa")
	require.NoError(t, err)
	fp2, err := finding.NewFingerprintID("fp-search-bbb")
	require.NoError(t, err)

	f1, err := finding.NewFinding("pmd", "UnusedVariable", finding.SeverityWarning, "src/Foo.java", 10, "unused var x", fp1)
	require.NoError(t, err)
	f2, err := finding.NewFinding("spotbugs", "NP_NULL_DEREF", finding.SeverityError, "src/Bar.java", 25, "null deref", fp2)
	require.NoError(t, err)

	fc, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f1, f2})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc, now)
	require.NoError(t, err)
	require.NoError(t, caseFile.AddEvidence(ev, now))

	r1 := verdict.NewRuling(evidence.SubtypeCodeQuality, true, "0 errors")
	v, err := verdict.Compose([]verdict.Ruling{r1}, now)
	require.NoError(t, err)
	caseFile, err = casefile.ReconstituteCaseFile(caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(), caseFile.StartedAt(), caseFile.Evidences(), &v, false)
	require.NoError(t, err)

	repo := casefilepostgres.NewRepository(database)
	require.NoError(t, repo.Save(context.Background(), caseFile))
}
