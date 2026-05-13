package list

import "context"

type Finder interface {
	ListByProject(ctx context.Context, projectID, gavelspace string, limit, offset int) ([]CaseFileSummary, int, error)
}
