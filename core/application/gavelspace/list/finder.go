package list

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type Finder interface {
	List(ctx context.Context, tenantID tenant.TenantID, limit, offset int) ([]GavelspaceSummary, int, error)
}
