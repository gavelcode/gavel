package create

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
		panic("iam/tenant/create: tenants repository must not be nil")
	}
	return &Handler{tenants: tenants}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	slug, err := tenant.NewSlug(cmd.Slug())
	if err != nil {
		return Result{}, err
	}
	newTenant, err := tenant.NewTenant(slug, cmd.DisplayName(), cmd.OccurredAt())
	if err != nil {
		return Result{}, fmt.Errorf("new tenant: %w", err)
	}

	events := newTenant.Events()
	newTenant.ClearEvents()

	if err := h.tenants.Save(ctx, newTenant); err != nil {
		return Result{}, fmt.Errorf("save tenant: %w", err)
	}

	return Result{
		TenantID:    newTenant.ID().String(),
		Slug:        newTenant.Slug().String(),
		DisplayName: newTenant.DisplayName(),
		Events:      event.EventsFromDomain(events),
	}, nil
}
