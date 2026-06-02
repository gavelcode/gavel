package get

import "context"

type Finder interface {
	GetByID(ctx context.Context, id string) (*PleadingDetail, error)
}
