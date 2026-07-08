package updatetargetpattern

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	projectservice "github.com/usegavel/gavel/core/domain/project/service"
)

type Handler struct {
	projects projectservice.ProjectRepository
}

func NewHandler(projects projectservice.ProjectRepository) *Handler {
	if projects == nil {
		panic("updatetargetpattern: projects repository must not be nil")
	}
	return &Handler{projects: projects}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	tenantID, err := tenant.ParseTenantID(cmd.TenantID())
	if err != nil {
		return Result{}, fmt.Errorf("tenant id: %w", err)
	}

	projectID, err := projectmodel.ParseProjectID(cmd.ProjectID())
	if err != nil {
		return Result{}, fmt.Errorf("project id: %w", err)
	}

	project, err := h.projects.FindByID(ctx, tenantID, projectID)
	if err != nil {
		return Result{}, fmt.Errorf("load project: %w", err)
	}

	if err := project.UpdateTargetPattern(cmd.TargetPattern(), time.Now().UTC()); err != nil {
		return Result{}, fmt.Errorf("update target pattern: %w", err)
	}

	events := project.Events()
	project.ClearEvents()

	if err := h.projects.Save(ctx, project); err != nil {
		return Result{}, fmt.Errorf("save project: %w", err)
	}

	return Result{Events: event.EventsFromDomain(events)}, nil
}
