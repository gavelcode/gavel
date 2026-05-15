package get

import (
	"context"

	"github.com/usegavel/gavel/core/application/project/projectview"
)

type Finder interface {
	GetByID(ctx context.Context, id string) (*projectview.ProjectDetail, error)
}
