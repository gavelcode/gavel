package registerproject

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
)

type stubRepo struct {
	gs      gsmodel.Gavelspace
	findErr error
}

func (s *stubRepo) FindByName(_ context.Context, _ gsmodel.GavelspaceID) (gsmodel.Gavelspace, error) {
	if s.findErr != nil {
		return gsmodel.Gavelspace{}, s.findErr
	}
	return s.gs, nil
}

func (s *stubRepo) Save(_ context.Context, _ gsmodel.Gavelspace) error { return nil }

func TestExecuteInvalidGavelspaceNameFromDomain(t *testing.T) {
	handler := &Handler{gavelspaces: &stubRepo{}}

	cmd := Command{
		gavelspaceName: "   ",
		projectID:      uuid.NewString(),
		targetPattern:  "//svc/...",
	}

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gavelspace name")
}

func TestExecuteInvalidTargetPatternFromDomain(t *testing.T) {
	gs, err := gsmodel.NewGavelspace("alpha")
	require.NoError(t, err)

	handler := &Handler{gavelspaces: &stubRepo{gs: gs}}

	cmd := Command{
		gavelspaceName: "alpha",
		projectID:      uuid.NewString(),
		targetPattern:  "   ",
	}

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new project ref")
}
