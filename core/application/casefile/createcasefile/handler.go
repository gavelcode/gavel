package createcasefile

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/shared/event"
	casefilemodel "github.com/usegavel/gavel/core/domain/casefile/model"
	caseservice "github.com/usegavel/gavel/core/domain/casefile/service"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	projectservice "github.com/usegavel/gavel/core/domain/project/service"
)

type Handler struct {
	caseFiles caseservice.CaseFileRepository
	projects  projectservice.ProjectRepository
}

func NewHandler(caseFiles caseservice.CaseFileRepository, projects projectservice.ProjectRepository) *Handler {
	if caseFiles == nil {
		panic("createcasefile: caseFiles repository must not be nil")
	}
	if projects == nil {
		panic("createcasefile: projects repository must not be nil")
	}
	return &Handler{caseFiles: caseFiles, projects: projects}
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

	if _, err := h.projects.FindByID(ctx, tenantID, projectID); err != nil {
		return Result{}, fmt.Errorf("load project: %w", err)
	}

	caseFile, err := casefilemodel.NewCaseFile(tenantID, projectID, cmd.CommitSHA(), cmd.Branch(), cmd.StartedAt(), time.Now().UTC())
	if err != nil {
		return Result{}, fmt.Errorf("new case file: %w", err)
	}

	if cmd.FreshEvaluation() {
		caseFile.MarkFreshEvaluation()
	}

	events := caseFile.Events()
	caseFile.ClearEvents()

	if err := h.caseFiles.Save(ctx, caseFile); err != nil {
		return Result{}, fmt.Errorf("save case file: %w", err)
	}

	return Result{
		CaseFileID: caseFile.ID().String(),
		Events:     event.EventsFromDomain(events),
	}, nil
}
