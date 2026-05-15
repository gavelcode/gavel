package list

import "context"

type Finder interface {
	List(ctx context.Context, limit, offset int) ([]GavelspaceSummary, int, error)
}
