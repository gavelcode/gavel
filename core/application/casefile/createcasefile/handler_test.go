package createcasefile_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	casefilememory "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	projectmemory "github.com/usegavel/gavel/core/infrastructure/project/memory"
)

var testTenantExternal = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

var testTime = time.Date(2026, time.June, 10, 12, 0, 0, 0, time.UTC)

func TestExecutePersistsEmptyCaseFile(t *testing.T) {
	caseFiles, project := setup(t)

	handler := createcasefile.NewHandler(caseFiles, project)
	cmd, err := createcasefile.NewCommand(
		testTenantExternal.String(),
		seededProjectID(t, project).String(),
		"abc123",
		"main",
		testTime,
	)
	require.NoError(t, err)

	res, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	require.NotEmpty(t, res.CaseFileID)
	require.Len(t, res.Events, 1, "the CaseFileOpened domain event should be drained")

	id, err := uuid.Parse(res.CaseFileID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestExecuteFreshEvaluationOption(t *testing.T) {
	caseFiles, project := setup(t)
	handler := createcasefile.NewHandler(caseFiles, project)

	cmd, err := createcasefile.NewCommand(
		testTenantExternal.String(),
		seededProjectID(t, project).String(),
		"abc123",
		"main",
		testTime,
		createcasefile.WithFreshEvaluation(),
	)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
}

func TestExecuteProjectNotFound(t *testing.T) {
	caseFiles, project := setup(t)
	handler := createcasefile.NewHandler(caseFiles, project)

	cmd, _ := createcasefile.NewCommand(
		testTenantExternal.String(),
		projectmodel.NewProjectID(uuid.New()).String(),
		"abc123",
		"main",
		testTime,
	)

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load project")
}

func TestNewHandlerPanicsOnNilCaseFiles(t *testing.T) {
	_, projects := setup(t)
	assert.Panics(t, func() { createcasefile.NewHandler(nil, projects) })
}

func TestNewHandlerPanicsOnNilProjects(t *testing.T) {
	caseFiles, _ := setup(t)
	assert.Panics(t, func() { createcasefile.NewHandler(caseFiles, nil) })
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	_, err := createcasefile.NewCommand("", "proj", "abc", "main", testTime)
	require.ErrorIs(t, err, createcasefile.ErrInvalidCommand)

	_, err = createcasefile.NewCommand("t", "", "abc", "main", testTime)
	require.ErrorIs(t, err, createcasefile.ErrInvalidCommand)

	_, err = createcasefile.NewCommand("t", "p", "", "main", testTime)
	require.ErrorIs(t, err, createcasefile.ErrInvalidCommand)

	_, err = createcasefile.NewCommand("t", "p", "abc", "", testTime)
	require.ErrorIs(t, err, createcasefile.ErrInvalidCommand)

	_, err = createcasefile.NewCommand("t", "p", "abc", "main", time.Time{})
	require.ErrorIs(t, err, createcasefile.ErrInvalidCommand)
}

func setup(t *testing.T) (*casefilememory.CaseFileRepository, *projectmemory.ProjectRepository) {
	t.Helper()
	return casefilememory.NewCaseFileRepository(), projectmemory.NewProjectRepository()
}

func seededProjectID(t *testing.T, repo *projectmemory.ProjectRepository) projectmodel.ProjectID {
	t.Helper()
	p, err := projectmodel.NewProject(testTenantExternal, "test-project", "Test Project", "//...")
	require.NoError(t, err)
	p.ClearEvents()
	require.NoError(t, repo.Save(context.Background(), p))
	return p.ID()
}
