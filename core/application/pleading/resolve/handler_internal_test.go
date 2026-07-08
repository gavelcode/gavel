package resolve

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/pleading/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type stubRepo struct {
	pleading model.Pleading
	findErr  error
}

func (s *stubRepo) Save(_ context.Context, _ model.Pleading) error { return nil }
func (s *stubRepo) FindByID(_ context.Context, _ tenant.TenantID, _ model.PleadingID) (model.Pleading, error) {
	if s.findErr != nil {
		return model.Pleading{}, s.findErr
	}
	return s.pleading, nil
}

func TestExecuteInvalidOutcomeStatus(t *testing.T) {
	projectID := projectmodel.NewProjectID(uuid.New())
	pleading, err := model.FilePleading(tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222")), projectID, 1, "title", "alice", "src", "dst", "sha")
	require.NoError(t, err)

	handler := &Handler{pleadings: &stubRepo{pleading: pleading}}

	cmd := Command{
		pleadingID: pleading.ID().String(),
		outcome:    "bogus",
	}

	_, err = handler.Execute(context.Background(), cmd)
	assert.Error(t, err)
}

func TestApplyTransitionUnsupportedStatus(t *testing.T) {
	projectID := projectmodel.NewProjectID(uuid.New())
	pleading, err := model.FilePleading(tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222")), projectID, 1, "title", "alice", "src", "dst", "sha")
	require.NoError(t, err)

	err = applyTransition(&pleading, model.StatusOpen, time.Now())
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrInvalidTransition)
}
