package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/gavelspace/service"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

var _ service.GavelspaceRepository = (*GavelspaceRepository)(nil)

var ErrGavelspaceNotFound = failure.New("gavelspace not found", failure.NotFound)

type GavelspaceRepository struct {
	mu       sync.RWMutex
	byTenant map[string]map[string]model.Gavelspace
}

func NewGavelspaceRepository() *GavelspaceRepository {
	return &GavelspaceRepository{
		byTenant: make(map[string]map[string]model.Gavelspace),
	}
}

func (r *GavelspaceRepository) Save(_ context.Context, gspace model.Gavelspace) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tenantKey := gspace.TenantID().String()
	if r.byTenant[tenantKey] == nil {
		r.byTenant[tenantKey] = make(map[string]model.Gavelspace)
	}
	r.byTenant[tenantKey][gspace.ID().String()] = gspace
	return nil
}

func (r *GavelspaceRepository) FindByName(_ context.Context, tenantID tenant.TenantID, name model.GavelspaceID) (model.Gavelspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	gspace, ok := r.byTenant[tenantID.String()][name.String()]
	if !ok {
		return model.Gavelspace{}, fmt.Errorf("%w: %s", ErrGavelspaceNotFound, name.String())
	}
	return gspace, nil
}
