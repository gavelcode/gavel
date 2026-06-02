package list

import "context"

type Finder interface {
	ListByProject(ctx context.Context, projectID, status, gavelspace string, limit, offset int) ([]PleadingSummary, int, error)
}
