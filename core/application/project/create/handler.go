package create

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
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
	project, err := projectmodel.NewProject(cmd.Key(), cmd.Name(), cmd.TargetPattern())
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
