package updatequalitygate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/project/updatequalitygate"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func TestHandlerExecuteSuccessful(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatequalitygate.NewHandler(projects)
	cmd := mustCommand(t, project.ID().String(), strictGateInput())

	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	persisted, err := projects.FindByID(context.Background(), project.ID())
	require.NoError(t, err)
	assert.Equal(t, 1, len(persisted.Gate().Rules()), "quality gate updated and saved")
}

func TestHandlerExecuteInvalidProjectID(t *testing.T) {
	projects := newFakeProjectRepo()

	handler := updatequalitygate.NewHandler(projects)
	cmd := mustCommand(t, "missing", updatequalitygate.Input{})

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteProjectNotFound(t *testing.T) {
	projects := newFakeProjectRepo()

	handler := updatequalitygate.NewHandler(projects)
	cmd, err := updatequalitygate.NewCommand("11111111-1111-1111-1111-111111111111", updatequalitygate.Input{})
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteSaveErrorPropagated(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)
	projects.saveErr = errors.New("disk full")

	handler := updatequalitygate.NewHandler(projects)

	_, err := handler.Execute(context.Background(), mustCommand(t, project.ID().String(), strictGateInput()))
	require.Error(t, err)
}

func TestHandlerExecuteEmptyQualityGateAccepted(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatequalitygate.NewHandler(projects)
	cmd := mustCommand(t, project.ID().String(), updatequalitygate.Input{})

	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err, "empty quality gate is a valid configuration (no rules)")
}

func TestHandlerExecuteDrainsEvent(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatequalitygate.NewHandler(projects)
	result, err := handler.Execute(context.Background(), mustCommand(t, project.ID().String(), strictGateInput()))
	require.NoError(t, err)

	require.NotEmpty(t, result.Events, "QualityGateUpdated event drained to caller")
	assert.Equal(t, projectmodel.EventNameQualityGateUpdated, result.Events[len(result.Events)-1].Name)

	persisted, err := projects.FindByID(context.Background(), project.ID())
	require.NoError(t, err)
	assert.Empty(t, persisted.Events(), "events drained before persistence; not retained")
}

func TestHandlerExecuteReturnsChangedTrueOnDiff(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatequalitygate.NewHandler(projects)
	result, err := handler.Execute(context.Background(), mustCommand(t, project.ID().String(), strictGateInput()))
	require.NoError(t, err)
	assert.True(t, result.Changed, "different gate should report Changed=true")
}

func TestHandlerExecuteReturnsChangedFalseWhenEqual(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatequalitygate.NewHandler(projects)

	_, err := handler.Execute(context.Background(), mustCommand(t, project.ID().String(), strictGateInput()))
	require.NoError(t, err)

	result, err := handler.Execute(context.Background(), mustCommand(t, project.ID().String(), strictGateInput()))
	require.NoError(t, err)
	assert.False(t, result.Changed, "equal gate should report Changed=false")
	assert.Empty(t, result.Events, "no events emitted for no-op update")
}

func TestNewHandlerRejectsNilRepo(t *testing.T) {
	assert.Panics(t, func() { updatequalitygate.NewHandler(nil) })
}

func mustCommand(t *testing.T, projectID string, input updatequalitygate.Input) updatequalitygate.Command {
	t.Helper()
	cmd, err := updatequalitygate.NewCommand(projectID, input)
	require.NoError(t, err)
	return cmd
}

func mustProject(t *testing.T) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject("svc", "svc", "//svc/...")
	require.NoError(t, err)
	return p
}

func TestHandlerExecuteMaxViolationsStrategy(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatequalitygate.NewHandler(projects)
	input := updatequalitygate.Input{
		Rules: []updatequalitygate.RuleInput{
			{
				Subtype: "architecture",
				Strategy: updatequalitygate.StrategyInput{
					Type:          updatequalitygate.StrategyTypeMaxViolations,
					MaxViolations: &updatequalitygate.MaxViolations{Max: 5},
				},
			},
		},
	}

	_, err := handler.Execute(context.Background(), mustCommand(t, project.ID().String(), input))
	require.NoError(t, err)

	persisted, err := projects.FindByID(context.Background(), project.ID())
	require.NoError(t, err)
	assert.Equal(t, 1, len(persisted.Gate().Rules()))
}

func TestHandlerExecuteMinNewCodeCoverageStrategy(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatequalitygate.NewHandler(projects)
	input := updatequalitygate.Input{
		Rules: []updatequalitygate.RuleInput{
			{
				Subtype: "new_code_coverage",
				Strategy: updatequalitygate.StrategyInput{
					Type:               updatequalitygate.StrategyTypeMinNewCodeCoverage,
					MinNewCodeCoverage: &updatequalitygate.MinNewCodeCoverage{Min: 80},
				},
			},
		},
	}

	_, err := handler.Execute(context.Background(), mustCommand(t, project.ID().String(), input))
	require.NoError(t, err)

	persisted, err := projects.FindByID(context.Background(), project.ID())
	require.NoError(t, err)
	assert.Equal(t, 1, len(persisted.Gate().Rules()))
}

func strictGateInput() updatequalitygate.Input {
	return updatequalitygate.Input{
		Rules: []updatequalitygate.RuleInput{
			{
				Subtype:  "code_quality",
				Strategy: updatequalitygate.StrategyInput{Type: updatequalitygate.StrategyTypeZeroTolerance},
			},
		},
	}
}
