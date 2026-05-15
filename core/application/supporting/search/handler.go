package search

import "context"

type Handler struct {
	query Finder
}

func NewHandler(query Finder) *Handler {
	if query == nil {
		panic("query/search: query must not be nil")
	}
	return &Handler{query: query}
}

func (h *Handler) Execute(ctx context.Context, q Query) (Result, error) {
	items, err := h.query.Search(ctx, q.Text(), q.Limit())
	if err != nil {
		return Result{}, err
	}
	return Result{Items: items}, nil
}
