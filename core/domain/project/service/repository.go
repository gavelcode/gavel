package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/project/model"
)

type ProjectRepository interface {
	Save(ctx context.Context, project model.Project) error
	FindByID(ctx context.Context, id model.ProjectID) (model.Project, error)
	FindByName(ctx context.Context, name string) (model.Project, error)
	FindByKey(ctx context.Context, key string) (model.Project, error)
}
