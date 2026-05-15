package loadgavelspace

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type Handler struct {
	finder       Finder
	archLoader   ArchPolicyLoader
	projectSaver ProjectSaver
	log          *slog.Logger
}

func NewHandler(finder Finder, opts ...HandlerOption) *Handler {
	if finder == nil {
		panic("loadgavelspace: finder must not be nil")
	}
	h := &Handler{finder: finder, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

type HandlerOption func(*Handler)

func WithArchPolicyLoader(loader ArchPolicyLoader) HandlerOption {
	return func(h *Handler) { h.archLoader = loader }
}

func WithProjectSaver(saver ProjectSaver) HandlerOption {
	return func(h *Handler) { h.projectSaver = saver }
}

func WithLogger(log *slog.Logger) HandlerOption {
	return func(h *Handler) { h.log = log }
}

func (h *Handler) Execute(ctx context.Context, query Query) (Result, error) {
	gavelspace, projects, err := h.finder.LoadFromConfig(query.ConfigPath())
	if err != nil {
		return Result{}, fmt.Errorf("load workspace: %w", err)
	}

	if h.archLoader != nil && query.Workspace() != "" {
		policy, err := h.archLoader.LoadPolicy(query.Workspace())
		if err != nil {
			h.log.Debug("no architecture policy loaded", "reason", err)
		} else {
			now := time.Now().UTC()
			for i := range projects {
				projects[i].UpdateArchitecturePolicy(policy, now)
			}
			h.log.Debug("architecture policy loaded", "layers", len(policy.Layers()), "rules", len(policy.DenyRules()))
		}
	}

	if query.ProjectFilter() != "" {
		filtered := filterByName(projects, query.ProjectFilter())
		if len(filtered) == 0 {
			return Result{}, fmt.Errorf("%w: project %query not found in config", ErrInvalidQuery, query.ProjectFilter())
		}
		projects = filtered
	}

	if h.projectSaver != nil {
		for index, project := range projects {
			if err := h.projectSaver.Save(ctx, project); err != nil {
				return Result{}, fmt.Errorf("save project %s: %w", project.Name(), err)
			}
			saved, err := h.projectSaver.FindByKey(ctx, project.Key())
			if err == nil {
				projects[index] = saved
			}
		}
	}

	return Result{Gavelspace: gavelspace, Projects: projects, View: buildView(gavelspace, projects)}, nil
}

func filterByName(projects []projectmodel.Project, name string) []projectmodel.Project {
	for _, p := range projects {
		if p.Name() == name {
			return []projectmodel.Project{p}
		}
	}
	return nil
}
