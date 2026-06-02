package file_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/pleading/file"
	"github.com/usegavel/gavel/core/domain/pleading/model"
)

var testProjectID = uuid.NewString()

func validCommand(t *testing.T) file.Command {
	t.Helper()
	cmd, err := file.NewCommand(testProjectID, 42, "Add login", "alice", "feature/login", "main", "abc123")
	require.NoError(t, err)
	return cmd
}

func TestHandlerExecuteFilesPleading(t *testing.T) {
	repo := newFakePleadingRepo()
	handler := file.NewHandler(repo)

	result, err := handler.Execute(context.Background(), validCommand(t))
	require.NoError(t, err)

	assert.NotEmpty(t, result.PleadingID)
	assert.Equal(t, model.StatusOpen.String(), result.Status)
	assert.Equal(t, 1, repo.count())
}

func TestHandlerExecutePersistsAllFields(t *testing.T) {
	repo := newFakePleadingRepo()
	handler := file.NewHandler(repo)

	result, err := handler.Execute(context.Background(), validCommand(t))
	require.NoError(t, err)

	id, err := model.ParsePleadingID(result.PleadingID)
	require.NoError(t, err)

	stored, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, testProjectID, stored.ProjectID().String())
	assert.Equal(t, 42, stored.Number())
	assert.Equal(t, "Add login", stored.Title())
	assert.Equal(t, "alice", stored.Petitioner())
	assert.Equal(t, "feature/login", stored.SourceBranch())
	assert.Equal(t, "main", stored.TargetBranch())
	assert.Equal(t, "abc123", stored.CommitSHA())
	assert.True(t, stored.Status().Equal(model.StatusOpen))
}

func TestHandlerExecuteFilingDoesNotRecordDomainEvents(t *testing.T) {
	repo := newFakePleadingRepo()
	handler := file.NewHandler(repo)

	result, err := handler.Execute(context.Background(), validCommand(t))
	require.NoError(t, err)

	assert.Empty(t, result.Events, "creating a pleading records no domain events; only resolutions do")
}

func TestHandlerExecuteSaveErrorPropagates(t *testing.T) {
	repo := newFakePleadingRepo()
	repo.saveErr = errors.New("disk full")
	handler := file.NewHandler(repo)

	_, err := handler.Execute(context.Background(), validCommand(t))
	require.Error(t, err)
}

func TestHandlerExecuteRejectsZeroCommand(t *testing.T) {
	repo := newFakePleadingRepo()
	handler := file.NewHandler(repo)

	_, err := handler.Execute(context.Background(), file.Command{})
	require.Error(t, err)
	assert.Equal(t, 0, repo.count(), "no save when projectID parse fails")
}

func TestNewHandlerRejectsNilDependencies(t *testing.T) {
	assert.Panics(t, func() { file.NewHandler(nil) })
}
