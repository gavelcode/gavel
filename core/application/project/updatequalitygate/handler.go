package updatequalitygate

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/shared/event"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	projectservice "github.com/usegavel/gavel/core/domain/project/service"
)

type Handler struct {
	projects projectservice.ProjectRepository
}

func NewHandler(projects projectservice.ProjectRepository) *Handler {
	if projects == nil {
		panic("updatequalitygate: projects repository must not be nil")
	}
	return &Handler{projects: projects}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	projectID, err := projectmodel.ParseProjectID(cmd.ProjectID())
	if err != nil {
		return Result{}, fmt.Errorf("project id: %w", err)
	}

	project, err := h.projects.FindByID(ctx, projectID)
	if err != nil {
		return Result{}, fmt.Errorf("load project: %w", err)
	}

	if project.Gate().Equal(cmd.Gate()) {
		return Result{Changed: false}, nil
	}

	project.UpdateQualityGate(cmd.Gate(), time.Now().UTC())

	events := project.Events()
	project.ClearEvents()

	if err := h.projects.Save(ctx, project); err != nil {
		return Result{}, fmt.Errorf("save project: %w", err)
	}

	return Result{Changed: true, Events: event.EventsFromDomain(events)}, nil
}
