package iam

import (
	"context"

	tenantmodel "github.com/usegavel/gavel/core/domain/iam/model/tenant"
	usermodel "github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

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
	if err := p.users.Save(ctx, admin); err != nil {
		p.tenants.remove(tenant.ID())
		return err
	}
	return nil
}
