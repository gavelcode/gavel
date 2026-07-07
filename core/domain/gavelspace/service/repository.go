package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type GavelspaceRepository interface {
	Save(ctx context.Context, gavelspace model.Gavelspace) error
	FindByName(ctx context.Context, tenantID tenant.TenantID, name model.GavelspaceID) (model.Gavelspace, error)
}
