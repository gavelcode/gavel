package getbaseline

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/project/getbykey"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	projectservice "github.com/usegavel/gavel/core/domain/project/service"
)

type Handler struct {
	byKey    getbykey.Finder
	projects projectservice.ProjectRepository
}

func NewHandler(byKey getbykey.Finder, projects projectservice.ProjectRepository) *Handler {
	if byKey == nil {
		panic("getbaseline: byKey finder must not be nil")
	}
	if projects == nil {
		panic("getbaseline: projects repository must not be nil")
	}
	return &Handler{byKey: byKey, projects: projects}
}

func (h *Handler) Execute(ctx context.Context, query Query) (Result, error) {
	detail, err := h.byKey.GetByKey(ctx, query.Key())
	if err != nil {
		return Result{}, err
	}

	branch := query.Branch()
	if branch == "" {
		branch = detail.DefaultBranch
	}

	projectID, err := projectmodel.ParseProjectID(detail.ID)
	if err != nil {
		return Result{}, fmt.Errorf("parse project id: %w", err)
	}

	project, err := h.projects.FindByID(ctx, projectID)
	if err != nil {
		return Result{}, err
	}

	bl := project.Baseline(branch)
	return Result{
		Fingerprints: nonNilStrings(bl.Fingerprints()),
		ArchIDs:      nonNilStrings(bl.ArchIDs()),
		HasPrevious:  bl.HasPrevious(),
	}, nil
}

func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
