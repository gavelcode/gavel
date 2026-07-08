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
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	casefilepostgres "github.com/usegavel/gavel/core/infrastructure/casefile/postgres"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"
)

func TestProjectQueryList(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	projectAlpha, err := projectmodel.NewProject(testTenantID, "alpha", "Alpha", "//alpha/...")
	require.NoError(t, err)
	p2, err := projectmodel.NewProject(testTenantID, "beta", "Beta", "//beta/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, projectAlpha))
	require.NoError(t, projRepo.Save(ctx, p2))

	caseFile := newTestCaseFile(t, projectAlpha.ID(), "abc", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	verdict := newTestVerdict(t)
	caseFile, err = casefile.ReconstituteCaseFile(caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(), caseFile.StartedAt(), caseFile.Evidences(), &verdict, false)
	require.NoError(t, err)
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := query.List(ctx, testTenantID, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 2)
	assert.GreaterOrEqual(t, len(items), 2)

	found := false
	for _, item := range items {
		if item.Key == "alpha" {
			assert.Equal(t, "pass", item.LatestVerdict)
			assert.Equal(t, 2, item.TotalFindings)
			found = true
		}
	}
	assert.True(t, found)
}

func TestProjectQueryGetByID(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)
	zt := qualitygate.NewZeroTolerance()
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, zt)
	require.NoError(t, err)
	qualityGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	project, err := projectmodel.NewProject(testTenantID, "detailed", "Detailed Project", "//detailed/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), testTenantID, project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), []coverage.Language{java}, qualityGate, nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	caseFile := newTestCaseFile(t, project.ID(), "abc123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	verdict := newTestVerdict(t)
	caseFile, err = casefile.ReconstituteCaseFile(caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(), caseFile.StartedAt(), caseFile.Evidences(), &verdict, false)
	require.NoError(t, err)
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	detail, err := query.GetByID(ctx, testTenantID, project.ID().String())
	require.NoError(t, err)
	assert.Equal(t, "detailed", detail.Key)
	assert.Equal(t, "Detailed Project", detail.Name)
	assert.Equal(t, "//detailed/...", detail.TargetPattern)
	require.Len(t, detail.Languages, 1)
	assert.Equal(t, "java", detail.Languages[0])
	require.Len(t, detail.QualityGateRules, 1)
	assert.Equal(t, "code_quality", detail.QualityGateRules[0].Subtype)
	assert.Equal(t, "zero_tolerance", detail.QualityGateRules[0].StrategyType)

	assert.Equal(t, 1, detail.SeverityCounts["error"])
	assert.Equal(t, 1, detail.SeverityCounts["warning"])
	assert.Equal(t, "pass", detail.LatestVerdict)
	assert.Equal(t, 2, detail.TotalFindings)
}

func TestProjectQueryGetByKey(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "by-key", "By Key", "//key/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	detail, err := query.GetByKey(ctx, testTenantID, "by-key")
	require.NoError(t, err)
	assert.Equal(t, project.ID().String(), detail.ID)
	assert.Equal(t, "By Key", detail.Name)
}

func TestProjectQueryReturnsEmptySlicesNotNil(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "no-extras", "No Extras", "//bare/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	t.Run("GetByKey", func(t *testing.T) {
		detail, err := query.GetByKey(ctx, testTenantID, "no-extras")
		require.NoError(t, err)

		assert.NotNil(t, detail.Languages, "Languages must be empty slice, not nil")
		assert.Empty(t, detail.Languages)
		assert.NotNil(t, detail.QualityGateRules, "QualityGateRules must be empty slice, not nil")
		assert.Empty(t, detail.QualityGateRules)
	})

	t.Run("GetByID", func(t *testing.T) {
		detail, err := query.GetByID(ctx, testTenantID, project.ID().String())
		require.NoError(t, err)

		assert.NotNil(t, detail.Languages, "Languages must be empty slice, not nil")
		assert.Empty(t, detail.Languages)
		assert.NotNil(t, detail.QualityGateRules, "QualityGateRules must be empty slice, not nil")
		assert.Empty(t, detail.QualityGateRules)
	})
}

func TestProjectQueryGetByIDNotFound(t *testing.T) {
	db := setupDB(t)
	query := projectpostgres.NewProjectFinder(db)

	_, err := query.GetByID(context.Background(), testTenantID, "nonexistent-id")
	assert.Error(t, err)
}

func TestProjectQueryGetByKeyNotFound(t *testing.T) {
	db := setupDB(t)
	query := projectpostgres.NewProjectFinder(db)

	_, err := query.GetByKey(context.Background(), testTenantID, "nonexistent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProjectQueryListReturnsErrorOnCancelledContext(t *testing.T) {
	db := setupDB(t)
	query := projectpostgres.NewProjectFinder(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := query.List(ctx, testTenantID, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count projects")
}

func TestProjectQueryGetByIDReturnsErrorOnCancelledContext(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "ctx-get-id", "Ctx Get ID", "//ctx/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = query.GetByID(cancelledCtx, testTenantID, project.ID().String())
	assert.Error(t, err)
}

func TestProjectQueryGetByKeyReturnsErrorOnCancelledContext(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "ctx-get-key", "Ctx Get Key", "//ctx/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = query.GetByKey(cancelledCtx, testTenantID, "ctx-get-key")
	assert.Error(t, err)
}

func TestProjectQueryListPagination(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	for i := range 3 {
		key := fmt.Sprintf("page-%d", i)
		p, err := projectmodel.NewProject(testTenantID, key, key, "//page/...")
		require.NoError(t, err)
		require.NoError(t, projRepo.Save(ctx, p))
	}

	items, total, err := query.List(ctx, testTenantID, 2, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 3)
	assert.Len(t, items, 2)

	items2, total2, err := query.List(ctx, testTenantID, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, total, total2)
	assert.GreaterOrEqual(t, len(items2), 1)
}

func TestProjectQueryGetByKeyWithQualityGateRules(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	zt := qualitygate.NewZeroTolerance()
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, zt)
	require.NoError(t, err)
	qualityGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)

	project, err := projectmodel.NewProject(testTenantID, "qg-key", "QG Key Project", "//qg/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), testTenantID, project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), []coverage.Language{java}, qualityGate, nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	caseFile := newTestCaseFile(t, project.ID(), "qgkey123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	verdict := newTestVerdict(t)
	caseFile, err = casefile.ReconstituteCaseFile(caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(), caseFile.StartedAt(), caseFile.Evidences(), &verdict, false)
	require.NoError(t, err)
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	detail, err := query.GetByKey(ctx, testTenantID, "qg-key")
	require.NoError(t, err)
	assert.Equal(t, "QG Key Project", detail.Name)
	require.Len(t, detail.QualityGateRules, 1)
	assert.Equal(t, "code_quality", detail.QualityGateRules[0].Subtype)
	assert.Equal(t, "zero_tolerance", detail.QualityGateRules[0].StrategyType)
	assert.Equal(t, 1, detail.SeverityCounts["error"])
	assert.Equal(t, "pass", detail.LatestVerdict)
}

func TestProjectQueryGetByIDWithoutCasefiles(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)
	zt := qualitygate.NewZeroTolerance()
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, zt)
	require.NoError(t, err)
	qualityGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	project, err := projectmodel.NewProject(testTenantID, "no-cases", "No CaseFiles", "//nocase/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), testTenantID, project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), []coverage.Language{java}, qualityGate, nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	detail, err := query.GetByID(ctx, testTenantID, project.ID().String())
	require.NoError(t, err)
	assert.Equal(t, "no-cases", detail.Key)
	assert.Equal(t, "No CaseFiles", detail.Name)
	assert.Equal(t, "", detail.LatestVerdict)
	assert.Equal(t, 0, detail.TotalFindings)
	assert.Empty(t, detail.SeverityCounts)
	require.Len(t, detail.Languages, 1)
	assert.Equal(t, "java", detail.Languages[0])
	require.Len(t, detail.QualityGateRules, 1)
	assert.Equal(t, "code_quality", detail.QualityGateRules[0].Subtype)
}

func TestProjectQueryGetByKeyWithoutCasefiles(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "no-cases-key", "No CaseFiles Key", "//nocase/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	detail, err := query.GetByKey(ctx, testTenantID, "no-cases-key")
	require.NoError(t, err)
	assert.Equal(t, "", detail.LatestVerdict)
	assert.Equal(t, 0, detail.TotalFindings)
	assert.Empty(t, detail.SeverityCounts)
}

func TestProjectQueryGetByIDSeverityCountsAllTypes(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "sev-counts", "Severity Counts", "//sev/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	caseFile := newTestCaseFile(t, project.ID(), "sev123", "main")
	ev := newMultiSeverityEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))

	verdict := newTestVerdict(t)
	caseFile, err = casefile.ReconstituteCaseFile(caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(), caseFile.StartedAt(), caseFile.Evidences(), &verdict, false)
	require.NoError(t, err)
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	detail, err := query.GetByID(ctx, testTenantID, project.ID().String())
	require.NoError(t, err)
	assert.Equal(t, 1, detail.SeverityCounts["error"])
	assert.Equal(t, 2, detail.SeverityCounts["warning"])
	assert.Equal(t, 1, detail.SeverityCounts["note"])
	assert.Equal(t, "pass", detail.LatestVerdict)
	assert.Equal(t, 4, detail.TotalFindings)
}

func TestProjectQueryListWithCasefilesAndWithout(t *testing.T) {
	db := setupDB(t)
	projRepo := projectpostgres.NewRepository(db)
	cfRepo := casefilepostgres.NewRepository(db)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	withCF, err := projectmodel.NewProject(testTenantID, "with-cf", "With CaseFile", "//with/...")
	require.NoError(t, err)
	withoutCF, err := projectmodel.NewProject(testTenantID, "without-cf", "Without CaseFile", "//without/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, withCF))
	require.NoError(t, projRepo.Save(ctx, withoutCF))

	caseFile := newTestCaseFile(t, withCF.ID(), "listcf123", "main")
	ev := newFindingsEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	verdict := newTestVerdict(t)
	caseFile, err = casefile.ReconstituteCaseFile(caseFile.ID(), testTenantID, caseFile.ProjectID(), caseFile.CommitSHA(), caseFile.Branch(), caseFile.StartedAt(), caseFile.Evidences(), &verdict, false)
	require.NoError(t, err)
	require.NoError(t, cfRepo.Save(ctx, caseFile))

	items, total, err := query.List(ctx, testTenantID, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 2)

	var foundWith, foundWithout bool
	for _, item := range items {
		if item.Key == "with-cf" {
			assert.Equal(t, "pass", item.LatestVerdict)
			assert.Equal(t, 2, item.TotalFindings)
			foundWith = true
		}
		if item.Key == "without-cf" {
			assert.Equal(t, "", item.LatestVerdict)
			assert.Equal(t, 0, item.TotalFindings)
			foundWithout = true
		}
	}
	assert.True(t, foundWith)
	assert.True(t, foundWithout)
}

func TestProjectQueryGetByIDReturnsErrorOnCorruptedCreatedAt(t *testing.T) {
	testDB := setupDB(t)
	projRepo := projectpostgres.NewRepository(testDB)
	query := projectpostgres.NewProjectFinder(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "bad-date-id", "Bad Date", "//bad/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"UPDATE projects SET created_at = 'not-a-date' WHERE id = ?",
		project.ID().UUID())
	require.NoError(t, err)

	_, err = query.GetByID(ctx, testTenantID, project.ID().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created_at")
}

func TestProjectQueryGetByKeyReturnsErrorOnCorruptedCreatedAt(t *testing.T) {
	testDB := setupDB(t)
	projRepo := projectpostgres.NewRepository(testDB)
	query := projectpostgres.NewProjectFinder(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "bad-date-key", "Bad Date Key", "//bad/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"UPDATE projects SET created_at = 'not-a-date' WHERE id = ?",
		project.ID().UUID())
	require.NoError(t, err)

	_, err = query.GetByKey(ctx, testTenantID, "bad-date-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created_at")
}

func TestProjectQueryListReturnsErrorOnCorruptedCreatedAt(t *testing.T) {
	testDB := setupDB(t)
	projRepo := projectpostgres.NewRepository(testDB)
	query := projectpostgres.NewProjectFinder(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject(testTenantID, "bad-date-list", "Bad Date List", "//bad/...")
	require.NoError(t, err)
	require.NoError(t, projRepo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"UPDATE projects SET created_at = 'not-a-date' WHERE id = ?",
		project.ID().UUID())
	require.NoError(t, err)

	_, _, err = query.List(ctx, testTenantID, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created_at")
}

func TestProjectQueryListWithNoProjects(t *testing.T) {
	db := setupDB(t)
	query := projectpostgres.NewProjectFinder(db)
	ctx := context.Background()

	items, total, err := query.List(ctx, testTenantID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}
