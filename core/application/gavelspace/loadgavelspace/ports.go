package loadgavelspace

import (
	"context"

	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type ArchPolicyLoader interface {
	LoadPolicy(workspace string) (archpolicy.Policy, error)
}

type ProjectSaver interface {
	Save(ctx context.Context, project projectmodel.Project) error
	FindByKey(ctx context.Context, key string) (projectmodel.Project, error)
}
