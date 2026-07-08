package loadgavelspace

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
)

type ArchPolicyLoader interface {
	LoadPolicy(workspace string) (archpolicy.Policy, error)
}

type ProjectSaver interface {
	Save(ctx context.Context, project projectmodel.Project) error
	FindByKey(ctx context.Context, tenantID tenant.TenantID, key string) (projectmodel.Project, error)
}
