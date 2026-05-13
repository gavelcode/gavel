package get

import "context"

type Handler struct {
	getter Finder
}

func NewHandler(getter Finder) *Handler {
	if getter == nil {
		panic("query/casefile/get: getter must not be nil")
	}
	return &Handler{getter: getter}
}

func (h *Handler) Execute(ctx context.Context, q Query) (*CaseFileDetail, error) {
	return h.getter.GetByID(ctx, q.ID())
}
