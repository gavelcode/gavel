package get

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

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
	tenantID, err := tenant.ParseTenantID(q.TenantID())
	if err != nil {
		return nil, fmt.Errorf("tenant id: %w", err)
	}
	return h.getter.GetByID(ctx, tenantID, q.ID())
}
