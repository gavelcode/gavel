package getbykey

import (
	"context"

	"github.com/usegavel/gavel/core/application/project/projectview"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type Finder interface {
	GetByKey(ctx context.Context, tenantID tenant.TenantID, key string) (*projectview.ProjectDetail, error)
}
