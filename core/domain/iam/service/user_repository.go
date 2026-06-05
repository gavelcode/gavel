package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

type UserRepository interface {
	Save(ctx context.Context, u user.User) error
	ByID(ctx context.Context, id user.UserID) (user.User, error)
	ByEmail(ctx context.Context, tenantID tenant.TenantID, email user.Email) (user.User, error)
	CountByTenant(ctx context.Context, tenantID tenant.TenantID) (int, error)
}
