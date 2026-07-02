package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	projectpostgres "github.com/usegavel/gavel/core/infrastructure/project/postgres"
)

func TestCoreProjectRepoSaveAndFindByID(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("backend", "Backend Service", "//backend/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)
	assert.Equal(t, project.ID(), found.ID())
	assert.Equal(t, "backend", found.Key())
	assert.Equal(t, "Backend Service", found.Name())
	assert.Equal(t, "//backend/...", found.TargetPattern())
	assert.Equal(t, "main", found.DefaultBranch())
}

func TestCoreProjectRepoFindByKey(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("frontend", "Frontend App", "//web/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByKey(ctx, "frontend")
	require.NoError(t, err)
	assert.Equal(t, project.ID(), found.ID())
}

func TestCoreProjectRepoFindByName(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("api", "API Gateway", "//api/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByName(ctx, "API Gateway")
	require.NoError(t, err)
	assert.Equal(t, project.ID(), found.ID())
}

func TestCoreProjectRepoFindByIDNotFound(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)

	_, err := repo.FindByID(context.Background(), mustGenerateProjectID(t))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCoreProjectRepoFindByKeyNotFound(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)

	_, err := repo.FindByKey(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCoreProjectRepoSaveWithLanguages(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("polyglot", "Polyglot", "//...")
	require.NoError(t, err)

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)
	golang, err := coverage.NewLanguage("go")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), []coverage.Language{java, golang},
		project.Gate(), nil, nil,
	)
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)
	require.Len(t, found.Languages(), 2)

	langNames := map[string]bool{}
	for _, l := range found.Languages() {
		langNames[l.String()] = true
	}
	assert.True(t, langNames["java"])
	assert.True(t, langNames["go"])
}

func TestCoreProjectRepoSaveWithQualityGate(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	zeroTolerance := qualitygate.NewZeroTolerance()
	cbs, err := qualitygate.NewCountBySeverity(0, 5, 10)
	require.NoError(t, err)
	minPercentage, err := qualitygate.NewMinPercentage(80.0)
	require.NoError(t, err)
	forbiddenList, err := qualitygate.NewForbiddenList([]string{"eval", "exec"})
	require.NoError(t, err)
	maxViolations, err := qualitygate.NewMaxViolations(3)
	require.NoError(t, err)
	mncc, err := qualitygate.NewMinNewCodeCoverage(70.0)
	require.NoError(t, err)

	ruleCodeQuality, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, zeroTolerance)
	require.NoError(t, err)
	ruleSAST, err := qualitygate.NewRule(evidence.SubtypeSAST, cbs)
	require.NoError(t, err)
	ruleCoverage, err := qualitygate.NewRule(evidence.SubtypeCoverage, minPercentage)
	require.NoError(t, err)
	ruleLicense, err := qualitygate.NewRule(evidence.SubtypeLicense, forbiddenList)
	require.NoError(t, err)
	ruleArchitecture, err := qualitygate.NewRule(evidence.SubtypeArchitecture, maxViolations)
	require.NoError(t, err)
	ruleNewCodeCoverage, err := qualitygate.NewRule(evidence.SubtypeNewCodeCoverage, mncc)
	require.NoError(t, err)

	qualityGate, err := qualitygate.NewGate([]qualitygate.Rule{ruleCodeQuality, ruleSAST, ruleCoverage, ruleLicense, ruleArchitecture, ruleNewCodeCoverage})
	require.NoError(t, err)

	project, err := projectmodel.NewProject("gated", "Gated Project", "//gated/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), nil, qualityGate, nil, nil,
	)
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	rules := found.Gate().Rules()
	require.Len(t, rules, 6)

	assert.Equal(t, "code_quality", rules[0].Subtype().String())
	_, isZT := rules[0].Strategy().(qualitygate.ZeroTolerance)
	assert.True(t, isZT)

	assert.Equal(t, "sast", rules[1].Subtype().String())
	loadedCBS, isCBS := rules[1].Strategy().(qualitygate.CountBySeverity)
	require.True(t, isCBS)
	assert.Equal(t, 0, loadedCBS.MaxError())
	assert.Equal(t, 5, loadedCBS.MaxWarning())
	assert.Equal(t, 10, loadedCBS.MaxNote())

	assert.Equal(t, "coverage", rules[2].Subtype().String())
	loadedMP, isMP := rules[2].Strategy().(qualitygate.MinPercentage)
	require.True(t, isMP)
	assert.InDelta(t, 80.0, loadedMP.Min(), 0.01)

	assert.Equal(t, "license", rules[3].Subtype().String())
	loadedFL, isFL := rules[3].Strategy().(qualitygate.ForbiddenList)
	require.True(t, isFL)
	assert.Equal(t, []string{"eval", "exec"}, loadedFL.Forbidden())

	assert.Equal(t, "architecture", rules[4].Subtype().String())
	loadedMV, isMV := rules[4].Strategy().(qualitygate.MaxViolations)
	require.True(t, isMV)
	assert.Equal(t, 3, loadedMV.Max())

	assert.Equal(t, "new_code_coverage", rules[5].Subtype().String())
	loadedMNCC, isMNCC := rules[5].Strategy().(qualitygate.MinNewCodeCoverage)
	require.True(t, isMNCC)
	assert.InDelta(t, 70.0, loadedMNCC.Min(), 0.01)
}

func TestCoreProjectRepoSaveOverwritesExisting(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("svc", "Service V1", "//svc/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	updated, err := projectmodel.ReconstituteProject(
		project.ID(), project.Key(), "Service V2", "//svc/v2/...",
		project.DefaultBranch(), nil, project.Gate(), nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, updated))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)
	assert.Equal(t, "Service V2", found.Name())
	assert.Equal(t, "//svc/v2/...", found.TargetPattern())
}

func TestCoreProjectRepoDuplicateKeyConflict(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	p1, err := projectmodel.NewProject("shared-key", "Project One", "//one/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, p1))

	p2, err := projectmodel.NewProject("shared-key", "Project Two", "//two/...")
	require.NoError(t, err)

	err = repo.Save(ctx, p2)
	assert.Error(t, err)
}

func TestCoreProjectRepoBaselineRoundTrip(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("bl-test", "Baseline Test", "//bl/...")
	require.NoError(t, err)

	cov := 85.5
	project.UpdateBaseline("main", []string{"fp-2", "fp-1"}, []string{"arch-b", "arch-a"}, &cov, nil)
	project.UpdateBaseline("develop", []string{"fp-3"}, nil, nil, nil)

	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	mainBL := found.Baseline("main")
	assert.Equal(t, []string{"fp-1", "fp-2"}, mainBL.Fingerprints())
	assert.Equal(t, []string{"arch-a", "arch-b"}, mainBL.ArchIDs())
	require.NotNil(t, mainBL.CoveragePercent())
	assert.InDelta(t, 85.5, *mainBL.CoveragePercent(), 0.01)
	assert.True(t, mainBL.HasPrevious())

	devBL := found.Baseline("develop")
	assert.Equal(t, []string{"fp-3"}, devBL.Fingerprints())
	assert.True(t, devBL.HasPrevious())

	featureBL := found.Baseline("feature")
	assert.True(t, featureBL.HasPrevious(), "unknown branch falls back to default-branch baseline")
	assert.Equal(t, []string{"fp-1", "fp-2"}, featureBL.Fingerprints())
}

func TestCoreProjectRepoBaselineUpdateReplacesExisting(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("bl-update", "BL Update", "//bl/...")
	require.NoError(t, err)

	cov1 := 70.0
	project.UpdateBaseline("main", []string{"fp-old"}, nil, &cov1, nil)
	require.NoError(t, repo.Save(ctx, project))

	loaded, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	cov2 := 90.0
	loaded.UpdateBaseline("main", []string{"fp-new-1", "fp-new-2"}, []string{"arch-1"}, &cov2, nil)
	require.NoError(t, repo.Save(ctx, loaded))

	reloaded, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	mainBL := reloaded.Baseline("main")
	assert.Equal(t, []string{"fp-new-1", "fp-new-2"}, mainBL.Fingerprints())
	assert.Equal(t, []string{"arch-1"}, mainBL.ArchIDs())
	require.NotNil(t, mainBL.CoveragePercent())
	assert.InDelta(t, 90.0, *mainBL.CoveragePercent(), 0.01)
}

func TestCoreProjectRepoSaveReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)

	project, err := projectmodel.NewProject("ctx-cancel", "Ctx Cancel", "//ctx/...")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.Save(ctx, project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin tx")
}

func TestCoreProjectRepoFindByIDReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("ctx-find", "Ctx Find", "//ctx/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.FindByID(cancelledCtx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query project")
}

func TestCoreProjectRepoFindByKeyReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("ctx-key", "Ctx Key", "//ctx/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.FindByKey(cancelledCtx, "ctx-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query project")
}

func TestCoreProjectRepoFindByNameReturnsErrorOnCancelledContext(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("ctx-name", "Ctx Name", "//ctx/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.FindByName(cancelledCtx, "Ctx Name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query project")
}

func TestCoreProjectRepoFindByNameNotFound(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)

	_, err := repo.FindByName(context.Background(), "nonexistent-name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCoreProjectRepoFindReturnsErrorOnCorruptedStrategyType(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	zt := qualitygate.NewZeroTolerance()
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, zt)
	require.NoError(t, err)
	qualityGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	project, err := projectmodel.NewProject("corrupt-strat", "Corrupt Strategy", "//corrupt/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), nil, qualityGate, nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"UPDATE project_quality_gate_rules SET strategy_type = 'garbage_type' WHERE project_id = ?",
		project.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load quality gate")
}

func TestCoreProjectRepoFindReturnsErrorOnCorruptedStrategyParams(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	strategies := []struct {
		name     string
		key      string
		subtype  evidence.Subtype
		strategy qualitygate.Strategy
	}{
		{"count_by_severity", "bad-cbs", evidence.SubtypeCodeQuality, mustCountBySeverity(t, 0, 5, 10)},
		{"min_percentage", "bad-mp", evidence.SubtypeCoverage, mustMinPercentage(t, 80.0)},
		{"forbidden_list", "bad-fl", evidence.SubtypeLicense, mustForbiddenList(t, []string{"eval"})},
		{"max_violations", "bad-mv", evidence.SubtypeArchitecture, mustMaxViolations(t, 3)},
		{"min_new_code_coverage", "bad-mncc", evidence.SubtypeNewCodeCoverage, mustMinNewCodeCoverage(t, 70.0)},
	}

	for _, tc := range strategies {
		t.Run(tc.name, func(t *testing.T) {
			key := tc.key
			rule, err := qualitygate.NewRule(tc.subtype, tc.strategy)
			require.NoError(t, err)
			qualityGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
			require.NoError(t, err)

			project, err := projectmodel.NewProject(key, "Bad Params", "//bad/...")
			require.NoError(t, err)
			project, err = projectmodel.ReconstituteProject(
				project.ID(), project.Key(), project.Name(), project.TargetPattern(),
				project.DefaultBranch(), nil, qualityGate, nil, nil,
			)
			require.NoError(t, err)
			require.NoError(t, repo.Save(ctx, project))

			_, err = testDB.ExecContext(ctx,
				"UPDATE project_quality_gate_rules SET strategy_params = '<<<invalid-json>>>' WHERE project_id = ?",
				project.ID().UUID())
			require.NoError(t, err)

			_, err = repo.FindByID(ctx, project.ID())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "load quality gate")
		})
	}
}

func TestCoreProjectRepoFindReturnsErrorOnCorruptedSubtype(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	zt := qualitygate.NewZeroTolerance()
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, zt)
	require.NoError(t, err)
	qualityGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	project, err := projectmodel.NewProject("bad-subtype", "Bad Subtype", "//bad/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), nil, qualityGate, nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"UPDATE project_quality_gate_rules SET subtype = 'invalid_subtype' WHERE project_id = ?",
		project.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load quality gate")
}

func TestCoreProjectRepoFindReturnsErrorOnCorruptedLanguage(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)

	project, err := projectmodel.NewProject("bad-lang", "Bad Lang", "//bad/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), []coverage.Language{java}, project.Gate(), nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"UPDATE project_languages SET language = '' WHERE project_id = ?",
		project.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load languages")
}

func TestCoreProjectRepoFindReturnsErrorOnCorruptedBaselineFingerprints(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("bad-bl-fp", "Bad BL FP", "//bad/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1"}, nil, nil, nil)
	require.NoError(t, repo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"UPDATE project_baselines SET fingerprints = '{\"bad\": true}' WHERE project_id = ?",
		project.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal fingerprints")
}

func TestCoreProjectRepoFindReturnsErrorOnCorruptedBaselineArchIDs(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("bad-bl-arch", "Bad BL Arch", "//bad/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1"}, []string{"arch-1"}, nil, nil)
	require.NoError(t, repo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"UPDATE project_baselines SET arch_ids = '{\"bad\": true}' WHERE project_id = ?",
		project.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal arch ids")
}

func TestCoreProjectRepoFindReturnsErrorOnCancelledContextForLanguages(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)

	project, err := projectmodel.NewProject("ctx-lang", "Ctx Lang", "//ctx/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), []coverage.Language{java}, project.Gate(), nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.FindByID(cancelledCtx, project.ID())
	assert.Error(t, err)
}

func TestCoreProjectRepoBaselineNilCoverageRoundTrip(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("nil-cov", "Nil Cov", "//nil/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1"}, []string{"arch-1"}, nil, nil)
	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	mainBL := found.Baseline("main")
	assert.Nil(t, mainBL.CoveragePercent())
	assert.Equal(t, []string{"fp-1"}, mainBL.Fingerprints())
	assert.Equal(t, []string{"arch-1"}, mainBL.ArchIDs())
}

func TestCoreProjectRepoBaselineFileCoverageRoundTrip(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("fc-test", "FC Test", "//fc/...")
	require.NoError(t, err)

	entry1, err := projectmodel.NewFileCoverageEntry("pkg/a.go", []int{1, 2, 3}, []int{4, 5})
	require.NoError(t, err)
	entry2, err := projectmodel.NewFileCoverageEntry("pkg/b.go", []int{10, 20}, nil)
	require.NoError(t, err)

	cov := 75.0
	project.UpdateBaseline("main", []string{"fp-1"}, nil, &cov, []projectmodel.FileCoverageEntry{entry1, entry2})
	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	mainBL := found.Baseline("main")
	require.NotNil(t, mainBL.CoveragePercent())
	assert.InDelta(t, 75.0, *mainBL.CoveragePercent(), 0.01)

	fileCoverage := mainBL.FileCoverage()
	require.Len(t, fileCoverage, 2)
	assert.Equal(t, "pkg/a.go", fileCoverage[0].FilePath())
	assert.Equal(t, []int{1, 2, 3}, fileCoverage[0].Covered())
	assert.Equal(t, []int{4, 5}, fileCoverage[0].Uncovered())
	assert.Equal(t, "pkg/b.go", fileCoverage[1].FilePath())
	assert.Equal(t, []int{10, 20}, fileCoverage[1].Covered())
	assert.Nil(t, fileCoverage[1].Uncovered())
}

func TestCoreProjectRepoBaselineNilFileCoverageRoundTrip(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("no-fc", "No FC", "//no/...")
	require.NoError(t, err)

	cov := 50.0
	project.UpdateBaseline("main", []string{"fp-1"}, nil, &cov, nil)
	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	mainBL := found.Baseline("main")
	assert.Nil(t, mainBL.FileCoverage())
	require.NotNil(t, mainBL.CoveragePercent())
	assert.InDelta(t, 50.0, *mainBL.CoveragePercent(), 0.01)
}

func TestCoreProjectRepoLanguageReplacement(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)
	golang, err := coverage.NewLanguage("go")
	require.NoError(t, err)

	project, err := projectmodel.NewProject("lang-replace", "Lang Replace", "//lang/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), []coverage.Language{java, golang},
		project.Gate(), nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)
	require.Len(t, found.Languages(), 2)

	python, err := coverage.NewLanguage("python")
	require.NoError(t, err)
	typescript, err := coverage.NewLanguage("typescript")
	require.NoError(t, err)
	rust, err := coverage.NewLanguage("rust")
	require.NoError(t, err)

	updated, err := projectmodel.ReconstituteProject(
		found.ID(), found.Key(), found.Name(), found.TargetPattern(),
		found.DefaultBranch(), []coverage.Language{python, typescript, rust},
		found.Gate(), nil, found.Baselines(),
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, updated))

	reloaded, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)
	require.Len(t, reloaded.Languages(), 3)

	langNames := map[string]bool{}
	for _, l := range reloaded.Languages() {
		langNames[l.String()] = true
	}
	assert.True(t, langNames["python"])
	assert.True(t, langNames["typescript"])
	assert.True(t, langNames["rust"])
	assert.False(t, langNames["java"], "old language should be removed")
	assert.False(t, langNames["go"], "old language should be removed")
}

func TestCoreProjectRepoLanguageClearedOnUpdate(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)

	project, err := projectmodel.NewProject("lang-clear", "Lang Clear", "//lang/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), []coverage.Language{java},
		project.Gate(), nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	cleared, err := projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), nil,
		project.Gate(), nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, cleared))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)
	assert.Empty(t, found.Languages())
}

func TestCoreProjectRepoQualityGateReplacement(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	zt := qualitygate.NewZeroTolerance()
	rule1, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, zt)
	require.NoError(t, err)
	gate1, err := qualitygate.NewGate([]qualitygate.Rule{rule1})
	require.NoError(t, err)

	project, err := projectmodel.NewProject("qg-replace", "QG Replace", "//qg/...")
	require.NoError(t, err)
	project, err = projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), nil, gate1, nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	cbs, err := qualitygate.NewCountBySeverity(1, 2, 3)
	require.NoError(t, err)
	mp, err := qualitygate.NewMinPercentage(75.0)
	require.NoError(t, err)
	rule2, err := qualitygate.NewRule(evidence.SubtypeSAST, cbs)
	require.NoError(t, err)
	rule3, err := qualitygate.NewRule(evidence.SubtypeCoverage, mp)
	require.NoError(t, err)
	gate2, err := qualitygate.NewGate([]qualitygate.Rule{rule2, rule3})
	require.NoError(t, err)

	updated, err := projectmodel.ReconstituteProject(
		project.ID(), project.Key(), project.Name(), project.TargetPattern(),
		project.DefaultBranch(), nil, gate2, nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, updated))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	rules := found.Gate().Rules()
	require.Len(t, rules, 2, "old rule should be replaced, not appended")

	assert.Equal(t, "sast", rules[0].Subtype().String())
	loadedCBS, isCBS := rules[0].Strategy().(qualitygate.CountBySeverity)
	require.True(t, isCBS)
	assert.Equal(t, 1, loadedCBS.MaxError())
	assert.Equal(t, 2, loadedCBS.MaxWarning())
	assert.Equal(t, 3, loadedCBS.MaxNote())

	assert.Equal(t, "coverage", rules[1].Subtype().String())
	loadedMP, isMP := rules[1].Strategy().(qualitygate.MinPercentage)
	require.True(t, isMP)
	assert.InDelta(t, 75.0, loadedMP.Min(), 0.01)
}

func TestCoreProjectRepoBaselineEmptyArchIDs(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("bl-no-arch", "BL No Arch", "//bl/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1", "fp-2"}, nil, nil, nil)
	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	mainBL := found.Baseline("main")
	assert.Equal(t, []string{"fp-1", "fp-2"}, mainBL.Fingerprints())
	assert.Nil(t, mainBL.ArchIDs())
	assert.Nil(t, mainBL.CoveragePercent())
	assert.True(t, mainBL.HasPrevious())
}

func TestCoreProjectRepoBaselineMultipleBranches(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("bl-multi", "BL Multi", "//bl/...")
	require.NoError(t, err)

	cov1 := 80.0
	cov2 := 95.0
	project.UpdateBaseline("main", []string{"fp-a", "fp-b"}, []string{"arch-x"}, &cov1, nil)
	project.UpdateBaseline("develop", []string{"fp-c"}, []string{"arch-y", "arch-z"}, &cov2, nil)
	project.UpdateBaseline("release", []string{"fp-d"}, nil, nil, nil)

	require.NoError(t, repo.Save(ctx, project))

	found, err := repo.FindByID(ctx, project.ID())
	require.NoError(t, err)

	mainBL := found.Baseline("main")
	assert.Equal(t, []string{"fp-a", "fp-b"}, mainBL.Fingerprints())
	assert.Equal(t, []string{"arch-x"}, mainBL.ArchIDs())
	require.NotNil(t, mainBL.CoveragePercent())
	assert.InDelta(t, 80.0, *mainBL.CoveragePercent(), 0.01)

	devBL := found.Baseline("develop")
	assert.Equal(t, []string{"fp-c"}, devBL.Fingerprints())
	assert.Equal(t, []string{"arch-y", "arch-z"}, devBL.ArchIDs())
	require.NotNil(t, devBL.CoveragePercent())
	assert.InDelta(t, 95.0, *devBL.CoveragePercent(), 0.01)

	relBL := found.Baseline("release")
	assert.Equal(t, []string{"fp-d"}, relBL.Fingerprints())
	assert.Nil(t, relBL.ArchIDs())
	assert.Nil(t, relBL.CoveragePercent())
}

func mustCountBySeverity(t *testing.T, maxErr, maxWarn, maxNote int) qualitygate.CountBySeverity {
	t.Helper()
	s, err := qualitygate.NewCountBySeverity(maxErr, maxWarn, maxNote)
	require.NoError(t, err)
	return s
}

func mustMinPercentage(t *testing.T, min float64) qualitygate.MinPercentage {
	t.Helper()
	s, err := qualitygate.NewMinPercentage(min)
	require.NoError(t, err)
	return s
}

func mustForbiddenList(t *testing.T, forbidden []string) qualitygate.ForbiddenList {
	t.Helper()
	s, err := qualitygate.NewForbiddenList(forbidden)
	require.NoError(t, err)
	return s
}

func mustMaxViolations(t *testing.T, max int) qualitygate.MaxViolations {
	t.Helper()
	s, err := qualitygate.NewMaxViolations(max)
	require.NoError(t, err)
	return s
}

func mustMinNewCodeCoverage(t *testing.T, min float64) qualitygate.MinNewCodeCoverage {
	t.Helper()
	s, err := qualitygate.NewMinNewCodeCoverage(min)
	require.NoError(t, err)
	return s
}

func TestCoreProjectRepoSaveReturnsErrorOnLanguageInsertFailure(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	lang, err := coverage.NewLanguage("go")
	require.NoError(t, err)
	project, err := projectmodel.NewProject("lang-fail", "Lang Fail", "//lang/...")
	require.NoError(t, err)
	now := time.Now().UTC()
	project.UpdateLanguages([]coverage.Language{lang}, now)

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE project_languages ADD COLUMN extra TEXT NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE project_languages DROP COLUMN extra")
	})

	err = repo.Save(ctx, project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replace languages")
}

func TestCoreProjectRepoSaveReturnsErrorOnQualityGateInsertFailure(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("qg-fail", "QG Fail", "//qg/...")
	require.NoError(t, err)
	zt := qualitygate.NewZeroTolerance()
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, zt)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)
	project.UpdateQualityGate(gate, time.Now().UTC())

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE project_quality_gate_rules ADD COLUMN extra TEXT NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE project_quality_gate_rules DROP COLUMN extra")
	})

	err = repo.Save(ctx, project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replace quality gate rules")
}

func TestCoreProjectRepoSaveReturnsErrorOnBaselineInsertFailure(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("bl-fail", "BL Fail", "//bl/...")
	require.NoError(t, err)
	project.UpdateBaseline("main", []string{"fp-1"}, nil, nil, nil)

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE project_baselines ADD COLUMN extra TEXT NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE project_baselines DROP COLUMN extra")
	})

	err = repo.Save(ctx, project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replace baselines")
}

func TestCoreProjectRepoFindReturnsErrorOnLanguagesQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("lang-query", "Lang Query", "//lang/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE project_languages RENAME TO project_languages_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE project_languages_corrupted RENAME TO project_languages")
	})

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load languages")
}

func TestCoreProjectRepoFindReturnsErrorOnQualityGateQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("qg-query", "QG Query", "//qg/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE project_quality_gate_rules RENAME TO project_quality_gate_rules_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE project_quality_gate_rules_corrupted RENAME TO project_quality_gate_rules")
	})

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load quality gate")
}

func TestCoreProjectRepoFindReturnsErrorOnBaselinesQueryFailure(t *testing.T) {
	testDB := setupDB(t)
	repo := projectpostgres.NewRepository(testDB)
	ctx := context.Background()

	project, err := projectmodel.NewProject("bl-query", "BL Query", "//bl/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, project))

	_, err = testDB.ExecContext(ctx,
		"ALTER TABLE project_baselines RENAME TO project_baselines_corrupted")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"ALTER TABLE project_baselines_corrupted RENAME TO project_baselines")
	})

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load baselines")
}

func TestCoreProjectRepoFindReturnsErrorOnCoverageByFile(t *testing.T) {
	database := setupDB(t)
	repo := projectpostgres.NewRepository(database)
	ctx := context.Background()

	project, err := projectmodel.NewProject("bad-fc", "Bad FC", "//bad/...")
	require.NoError(t, err)
	entry, err := projectmodel.NewFileCoverageEntry("pkg/a.go", []int{1}, nil)
	require.NoError(t, err)
	cov := 50.0
	project.UpdateBaseline("main", []string{"fp-1"}, nil, &cov, []projectmodel.FileCoverageEntry{entry})
	require.NoError(t, repo.Save(ctx, project))

	_, err = database.ExecContext(ctx,
		"UPDATE project_baselines SET coverage_by_file = '42' WHERE project_id = ?",
		project.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file coverage")
}

func TestCoreProjectRepoFindReturnsErrorOnInvalidFileCoverageEntry(t *testing.T) {
	database := setupDB(t)
	repo := projectpostgres.NewRepository(database)
	ctx := context.Background()

	project, err := projectmodel.NewProject("empty-fp", "Empty FP", "//empty/...")
	require.NoError(t, err)
	entry, err := projectmodel.NewFileCoverageEntry("pkg/a.go", []int{1}, nil)
	require.NoError(t, err)
	cov := 50.0
	project.UpdateBaseline("main", []string{"fp-1"}, nil, &cov, []projectmodel.FileCoverageEntry{entry})
	require.NoError(t, repo.Save(ctx, project))

	_, err = database.ExecContext(ctx,
		`UPDATE project_baselines SET coverage_by_file = '[{"file_path":"","covered":[],"uncovered":[]}]' WHERE project_id = ?`,
		project.ID().UUID())
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, project.ID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file coverage")
}
