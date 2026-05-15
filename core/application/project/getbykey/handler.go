package getbykey

import "context"

type Handler struct {
	getter Finder
}

func NewHandler(getter Finder) *Handler {
	if getter == nil {
		panic("query/project/getbykey: getter must not be nil")
	}
	return &Handler{getter: getter}
}

func (h *Handler) Execute(ctx context.Context, q Query) (*ProjectDetail, error) {
	return h.getter.GetByKey(ctx, q.Key())
}
