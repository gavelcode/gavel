package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/pleading/model"
	"github.com/usegavel/gavel/core/domain/pleading/service"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

var _ service.PleadingRepository = (*PleadingRepository)(nil)

var ErrPleadingNotFound = failure.New("pleading not found", failure.NotFound)

type PleadingRepository struct {
	mu       sync.RWMutex
	byTenant map[string]map[string]model.Pleading
}

func NewPleadingRepository() *PleadingRepository {
	return &PleadingRepository{
		byTenant: make(map[string]map[string]model.Pleading),
	}
}

func (r *PleadingRepository) Save(_ context.Context, pleading model.Pleading) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tenantKey := pleading.TenantID().String()
	if r.byTenant[tenantKey] == nil {
		r.byTenant[tenantKey] = make(map[string]model.Pleading)
	}
	r.byTenant[tenantKey][pleading.ID().String()] = pleading
	return nil
}

func (r *PleadingRepository) FindByID(_ context.Context, tenantID tenant.TenantID, id model.PleadingID) (model.Pleading, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byTenant[tenantID.String()][id.String()]
	if !ok {
		return model.Pleading{}, fmt.Errorf("%w: %s", ErrPleadingNotFound, id)
	}
	return p, nil
}
