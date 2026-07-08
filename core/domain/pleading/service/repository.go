package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/pleading/model"
)

type PleadingRepository interface {
	Save(ctx context.Context, pleading model.Pleading) error
	FindByID(ctx context.Context, tenantID tenant.TenantID, id model.PleadingID) (model.Pleading, error)
}
