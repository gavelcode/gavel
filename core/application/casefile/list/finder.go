package list

import "context"

type Finder interface {
	ListByProject(ctx context.Context, tenantID, projectID, gavelspace string, limit, offset int) ([]CaseFileSummary, int, error)
}
