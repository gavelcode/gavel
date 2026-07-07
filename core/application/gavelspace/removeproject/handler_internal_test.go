package removeproject

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type stubRepo struct {
	gs      gsmodel.Gavelspace
	findErr error
}

func (s *stubRepo) FindByName(_ context.Context, _ tenant.TenantID, _ gsmodel.GavelspaceID) (gsmodel.Gavelspace, error) {
	if s.findErr != nil {
		return gsmodel.Gavelspace{}, s.findErr
	}
	return s.gs, nil
}

func (s *stubRepo) Save(_ context.Context, _ gsmodel.Gavelspace) error { return nil }

func TestExecuteInvalidGavelspaceNameFromDomain(t *testing.T) {
	handler := &Handler{gavelspaces: &stubRepo{}}

	cmd := Command{
		tenantID:       testTenant,
		gavelspaceName: "   ",
		projectID:      uuid.NewString(),
	}

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gavelspace name")
}

const testTenant = "22222222-2222-2222-2222-222222222222"
