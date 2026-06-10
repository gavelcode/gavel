package search

import (
	"context"
	"strings"

	searchquery "github.com/usegavel/gavel/core/application/supporting/search"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
)

type Deps struct {
	Search *searchquery.Handler
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) Search(ctx context.Context, req gen.SearchRequestObject) (gen.SearchResponseObject, error) {
	query := ""
	if req.Params.Q != nil {
		query = *req.Params.Q
	}
	if strings.TrimSpace(query) == "" {
		return gen.Search200JSONResponse{Results: []gen.SearchResult{}}, nil
	}
	limit := 10
	if req.Params.Limit != nil && *req.Params.Limit > 0 {
		limit = *req.Params.Limit
	}
	q, err := searchquery.NewQuery(query, limit)
	if err != nil {
		return nil, err
	}
	res, err := h.deps.Search.Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	items := make([]gen.SearchResult, 0, len(res.Items))
	for _, sr := range res.Items {
		items = append(items, resultFromQuery(sr))
	}
	return gen.Search200JSONResponse{Results: items}, nil
}
