package list

import "context"

type Handler struct {
	lister Finder
}

func NewHandler(lister Finder) *Handler {
	if lister == nil {
		panic("query/project/list: lister must not be nil")
	}
	return &Handler{lister: lister}
}

func (h *Handler) Execute(ctx context.Context, q Query) (Result, error) {
	items, total, err := h.lister.List(ctx, q.Limit(), q.Offset())
	if err != nil {
		return Result{}, err
	}
	return Result{Items: items, Total: total}, nil
}
