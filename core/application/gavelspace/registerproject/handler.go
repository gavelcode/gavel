package registerproject

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/shared/event"
	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	gsservice "github.com/usegavel/gavel/core/domain/gavelspace/service"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type Handler struct {
	gavelspaces gsservice.GavelspaceRepository
}

func NewHandler(gavelspaces gsservice.GavelspaceRepository) *Handler {
	if gavelspaces == nil {
		panic("registerproject: gavelspaces repository must not be nil")
	}
	return &Handler{gavelspaces: gavelspaces}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	name, err := gsmodel.NewGavelspaceID(cmd.GavelspaceID())
	if err != nil {
		return Result{}, fmt.Errorf("gavelspace name: %w", err)
	}

	projectID, err := projectmodel.ParseProjectID(cmd.ProjectID())
	if err != nil {
		return Result{}, fmt.Errorf("project id: %w", err)
	}

	gavelspace, err := h.gavelspaces.FindByName(ctx, name)
	if err != nil {
		return Result{}, fmt.Errorf("load gavelspace: %w", err)
	}

	ref, err := gsmodel.NewProjectRef(projectID, cmd.TargetPattern())
	if err != nil {
		return Result{}, fmt.Errorf("new project ref: %w", err)
	}

	if err := gavelspace.AddProject(ref, time.Now().UTC()); err != nil {
		return Result{}, fmt.Errorf("add project to gavelspace: %w", err)
	}

	events := gavelspace.Events()
	gavelspace.ClearEvents()

	if err := h.gavelspaces.Save(ctx, gavelspace); err != nil {
		return Result{}, fmt.Errorf("save gavelspace: %w", err)
	}

	return Result{Events: event.EventsFromDomain(events)}, nil
}
