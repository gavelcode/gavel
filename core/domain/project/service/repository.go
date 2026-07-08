package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/project/model"
)

type ProjectRepository interface {
	Save(ctx context.Context, project model.Project) error
	FindByID(ctx context.Context, tenantID tenant.TenantID, id model.ProjectID) (model.Project, error)
	FindByName(ctx context.Context, tenantID tenant.TenantID, name string) (model.Project, error)
	FindByKey(ctx context.Context, tenantID tenant.TenantID, key string) (model.Project, error)
}
