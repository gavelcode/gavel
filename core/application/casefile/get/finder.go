package get

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type Finder interface {
	GetByID(ctx context.Context, tenantID tenant.TenantID, id string) (*CaseFileDetail, error)
}
