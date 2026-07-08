package get

import "context"

type Finder interface {
	GetByID(ctx context.Context, tenantID, id string) (*CaseFileDetail, error)
}
