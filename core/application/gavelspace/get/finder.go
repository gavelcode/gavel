package get

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type Finder interface {
	GetByName(ctx context.Context, tenantID tenant.TenantID, name string) (*GavelspaceDetail, error)
}
