package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	casefilepostgres "github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
)

func TestCaseFileQueryListByProject(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := casefilepostgres.NewCaseFileFinder(db)
	ctx := context.Background()

	for i := range 3 {
		caseFile := newTestCaseFile(t, project.ID(), "sha"+string(rune('a'+i)), "main")
		require.NoError(t, cfRepo.Save(ctx, caseFile))
	}

	items, total, err := query.ListByProject(ctx, project.ID().String(), "", 2, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, items, 2)

	items2, total2, err := query.ListByProject(ctx, project.ID().String(), "", 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, total2)
	assert.Len(t, items2, 1)
}

func TestCaseFileQueryGetByID(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := casefilepostgres.NewCaseFileFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	var err error
	verdict := newTestVerdict(t)
	caseFile, err = casefile.ReconstituteCaseFile(caseFile.ID(), caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(), caseFile.StartedAt(), caseFile.Evidences(), &verdict, false)
	require.NoError(t, err)
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	detail, err := query.GetByID(ctx, caseFile.ID().String())
	require.NoError(t, err)
	assert.Equal(t, caseFile.ID().String(), detail.ID)
	assert.Equal(t, "abc123", detail.CommitSHA)
	assert.Equal(t, "fail", detail.VerdictOutcome)

	require.Len(t, detail.Evidences, 1)
	assert.Equal(t, "code_quality", detail.Evidences[0].Subtype)
	assert.Equal(t, "pmd", detail.Evidences[0].Source)

	require.Len(t, detail.Rulings, 2)
	assert.True(t, detail.Rulings[0].Passed)
	assert.False(t, detail.Rulings[1].Passed)
}

func TestCaseFileQueryGetByIDNotFound(t *testing.T) {
	db := setupDB(t)
	query := casefilepostgres.NewCaseFileFinder(db)

	_, err := query.GetByID(context.Background(), "nonexistent-id")
	assert.Error(t, err)
}

func TestFindingQueryList(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := casefilepostgres.NewFindingFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := query.List(ctx, findinglist.Filters{CaseFileID: caseFile.ID().String()}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
}

func TestFindingQueryList_IncludesCommitAndProjectKey(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := casefilepostgres.NewFindingFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123commit", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, _, err := query.List(ctx, findinglist.Filters{CaseFileID: caseFile.ID().String()}, 100, 0)
	require.NoError(t, err)
	require.NotEmpty(t, items)

	for _, item := range items {
		assert.Equal(t, "abc123commit", item.CommitSHA, "every finding row must carry its case file commit sha")
		assert.Equal(t, project.Key(), item.ProjectKey, "every finding row must carry its project key")
	}
}

func TestFindingQueryListWithFilters(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := casefilepostgres.NewFindingFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := query.List(ctx, findinglist.Filters{
		CaseFileID: caseFile.ID().String(),
		Severity:   "error",
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "spotbugs", items[0].Tool)
	assert.Equal(t, "NP_NULL_DEREF", items[0].RuleID)
}

func TestFindingQueryListFilterByTool(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	finder := casefilepostgres.NewFindingFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := finder.List(ctx, findinglist.Filters{
		CaseFileID: caseFile.ID().String(),
		Tool:       "pmd",
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "pmd", items[0].Tool)
	assert.Equal(t, "UnusedVariable", items[0].RuleID)
}

func TestFindingQueryListFilterByStatus(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	finder := casefilepostgres.NewFindingFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := finder.List(ctx, findinglist.Filters{
		CaseFileID: caseFile.ID().String(),
		Status:     "new",
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)

	items, total, err = finder.List(ctx, findinglist.Filters{
		CaseFileID: caseFile.ID().String(),
		Status:     "resolved",
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

func TestFindingQueryListFilterByFilePath(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	finder := casefilepostgres.NewFindingFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := finder.List(ctx, findinglist.Filters{
		CaseFileID: caseFile.ID().String(),
		FilePath:   "src/Foo",
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "src/Foo.java", items[0].FilePath)

	items, total, err = finder.List(ctx, findinglist.Filters{
		CaseFileID: caseFile.ID().String(),
		FilePath:   "src/",
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
}

func TestFindingQueryListFilterByProjectID(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	finder := casefilepostgres.NewFindingFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := finder.List(ctx, findinglist.Filters{
		ProjectID: project.ID().String(),
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)

	items, total, err = finder.List(ctx, findinglist.Filters{
		ProjectID: mustGenerateProjectID(t).String(),
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

func TestFindingQueryListFilterByGavelspace(t *testing.T) {
	database := setupDB(t)
	project := insertTestProject(t, database)
	cfRepo := casefilepostgres.NewRepository(database)
	finder := casefilepostgres.NewFindingFinder(database)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	insertGavelspaceProject(t, database, "my-gavelspace", project.ID().String())

	items, total, err := finder.List(ctx, findinglist.Filters{
		Gavelspace: "my-gavelspace",
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)

	items, total, err = finder.List(ctx, findinglist.Filters{
		Gavelspace: "other-gavelspace",
	}, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

func TestCaseFileQueryListByProjectWithGavelspaceFilter(t *testing.T) {
	database := setupDB(t)
	project := insertTestProject(t, database)
	cfRepo := casefilepostgres.NewRepository(database)
	cfFinder := casefilepostgres.NewCaseFileFinder(database)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	insertGavelspaceProject(t, database, "my-gavelspace", project.ID().String())

	items, total, err := cfFinder.ListByProject(ctx, "", "my-gavelspace", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, caseFile.ID().String(), items[0].ID)

	items, total, err = cfFinder.ListByProject(ctx, "", "other-gavelspace", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

func TestCaseFileQueryGetByIDWithCoverage(t *testing.T) {
	db := setupDB(t)
	project := insertTestProject(t, db)
	cfRepo := casefilepostgres.NewRepository(db)
	cfFinder := casefilepostgres.NewCaseFileFinder(db)
	ctx := context.Background()

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newCoverageEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	detail, err := cfFinder.GetByID(ctx, caseFile.ID().String())
	require.NoError(t, err)
	require.NotNil(t, detail.CoveragePercent)
	assert.InDelta(t, 80.0, *detail.CoveragePercent, 0.01)
}
