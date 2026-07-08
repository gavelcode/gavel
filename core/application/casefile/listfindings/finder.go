package listfindings

import "context"

type Finder interface {
	List(ctx context.Context, tenantID string, filters Filters, limit, offset int) ([]FindingView, int, error)
}
