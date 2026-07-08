package iam

import (
	"context"
	"fmt"
	"sync"

	tenantmodel "github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

var _ service.TenantRepository = (*TenantRepository)(nil)

type TenantRepository struct {
	mu     sync.RWMutex
	byID   map[string]tenantmodel.Tenant
	bySlug map[string]string
}

func NewTenantRepository() *TenantRepository {
	return &TenantRepository{
		byID:   make(map[string]tenantmodel.Tenant),
		bySlug: make(map[string]string),
	}
}

func (r *TenantRepository) Save(_ context.Context, tenant tenantmodel.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	slug := tenant.Slug().String()
	if existingID, taken := r.bySlug[slug]; taken && existingID != tenant.ID().String() {
		return fmt.Errorf("%w: %s", tenantmodel.ErrSlugTaken, slug)
	}
	r.byID[tenant.ID().String()] = tenant
	r.bySlug[slug] = tenant.ID().String()
	return nil
}

func (r *TenantRepository) remove(tenantID tenantmodel.TenantID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.byID[tenantID.String()]
	if !ok {
		return
	}
	delete(r.bySlug, stored.Slug().String())
	delete(r.byID, tenantID.String())
}

func (r *TenantRepository) ByID(_ context.Context, id tenantmodel.TenantID) (tenantmodel.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tenant, ok := r.byID[id.String()]
	if !ok {
		return tenantmodel.Tenant{}, fmt.Errorf("%w: %s", tenantmodel.ErrTenantNotFound, id.String())
	}
	return tenant, nil
}

func (r *TenantRepository) BySlug(_ context.Context, slug tenantmodel.Slug) (tenantmodel.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.bySlug[slug.String()]
	if !ok {
		return tenantmodel.Tenant{}, fmt.Errorf("%w: %s", tenantmodel.ErrTenantNotFound, slug.String())
	}
	return r.byID[id], nil
}
