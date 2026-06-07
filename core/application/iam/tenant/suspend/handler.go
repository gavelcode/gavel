package suspend

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type Handler struct {
	tenants service.TenantRepository
}

func NewHandler(tenants service.TenantRepository) *Handler {
	if tenants == nil {
		panic("iam/tenant/suspend: tenants repository must not be nil")
	}
	return &Handler{tenants: tenants}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	tenantID, err := tenant.ParseTenantID(cmd.TenantID())
	if err != nil {
		return Result{}, err
	}
	foundTenant, err := h.tenants.ByID(ctx, tenantID)
	if err != nil {
		return Result{}, fmt.Errorf("find tenant: %w", err)
	}
	if err := foundTenant.Suspend(cmd.OccurredAt()); err != nil {
		return Result{}, fmt.Errorf("suspend tenant: %w", err)
	}

	events := foundTenant.Events()
	foundTenant.ClearEvents()

	if err := h.tenants.Save(ctx, foundTenant); err != nil {
		return Result{}, fmt.Errorf("save tenant: %w", err)
	}

	return Result{
		TenantID: foundTenant.ID().String(),
		Events:   event.EventsFromDomain(events),
	}, nil
}
