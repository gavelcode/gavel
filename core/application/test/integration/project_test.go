package appintegration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	projectcreate "github.com/usegavel/gavel/core/application/project/create"
	"github.com/usegavel/gavel/core/application/project/updatelanguages"
	"github.com/usegavel/gavel/core/application/project/updatequalitygate"
	"github.com/usegavel/gavel/core/application/project/updatetargetpattern"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
)

func mustCreateAndFindProject(t *testing.T, repo *memproject.ProjectRepository, key string) projectmodel.Project {
	t.Helper()
	ctx := context.Background()

	handler := projectcreate.NewHandler(repo)
	cmd, err := projectcreate.NewCommand(testTenant, key, key+"-name", "//"+key+"/...")
	require.NoError(t, err)

	result, err := handler.Execute(ctx, cmd)
	require.NoError(t, err)
	require.NotEmpty(t, result.ProjectID)

	projectID, err := projectmodel.ParseProjectID(result.ProjectID)
	require.NoError(t, err)

	project, err := repo.FindByID(ctx, testTenantID, projectID)
	require.NoError(t, err)

	return project
}

func TestProjectLifecycle_CreateAndRetrieve(t *testing.T) {
	ctx := context.Background()
	repo := memproject.NewProjectRepository()

	handler := projectcreate.NewHandler(repo)
	cmd, err := projectcreate.NewCommand(testTenant, "mylib", "My Library", "//mylib/...")
	require.NoError(t, err)

	result, err := handler.Execute(ctx, cmd)
	require.NoError(t, err)
	require.NotEmpty(t, result.ProjectID)

	found, err := repo.FindByKey(ctx, testTenantID, "mylib")
	require.NoError(t, err)

	assert.Equal(t, result.ProjectID, found.ID().String())
	assert.Equal(t, "mylib", found.Key())
	assert.Equal(t, "My Library", found.Name())
	assert.Equal(t, "//mylib/...", found.TargetPattern())
	assert.Equal(t, "main", found.DefaultBranch())
}

func TestProjectLifecycle_UpdateQualityGate(t *testing.T) {
	ctx := context.Background()
	repo := memproject.NewProjectRepository()

	project := mustCreateAndFindProject(t, repo, "gate-proj")

	handler := updatequalitygate.NewHandler(repo)
	input := updatequalitygate.Input{
		Rules: []updatequalitygate.RuleInput{
			{
				Subtype: "code_quality",
				Strategy: updatequalitygate.StrategyInput{
					Type: updatequalitygate.StrategyTypeZeroTolerance,
				},
			},
		},
	}
	cmd, err := updatequalitygate.NewCommand(testTenant, project.ID().String(), input)
	require.NoError(t, err)

	result, err := handler.Execute(ctx, cmd)
	require.NoError(t, err)
	assert.True(t, result.Changed)

	updated, err := repo.FindByID(ctx, testTenantID, project.ID())
	require.NoError(t, err)

	rules := updated.Gate().Rules()
	require.Len(t, rules, 1)
	assert.Equal(t, "code_quality", rules[0].Subtype().String())
}

func TestProjectLifecycle_UpdateLanguages(t *testing.T) {
	ctx := context.Background()
	repo := memproject.NewProjectRepository()

	project := mustCreateAndFindProject(t, repo, "lang-proj")

	handler := updatelanguages.NewHandler(repo)
	cmd, err := updatelanguages.NewCommand(testTenant, project.ID().String(), []string{"go", "java"})
	require.NoError(t, err)

	_, err = handler.Execute(ctx, cmd)
	require.NoError(t, err)

	updated, err := repo.FindByID(ctx, testTenantID, project.ID())
	require.NoError(t, err)

	languages := updated.Languages()
	require.Len(t, languages, 2)
	assert.Equal(t, "go", languages[0].String())
	assert.Equal(t, "java", languages[1].String())
}

func TestProjectLifecycle_UpdateTargetPattern(t *testing.T) {
	ctx := context.Background()
	repo := memproject.NewProjectRepository()

	project := mustCreateAndFindProject(t, repo, "target-proj")

	handler := updatetargetpattern.NewHandler(repo)
	cmd, err := updatetargetpattern.NewCommand(testTenant, project.ID().String(), "//newpkg/...")
	require.NoError(t, err)

	_, err = handler.Execute(ctx, cmd)
	require.NoError(t, err)

	updated, err := repo.FindByID(ctx, testTenantID, project.ID())
	require.NoError(t, err)

	assert.Equal(t, "//newpkg/...", updated.TargetPattern())
}

func TestProjectLifecycle_FullConfiguration(t *testing.T) {
	ctx := context.Background()
	repo := memproject.NewProjectRepository()

	project := mustCreateAndFindProject(t, repo, "full-proj")
	projectID := project.ID().String()

	gateHandler := updatequalitygate.NewHandler(repo)
	gateInput := updatequalitygate.Input{
		Rules: []updatequalitygate.RuleInput{
			{
				Subtype: "code_quality",
				Strategy: updatequalitygate.StrategyInput{
					Type:            updatequalitygate.StrategyTypeCountBySeverity,
					CountBySeverity: &updatequalitygate.CountBySeverity{MaxError: 0, MaxWarning: 5, MaxNote: 10},
				},
			},
			{
				Subtype: "coverage",
				Strategy: updatequalitygate.StrategyInput{
					Type:          updatequalitygate.StrategyTypeMinPercentage,
					MinPercentage: &updatequalitygate.MinPercentage{Min: 80},
				},
			},
		},
	}
	gateCmd, err := updatequalitygate.NewCommand(testTenant, projectID, gateInput)
	require.NoError(t, err)
	_, err = gateHandler.Execute(ctx, gateCmd)
	require.NoError(t, err)

	langHandler := updatelanguages.NewHandler(repo)
	langCmd, err := updatelanguages.NewCommand(testTenant, projectID, []string{"go", "typescript"})
	require.NoError(t, err)
	_, err = langHandler.Execute(ctx, langCmd)
	require.NoError(t, err)

	targetHandler := updatetargetpattern.NewHandler(repo)
	targetCmd, err := updatetargetpattern.NewCommand(testTenant, projectID, "//services/...")
	require.NoError(t, err)
	_, err = targetHandler.Execute(ctx, targetCmd)
	require.NoError(t, err)

	final, err := repo.FindByID(ctx, testTenantID, project.ID())
	require.NoError(t, err)

	assert.Equal(t, "full-proj", final.Key())
	assert.Equal(t, "//services/...", final.TargetPattern())

	languages := final.Languages()
	require.Len(t, languages, 2)
	assert.Equal(t, "go", languages[0].String())
	assert.Equal(t, "typescript", languages[1].String())

	rules := final.Gate().Rules()
	require.Len(t, rules, 2)
}
