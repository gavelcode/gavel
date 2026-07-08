package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

type TenantProvisioner interface {
	Provision(ctx context.Context, t tenant.Tenant, admin user.User) error
}
