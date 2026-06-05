package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type TenantRepository interface {
	Save(ctx context.Context, t tenant.Tenant) error
	ByID(ctx context.Context, id tenant.TenantID) (tenant.Tenant, error)
	BySlug(ctx context.Context, slug tenant.Slug) (tenant.Tenant, error)
}
