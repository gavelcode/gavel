package create

import (
	"context"
	"fmt"

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
		panic("create: projects repository must not be nil")
	}
	return &Handler{projects: projects}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	tenantID, err := tenant.ParseTenantID(cmd.TenantID())
	if err != nil {
		return Result{}, fmt.Errorf("tenant id: %w", err)
	}

	project, err := projectmodel.NewProject(tenantID, cmd.Key(), cmd.Name(), cmd.TargetPattern())
	if err != nil {
		return Result{}, fmt.Errorf("new project: %w", err)
	}

	events := project.Events()
	project.ClearEvents()

	if err := h.projects.Save(ctx, project); err != nil {
		return Result{}, fmt.Errorf("save project: %w", err)
	}

	return Result{
		ProjectID: project.ID().String(),
		Events:    event.EventsFromDomain(events),
	}, nil
}
