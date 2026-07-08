package listfindings

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type Finder interface {
	List(ctx context.Context, tenantID tenant.TenantID, filters Filters, limit, offset int) ([]FindingView, int, error)
}
