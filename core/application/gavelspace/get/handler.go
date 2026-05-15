package get

import "context"

type Handler struct {
	getter Finder
}

func NewHandler(getter Finder) *Handler {
	if getter == nil {
		panic("query/gavelspace/get: getter must not be nil")
	}
	return &Handler{getter: getter}
}

func (h *Handler) Execute(ctx context.Context, q Query) (*GavelspaceDetail, error) {
	return h.getter.GetByName(ctx, q.Name())
}
