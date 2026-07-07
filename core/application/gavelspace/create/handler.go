package create

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	gsservice "github.com/usegavel/gavel/core/domain/gavelspace/service"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type Handler struct {
	gavelspaces gsservice.GavelspaceRepository
}

func NewHandler(gavelspaces gsservice.GavelspaceRepository) *Handler {
	if gavelspaces == nil {
		panic("create: gavelspaces repository must not be nil")
	}
	return &Handler{gavelspaces: gavelspaces}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	tenantID, err := tenant.ParseTenantID(cmd.TenantID())
	if err != nil {
		return Result{}, fmt.Errorf("tenant id: %w", err)
	}

	gavelspace, err := gsmodel.NewGavelspace(tenantID, cmd.Name())
	if err != nil {
		return Result{}, fmt.Errorf("new gavelspace: %w", err)
	}

	events := gavelspace.Events()
	gavelspace.ClearEvents()

	if err := h.gavelspaces.Save(ctx, gavelspace); err != nil {
		return Result{}, fmt.Errorf("save gavelspace: %w", err)
	}

	return Result{
		Name:   gavelspace.ID().String(),
		Events: event.EventsFromDomain(events),
	}, nil
}
