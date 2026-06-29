package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	dbpkg "github.com/usegavel/gavel/core/infrastructure/platform/database"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func setupDB(t *testing.T) *dbpkg.DB { return testkit.TestDB(t) }

func mustGenerateProjectID(t *testing.T) projectmodel.ProjectID {
	t.Helper()
	id, err := projectmodel.ParseProjectID(uuid.NewString())
	require.NoError(t, err)
	return id
}

func insertTestProject(t *testing.T, db *dbpkg.DB) projectmodel.Project {
	t.Helper()
	project, err := projectmodel.NewProject("test-project", "Test Project", "//test/...")
	require.NoError(t, err)
	repo := projectpostgres.NewRepository(db)
	require.NoError(t, repo.Save(context.Background(), project))
	return project
}

func newTestCaseFile(t *testing.T, projectID projectmodel.ProjectID, commitSHA, branch string) casefile.CaseFile {
	t.Helper()
	return newCaseFileAt(t, projectID, commitSHA, branch, time.Now().UTC())
}

func newCaseFileAt(t *testing.T, projectID projectmodel.ProjectID, commitSHA, branch string, startedAt time.Time) casefile.CaseFile {
	t.Helper()
	caseFile, err := casefile.NewCaseFile(projectID, commitSHA, branch, startedAt, startedAt)
	require.NoError(t, err)
	return caseFile
}

func newFindingsEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	fp1, err := finding.NewFingerprintID("fp-aaa")
	require.NoError(t, err)
	fp2, err := finding.NewFingerprintID("fp-bbb")
	require.NoError(t, err)

	f1, err := finding.NewFinding("pmd", "UnusedVariable", finding.SeverityWarning, "src/Foo.java", 10, "unused var x", fp1)
	require.NoError(t, err)
	f2, err := finding.NewFinding("spotbugs", "NP_NULL_DEREF", finding.SeverityError, "src/Bar.java", 25, "null deref", fp2)
	require.NoError(t, err)

	fc, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f1, f2})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func newCoverageEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	lang, err := coverage.NewLanguage("java")
	require.NoError(t, err)
	lc, err := coverage.NewLanguageStats(lang, 100, 80)
	require.NoError(t, err)
	cc, err := coverage.NewContent(100, 80, []coverage.LanguageStats{lc})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCoverage, "jacoco", cc, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func newArchitectureEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	v1, err := architecture.NewViolation("no-domain-to-infra", "domain.user", "infra.db", "domain imports infrastructure")
	require.NoError(t, err)
	v2, err := architecture.NewViolation("no-circular-deps", "pkg.a", "pkg.b", "circular dependency")
	require.NoError(t, err)
	ac, err := architecture.NewContent([]architecture.Violation{v1, v2})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeArchitecture, "archtest", ac, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func newNewCodeCoverageEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	ncc, err := coverage.NewPatchContent(45, 50)
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeNewCodeCoverage, "diff-cover", ncc, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func newToolExecutionEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	f1, err := toolexecution.NewFailure("pmd", "exit code 1: analyzer crashed")
	require.NoError(t, err)
	f2, err := toolexecution.NewFailure("spotbugs", "timed out after 300s")
	require.NoError(t, err)
	tec, err := toolexecution.NewContent([]toolexecution.Failure{f1, f2})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeToolExecution, "sarif", tec, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func insertGavelspaceProject(t *testing.T, database *dbpkg.DB, gavelspaceName, projectID string) {
	t.Helper()
	ctx := context.Background()
	_, err := database.ExecContext(ctx,
		"INSERT INTO gavelspaces (name) VALUES (?) ON CONFLICT DO NOTHING", gavelspaceName)
	require.NoError(t, err)
	_, err = database.ExecContext(ctx,
		"INSERT INTO gavelspace_projects (gavelspace_name, project_id) VALUES (?, ?)",
		gavelspaceName, projectID)
	require.NoError(t, err)
}

func newTestVerdict(t *testing.T) verdict.Result {
	t.Helper()
	r1 := verdict.NewRuling(evidence.SubtypeCodeQuality, true, "0 errors found")
	r2 := verdict.NewRuling(evidence.SubtypeCoverage, false, "coverage 80% < 90%")
	verdictRes, err := verdict.Compose([]verdict.Ruling{r1, r2}, time.Now().UTC())
	require.NoError(t, err)
	return verdictRes
}
