package listfindings

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type Handler struct {
	lister Finder
}

func NewHandler(lister Finder) *Handler {
	if lister == nil {
		panic("query/finding/list: lister must not be nil")
	}
	return &Handler{lister: lister}
}

func (h *Handler) Execute(ctx context.Context, q Query) (Result, error) {
	tenantID, err := tenant.ParseTenantID(q.TenantID())
	if err != nil {
		return Result{}, fmt.Errorf("tenant id: %w", err)
	}
	items, total, err := h.lister.List(ctx, tenantID, q.Filters(), q.Limit(), q.Offset())
	if err != nil {
		return Result{}, err
	}
	return Result{Items: items, Total: total}, nil
}
