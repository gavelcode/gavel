package resolve_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/pleading/resolve"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/pleading/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

var (
	testTime = time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
)

func seedPleading(t *testing.T) model.Pleading {
	t.Helper()
	projectID := projectmodel.NewProjectID(uuid.New())
	pleading, err := model.FilePleading(tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222")), projectID, 1, "title", "alice", "src", "dst", "sha")
	require.NoError(t, err)
	return pleading
}

func cmdFor(t *testing.T, id, outcome string) resolve.Command {
	t.Helper()
	cmd, err := resolve.NewCommand("22222222-2222-2222-2222-222222222222", id, outcome)
	require.NoError(t, err)
	return cmd
}

func TestHandlerResolveMerged(t *testing.T) {
	repo := newFakePleadingRepo()
	pleading := seedPleading(t)
	repo.seed(pleading)
	handler := resolve.NewHandler(repo)

	result, err := handler.Execute(context.Background(), cmdFor(t, pleading.ID().String(), "merged"))
	require.NoError(t, err)

	assert.True(t, result.Changed)
	assert.Equal(t, model.StatusMerged.String(), result.Status)
	require.Len(t, result.Events, 1)
	assert.Equal(t, model.EventNameMerged, result.Events[0].Name)

	stored, err := repo.FindByID(context.Background(), tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222")), pleading.ID())
	require.NoError(t, err)
	assert.True(t, stored.Status().Equal(model.StatusMerged))
}

func TestHandlerResolveClosed(t *testing.T) {
	repo := newFakePleadingRepo()
	pleading := seedPleading(t)
	repo.seed(pleading)
	handler := resolve.NewHandler(repo)

	result, err := handler.Execute(context.Background(), cmdFor(t, pleading.ID().String(), "closed"))
	require.NoError(t, err)

	assert.True(t, result.Changed)
	assert.Equal(t, model.StatusClosed.String(), result.Status)
	require.Len(t, result.Events, 1)
	assert.Equal(t, model.EventNameClosed, result.Events[0].Name)
}

func TestHandlerResolveSameTargetIsNoOp(t *testing.T) {
	repo := newFakePleadingRepo()
	pleading := seedPleading(t)
	require.NoError(t, pleading.MarkMerged(testTime))
	pleading.ClearEvents()
	repo.seed(pleading)
	handler := resolve.NewHandler(repo)

	result, err := handler.Execute(context.Background(), cmdFor(t, pleading.ID().String(), "merged"))
	require.NoError(t, err)

	assert.False(t, result.Changed)
	assert.Equal(t, model.StatusMerged.String(), result.Status)
	assert.Empty(t, result.Events, "no events recorded on idempotent no-op")
}

func TestHandlerResolveCrossTerminalRejected(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*model.Pleading) error
		targetOutc  string
		targetState model.Status
	}{
		{name: "merged then close", setup: func(pleading *model.Pleading) error { return pleading.MarkMerged(testTime) }, targetOutc: "closed", targetState: model.StatusMerged},
		{name: "closed then merge", setup: func(pleading *model.Pleading) error { return pleading.MarkClosed(testTime) }, targetOutc: "merged", targetState: model.StatusClosed},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			repo := newFakePleadingRepo()
			pleading := seedPleading(t)
			require.NoError(t, testCase.setup(&pleading))
			pleading.ClearEvents()
			repo.seed(pleading)
			handler := resolve.NewHandler(repo)

			_, err := handler.Execute(context.Background(), cmdFor(t, pleading.ID().String(), testCase.targetOutc))
			require.Error(t, err)
			assert.ErrorIs(t, err, model.ErrInvalidTransition)

			stored, err := repo.FindByID(context.Background(), tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222")), pleading.ID())
			require.NoError(t, err)
			assert.True(t, stored.Status().Equal(testCase.targetState), "state unchanged after rejected transition")
		})
	}
}

func TestHandlerResolveNotFoundPropagates(t *testing.T) {
	repo := newFakePleadingRepo()
	handler := resolve.NewHandler(repo)

	id := model.NewPleadingID(uuid.New())

	_, err := handler.Execute(context.Background(), cmdFor(t, id.String(), "merged"))
	require.Error(t, err)
}

func TestHandlerResolveSaveErrorPropagates(t *testing.T) {
	repo := newFakePleadingRepo()
	repo.saveErr = errors.New("disk full")
	pleading := seedPleading(t)
	repo.seed(pleading)
	handler := resolve.NewHandler(repo)

	_, err := handler.Execute(context.Background(), cmdFor(t, pleading.ID().String(), "merged"))
	require.Error(t, err)
}

func TestHandlerResolveInvalidPleadingIDPropagates(t *testing.T) {
	repo := newFakePleadingRepo()
	handler := resolve.NewHandler(repo)

	_, err := handler.Execute(context.Background(), resolve.Command{})
	require.Error(t, err)
}

func TestNewHandlerRejectsNilRepo(t *testing.T) {
	assert.Panics(t, func() { resolve.NewHandler(nil) })
}
