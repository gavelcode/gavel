package list

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type Finder interface {
	ListByProject(ctx context.Context, tenantID tenant.TenantID, projectID, status, gavelspace string, limit, offset int) ([]PleadingSummary, int, error)
}
