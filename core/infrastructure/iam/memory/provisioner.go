package iam

import (
	"context"

	tenantmodel "github.com/usegavel/gavel/core/domain/iam/model/tenant"
	usermodel "github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

// Provisioner is the in-memory TenantProvisioner for tests. It saves the tenant
// then the admin, stopping on the first error, so a rejected tenant (e.g. a
// taken slug) never leaves a half-provisioned admin behind — the same
// all-or-nothing contract the Postgres provisioner gives via a transaction.
type Provisioner struct {
	tenants *TenantRepository
	users   *UserRepository
}

var _ service.TenantProvisioner = (*Provisioner)(nil)

func NewProvisioner(tenants *TenantRepository, users *UserRepository) *Provisioner {
	return &Provisioner{tenants: tenants, users: users}
}

func (p *Provisioner) Provision(ctx context.Context, tenant tenantmodel.Tenant, admin usermodel.User) error {
	if err := p.tenants.Save(ctx, tenant); err != nil {
		return err
	}
	return p.users.Save(ctx, admin)
}
