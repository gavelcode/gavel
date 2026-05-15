package get

import "context"

type Finder interface {
	GetByName(ctx context.Context, name string) (*GavelspaceDetail, error)
}
