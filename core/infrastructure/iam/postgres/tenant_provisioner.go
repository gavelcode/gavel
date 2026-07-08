package postgres

import (
	"context"
	"fmt"

	tenantmodel "github.com/usegavel/gavel/core/domain/iam/model/tenant"
	usermodel "github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type TenantProvisioner struct {
	db *database.DB
}

var _ service.TenantProvisioner = (*TenantProvisioner)(nil)

func NewTenantProvisioner(db *database.DB) *TenantProvisioner {
	return &TenantProvisioner{db: db}
}

func (p *TenantProvisioner) Provision(ctx context.Context, tenant tenantmodel.Tenant, admin usermodel.User) error {
	transaction, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin provision tx: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	if err := NewTenantRepo(transaction).Save(ctx, tenant); err != nil {
		return fmt.Errorf("save tenant: %w", err)
	}
	if err := NewUserRepo(transaction).Save(ctx, admin); err != nil {
		return fmt.Errorf("save admin: %w", err)
	}
	return transaction.Commit()
}
