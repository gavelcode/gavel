package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/pleading/model"
)

type PleadingRepository interface {
	Save(ctx context.Context, pleading model.Pleading) error
	FindByID(ctx context.Context, id model.PleadingID) (model.Pleading, error)
}
