package listfindings

import "context"

type Finder interface {
	List(ctx context.Context, filters Filters, limit, offset int) ([]FindingView, int, error)
}
