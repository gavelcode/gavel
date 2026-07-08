package updatelanguages_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/project/updatelanguages"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func TestHandlerExecuteSuccessful(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatelanguages.NewHandler(projects)
	cmd := mustCommand(t, project.ID().String(), []string{"java", "go"}, testTenant.String())

	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	persisted, err := projects.FindByID(context.Background(), testTenant, project.ID())
	require.NoError(t, err)
	langs := persisted.Languages()
	require.Len(t, langs, 2)
	assert.Equal(t, "java", langs[0].String())
	assert.Equal(t, "go", langs[1].String())
}

func TestHandlerExecuteEmptyLanguagesAccepted(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	require.NoError(t, projects.Save(context.Background(), project))

	handler := updatelanguages.NewHandler(projects)
	cmd := mustCommand(t, project.ID().String(), nil, testTenant.String())

	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	persisted, err := projects.FindByID(context.Background(), testTenant, project.ID())
	require.NoError(t, err)
	assert.Empty(t, persisted.Languages())
}

func TestHandlerExecuteInvalidProjectID(t *testing.T) {
	projects := newFakeProjectRepo()
	handler := updatelanguages.NewHandler(projects)
	cmd := mustCommand(t, "missing", nil, testTenant.String())

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteProjectNotFound(t *testing.T) {
	projects := newFakeProjectRepo()
	handler := updatelanguages.NewHandler(projects)

	cmd, err := updatelanguages.NewCommand(testTenant.String(), "11111111-1111-1111-1111-111111111111", []string{"go"})
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteSaveErrorPropagated(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)
	projects.saveErr = errors.New("disk full")

	handler := updatelanguages.NewHandler(projects)
	cmd := mustCommand(t, project.ID().String(), []string{"go"}, testTenant.String())

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteDrainsEvent(t *testing.T) {
	projects := newFakeProjectRepo()
	project := mustProject(t)
	projects.seed(project)

	handler := updatelanguages.NewHandler(projects)
	result, err := handler.Execute(context.Background(), mustCommand(t, project.ID().String(), []string{"go"}, testTenant.String()))
	require.NoError(t, err)

	require.NotEmpty(t, result.Events, "LanguagesUpdated event drained to caller")
	assert.Equal(t, projectmodel.EventNameLanguagesUpdated, result.Events[len(result.Events)-1].Name)

	persisted, err := projects.FindByID(context.Background(), testTenant, project.ID())
	require.NoError(t, err)
	assert.Empty(t, persisted.Events(), "events drained before persistence; not retained")
}

func TestNewHandlerRejectsNilRepo(t *testing.T) {
	assert.Panics(t, func() { updatelanguages.NewHandler(nil) })
}

func mustCommand(t *testing.T, projectID string, langs []string, tenantID string) updatelanguages.Command {
	t.Helper()
	cmd, err := updatelanguages.NewCommand(tenantID, projectID, langs)
	require.NoError(t, err)
	return cmd
}

func mustProject(t *testing.T) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject(testTenant, "svc", "svc", "//svc/...")
	require.NoError(t, err)
	return p
}
