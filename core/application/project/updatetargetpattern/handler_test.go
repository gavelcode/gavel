package updatetargetpattern_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/project/updatetargetpattern"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func TestHandlerExecuteUpdatesPattern(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatetargetpattern.NewHandler(projects)
	cmd := mustCommand(t, project.ID().String(), "//svc/v2/...")

	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	persisted, err := projects.FindByID(context.Background(), testTenant, project.ID())
	require.NoError(t, err)
	assert.Equal(t, "//svc/v2/...", persisted.TargetPattern())
}

func TestHandlerExecuteInvalidProjectID(t *testing.T) {
	projects := newFakeProjectRepo()
	handler := updatetargetpattern.NewHandler(projects)
	cmd := mustCommand(t, "missing", "//svc/...")

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteProjectNotFound(t *testing.T) {
	projects := newFakeProjectRepo()
	handler := updatetargetpattern.NewHandler(projects)

	cmd, err := updatetargetpattern.NewCommand(testTenant.String(), "11111111-1111-1111-1111-111111111111", "//svc/...")
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteInvalidPatternRejected(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatetargetpattern.NewHandler(projects)
	cmd, err := updatetargetpattern.NewCommand(testTenant.String(), project.ID().String(), "not-a-pattern")
	require.NoError(t, err, "command-layer accepts non-empty; domain enforces shape")

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, projectmodel.ErrInvalidProject)

	persisted, err := projects.FindByID(context.Background(), testTenant, project.ID())
	require.NoError(t, err)
	assert.Equal(t, "//svc/...", persisted.TargetPattern(), "invalid update does not mutate stored pattern")
}

func TestHandlerExecuteSaveErrorPropagated(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)
	projects.saveErr = errors.New("disk full")

	handler := updatetargetpattern.NewHandler(projects)
	cmd := mustCommand(t, project.ID().String(), "//svc/v2/...")

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteDrainsEvent(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatetargetpattern.NewHandler(projects)
	result, err := handler.Execute(context.Background(), mustCommand(t, project.ID().String(), "//svc/v2/..."))
	require.NoError(t, err)

	require.NotEmpty(t, result.Events, "TargetPatternUpdated event drained to caller")
	assert.Equal(t, projectmodel.EventNameTargetPatternUpdated, result.Events[len(result.Events)-1].Name)

	persisted, err := projects.FindByID(context.Background(), testTenant, project.ID())
	require.NoError(t, err)
	assert.Empty(t, persisted.Events(), "events drained before persistence; not retained")
}

func TestNewHandlerRejectsNilRepo(t *testing.T) {
	assert.Panics(t, func() { updatetargetpattern.NewHandler(nil) })
}

func mustCommand(t *testing.T, projectID, pattern string) updatetargetpattern.Command {
	t.Helper()
	cmd, err := updatetargetpattern.NewCommand(testTenant.String(), projectID, pattern)
	require.NoError(t, err)
	return cmd
}

func mustProject(t *testing.T) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject(testTenant, "svc", "svc", "//svc/...")
	require.NoError(t, err)
	return p
}
