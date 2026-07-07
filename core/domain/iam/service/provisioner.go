package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

// TenantProvisioner persists a tenant together with its first administrator as
// one atomic unit. A tenant without an admin is unusable, so the two are
// committed together or not at all — the invariant the provision use case
// exists to enforce. Mirrors Vernon's TenantProvisioningService: the domain
// builds both aggregates, this port owns the single transaction that saves them.
type TenantProvisioner interface {
	Provision(ctx context.Context, t tenant.Tenant, admin user.User) error
}
