package getbykey

import (
	"context"

	"github.com/usegavel/gavel/core/application/project/projectview"
)

type Finder interface {
	GetByKey(ctx context.Context, key string) (*projectview.ProjectDetail, error)
}
