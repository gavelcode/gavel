package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/gavelspace/model"
)

type GavelspaceRepository interface {
	Save(ctx context.Context, gavelspace model.Gavelspace) error
	FindByName(ctx context.Context, name model.GavelspaceID) (model.Gavelspace, error)
}
