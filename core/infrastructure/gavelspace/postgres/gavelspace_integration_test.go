package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	casefilepostgres "github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
	gspostgres "github.com/usegavel/gavel/core/infrastructure/gavelspace/postgres"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"
)

func setupDB(t *testing.T) *database.DB { return testkit.TestDB(t) }

func insertProject(t *testing.T, testDB *database.DB, key, name string) projectmodel.Project {
	t.Helper()
	project, err := projectmodel.NewProject(key, name, "//"+key+"/...")
	require.NoError(t, err)
	repo := projectpostgres.NewRepository(testDB)
	require.NoError(t, repo.Save(context.Background(), project))
	return project
}

func insertGavelspaceWithProjects(t *testing.T, testDB *database.DB, name string, projects []projectmodel.Project) gsmodel.Gavelspace {
	t.Helper()
	gavelspace, err := gsmodel.NewGavelspace(name)
	require.NoError(t, err)
	now := time.Now().UTC()
	for _, project := range projects {
		ref, err := gsmodel.NewProjectRef(project.ID(), project.TargetPattern())
		require.NoError(t, err)
		require.NoError(t, gavelspace.AddProject(ref, now))
	}
	repo := gspostgres.NewRepository(testDB)
	require.NoError(t, repo.Save(context.Background(), gavelspace))
	return gavelspace
}

func insertCaseFileWithPassingVerdict(t *testing.T, testDB *database.DB, projectID projectmodel.ProjectID, commitSHA string) {
	t.Helper()
	startedAt := time.Now().UTC()
	caseFile, err := casefile.NewCaseFile(projectID, commitSHA, "main", startedAt, startedAt)
	require.NoError(t, err)

	fp, err := finding.NewFingerprintID("fp-" + commitSHA)
	require.NoError(t, err)
	findingObj, err := finding.NewFinding("pmd", "Rule1", finding.SeverityError, "Foo.java", 1, "msg", fp)
	require.NoError(t, err)
	fc, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{findingObj})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc, startedAt)
	require.NoError(t, err)
	require.NoError(t, caseFile.AddEvidence(ev, startedAt))

	ruling := verdict.NewRuling(evidence.SubtypeCodeQuality, true, "0 errors")
	verdictResult, err := verdict.Compose([]verdict.Ruling{ruling}, startedAt)
	require.NoError(t, err)
	caseFile, err = casefile.ReconstituteCaseFile(
		caseFile.ID(), caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(),
		caseFile.StartedAt(), caseFile.Evidences(), &verdictResult, false,
	)
	require.NoError(t, err)

	repo := casefilepostgres.NewRepository(testDB)
	require.NoError(t, repo.Save(context.Background(), caseFile))
}

func TestRepoSaveAndFindByName(t *testing.T) {
	testDB := setupDB(t)
	repo := gspostgres.NewRepository(testDB)
	ctx := context.Background()

	project := insertProject(t, testDB, "core", "Core")

	gavelspace, err := gsmodel.NewGavelspace("my-monorepo")
	require.NoError(t, err)
	now := time.Now().UTC()
	ref, err := gsmodel.NewProjectRef(project.ID(), project.TargetPattern())
	require.NoError(t, err)
	require.NoError(t, gavelspace.AddProject(ref, now))

	require.NoError(t, repo.Save(ctx, gavelspace))

	found, err := repo.FindByName(ctx, gavelspace.ID())
	require.NoError(t, err)
	assert.Equal(t, gavelspace.ID(), found.ID())
	require.Len(t, found.Projects(), 1)
	assert.Equal(t, project.ID(), found.Projects()[0].ID())
	assert.Equal(t, "//core/...", found.Projects()[0].TargetPattern())
}

func TestRepoSaveUpdatesExistingGavelspace(t *testing.T) {
	testDB := setupDB(t)
	repo := gspostgres.NewRepository(testDB)
	ctx := context.Background()

	projectAlpha := insertProject(t, testDB, "alpha", "Alpha")
	projectBeta := insertProject(t, testDB, "beta", "Beta")

	gavelspace, err := gsmodel.NewGavelspace("evolving")
	require.NoError(t, err)
	now := time.Now().UTC()
	ref, err := gsmodel.NewProjectRef(projectAlpha.ID(), projectAlpha.TargetPattern())
	require.NoError(t, err)
	require.NoError(t, gavelspace.AddProject(ref, now))
	require.NoError(t, repo.Save(ctx, gavelspace))

	ref2, err := gsmodel.NewProjectRef(projectBeta.ID(), projectBeta.TargetPattern())
	require.NoError(t, err)
	require.NoError(t, gavelspace.AddProject(ref2, now))
	require.NoError(t, repo.Save(ctx, gavelspace))

	found, err := repo.FindByName(ctx, gavelspace.ID())
	require.NoError(t, err)
	assert.Len(t, found.Projects(), 2)
}

func TestRepoFindByNameNotFound(t *testing.T) {
	testDB := setupDB(t)
	repo := gspostgres.NewRepository(testDB)

	name, err := gsmodel.NewGavelspaceID("nonexistent")
	require.NoError(t, err)

	_, err = repo.FindByName(context.Background(), name)
	require.Error(t, err)
	assert.Equal(t, failure.NotFound, failure.Of(err))
}

func TestRepoFindByNameReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	repo := gspostgres.NewRepository(testDB)

	insertGavelspaceWithProjects(t, testDB, "ctx-test", nil)

	name, err := gsmodel.NewGavelspaceID("ctx-test")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.FindByName(ctx, name)
	assert.Error(t, err)
}

func TestRepoSaveReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	repo := gspostgres.NewRepository(testDB)

	gavelspace, err := gsmodel.NewGavelspace("will-fail")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.Save(ctx, gavelspace)
	assert.Error(t, err)
}

func TestRepoSaveWithEmptyProjects(t *testing.T) {
	testDB := setupDB(t)
	repo := gspostgres.NewRepository(testDB)
	ctx := context.Background()

	gavelspace, err := gsmodel.NewGavelspace("empty")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, gavelspace))

	found, err := repo.FindByName(ctx, gavelspace.ID())
	require.NoError(t, err)
	assert.Empty(t, found.Projects())
}

func TestFinderList(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)
	ctx := context.Background()

	projectAlpha := insertProject(t, testDB, "alpha", "Alpha")
	projectBeta := insertProject(t, testDB, "beta", "Beta")
	insertGavelspaceWithProjects(t, testDB, "mono-a", []projectmodel.Project{projectAlpha, projectBeta})
	insertGavelspaceWithProjects(t, testDB, "mono-b", nil)

	items, total, err := finder.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, items, 2)

	assert.Equal(t, "mono-a", items[0].Name)
	assert.Equal(t, 2, items[0].ProjectCount)
	assert.False(t, items[0].CreatedAt.IsZero())

	assert.Equal(t, "mono-b", items[1].Name)
	assert.Equal(t, 0, items[1].ProjectCount)
}

func TestFinderListPagination(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)
	ctx := context.Background()

	for index := range 3 {
		insertGavelspaceWithProjects(t, testDB, fmt.Sprintf("page-%d", index), nil)
	}

	items, total, err := finder.List(ctx, 2, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, items, 2)

	items2, _, err := finder.List(ctx, 2, 2)
	require.NoError(t, err)
	assert.Len(t, items2, 1)
}

func TestFinderListReturnsEmptyWhenNoGavelspaces(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)

	items, total, err := finder.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

func TestFinderListReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := finder.List(ctx, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count gavelspaces")
}

func TestFinderGetByName(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)
	ctx := context.Background()

	project := insertProject(t, testDB, "core", "Core")
	insertGavelspaceWithProjects(t, testDB, "detail-test", []projectmodel.Project{project})

	detail, err := finder.GetByName(ctx, "detail-test")
	require.NoError(t, err)
	assert.Equal(t, "detail-test", detail.Name)
	assert.False(t, detail.CreatedAt.IsZero())
	require.Len(t, detail.Projects, 1)
	assert.Equal(t, project.ID().String(), detail.Projects[0].ID)
	assert.Equal(t, "core", detail.Projects[0].Key)
	assert.Equal(t, "Core", detail.Projects[0].Name)
}

func TestFinderGetByNameWithLatestVerdict(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)
	ctx := context.Background()

	project := insertProject(t, testDB, "judged", "Judged")
	insertGavelspaceWithProjects(t, testDB, "verdict-test", []projectmodel.Project{project})
	insertCaseFileWithPassingVerdict(t, testDB, project.ID(), "sha-pass")

	detail, err := finder.GetByName(ctx, "verdict-test")
	require.NoError(t, err)
	require.Len(t, detail.Projects, 1)
	assert.Equal(t, "pass", detail.Projects[0].LatestVerdict)
}

func TestFinderGetByNameNotFound(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)

	_, err := finder.GetByName(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gavelspace not found")
}

func TestFinderGetByNameReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)

	insertGavelspaceWithProjects(t, testDB, "ctx-get", nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := finder.GetByName(ctx, "ctx-get")
	assert.Error(t, err)
}

func TestFinderFindGavelspaceNameByProjectID(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)
	ctx := context.Background()

	project := insertProject(t, testDB, "linked", "Linked")
	insertGavelspaceWithProjects(t, testDB, "find-by-proj", []projectmodel.Project{project})

	name, found, err := finder.FindGavelspaceNameByProjectID(ctx, project.ID().String())
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "find-by-proj", name)
}

func TestFinderFindGavelspaceNameByProjectIDNotFound(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)

	name, found, err := finder.FindGavelspaceNameByProjectID(context.Background(), "00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, "", name)
}

func TestFinderFindGavelspaceNameByProjectIDReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := finder.FindGavelspaceNameByProjectID(ctx, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

func TestFinderGetByNameWithEmptyProjects(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)

	insertGavelspaceWithProjects(t, testDB, "no-projects", nil)

	detail, err := finder.GetByName(context.Background(), "no-projects")
	require.NoError(t, err)
	assert.Equal(t, "no-projects", detail.Name)
	assert.Empty(t, detail.Projects)
}

func TestFinderListReturnsErrorOnCorruptedCreatedAt(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)
	ctx := context.Background()

	insertGavelspaceWithProjects(t, testDB, "bad-ts", nil)

	_, err := testDB.ExecContext(ctx,
		"UPDATE gavelspaces SET created_at = 'not-a-date' WHERE name = ?", "bad-ts")
	require.NoError(t, err)

	_, _, err = finder.List(ctx, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created_at")
}

func TestFinderGetByNameReturnsErrorOnCorruptedCreatedAt(t *testing.T) {
	testDB := setupDB(t)
	finder := gspostgres.NewGavelspaceFinder(testDB)
	ctx := context.Background()

	insertGavelspaceWithProjects(t, testDB, "bad-ts-get", nil)

	_, err := testDB.ExecContext(ctx,
		"UPDATE gavelspaces SET created_at = 'not-a-date' WHERE name = ?", "bad-ts-get")
	require.NoError(t, err)

	_, err = finder.GetByName(ctx, "bad-ts-get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created_at")
}

func TestFinderListReturnsErrorOnDataQuerySchemaCorruption(t *testing.T) {
	testDB := setupDB(t)
	ctx := context.Background()

	insertGavelspaceWithProjects(t, testDB, "schema-break", nil)

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE gavelspaces RENAME COLUMN created_at TO created_at_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE gavelspaces RENAME COLUMN created_at_corrupted TO created_at")
	})

	finder := gspostgres.NewGavelspaceFinder(testDB)
	_, _, err = finder.List(ctx, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list gavelspaces")
}

func TestFinderGetByNameReturnsErrorOnProjectRefQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	ctx := context.Background()

	project := insertProject(t, testDB, "ref-break", "Ref Break")
	insertGavelspaceWithProjects(t, testDB, "proj-ref-break", []projectmodel.Project{project})

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE projects RENAME TO projects_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE projects_corrupted RENAME TO projects")
	})

	finder := gspostgres.NewGavelspaceFinder(testDB)
	_, err = finder.GetByName(ctx, "proj-ref-break")
	assert.Error(t, err)
}

func TestRepoSaveReturnsErrorOnForeignKeyViolation(t *testing.T) {
	testDB := setupDB(t)
	repo := gspostgres.NewRepository(testDB)
	ctx := context.Background()

	gavelspace, err := gsmodel.NewGavelspace("fk-fail")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, gavelspace))

	fakeID := projectmodel.NewProjectID(uuid.New())
	ref, err := gsmodel.NewProjectRef(fakeID, "//fake/...")
	require.NoError(t, err)
	require.NoError(t, gavelspace.AddProject(ref, time.Now().UTC()))

	err = repo.Save(ctx, gavelspace)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replace project refs")
}

func TestRepoFindByNameReturnsErrorOnProjectRefSchemaCorruption(t *testing.T) {
	testDB := setupDB(t)
	repo := gspostgres.NewRepository(testDB)
	ctx := context.Background()

	project := insertProject(t, testDB, "repo-break", "Repo Break")
	gavelspace := insertGavelspaceWithProjects(t, testDB, "repo-schema", []projectmodel.Project{project})

	_, err := testDB.ExecContext(ctx,
		"ALTER TABLE gavelspace_projects RENAME COLUMN target_pattern TO target_pattern_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE gavelspace_projects RENAME COLUMN target_pattern_corrupted TO target_pattern")
	})

	_, err = repo.FindByName(ctx, gavelspace.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load project refs")
}

func TestRepoFindByNameReturnsErrorOnCorruptedTargetPattern(t *testing.T) {
	testDB := setupDB(t)
	repo := gspostgres.NewRepository(testDB)
	ctx := context.Background()

	project := insertProject(t, testDB, "corrupt-ref", "Corrupt Ref")
	gavelspace := insertGavelspaceWithProjects(t, testDB, "corrupt-pattern", []projectmodel.Project{project})

	_, err := testDB.ExecContext(ctx,
		"UPDATE gavelspace_projects SET target_pattern = '' WHERE gavelspace_name = ?",
		gavelspace.ID().String())
	require.NoError(t, err)

	_, err = repo.FindByName(ctx, gavelspace.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project ref")
}
