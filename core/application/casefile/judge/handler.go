package judge

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/shared/event"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/tracking"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	caseservice "github.com/usegavel/gavel/core/domain/casefile/service"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectservice "github.com/usegavel/gavel/core/domain/project/service"
)

type Handler struct {
	caseFiles caseservice.CaseFileRepository
	projects  projectservice.ProjectRepository
}

func NewHandler(caseFiles caseservice.CaseFileRepository, projects projectservice.ProjectRepository) *Handler {
	if caseFiles == nil {
		panic("judge: caseFiles repository must not be nil")
	}
	if projects == nil {
		panic("judge: projects repository must not be nil")
	}
	return &Handler{caseFiles: caseFiles, projects: projects}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	tenantID, err := tenant.ParseTenantID(cmd.TenantID())
	if err != nil {
		return Result{}, fmt.Errorf("tenant id: %w", err)
	}

	caseFileID, err := casefile.ParseCaseFileID(cmd.CaseFileID())
	if err != nil {
		return Result{}, fmt.Errorf("reconstitute case file ID: %w", err)
	}

	caseFile, err := h.caseFiles.FindByID(ctx, tenantID, caseFileID)
	if err != nil {
		return Result{}, fmt.Errorf("load case file: %w", err)
	}

	project, err := h.projects.FindByID(ctx, caseFile.TenantID(), caseFile.ProjectID())
	if err != nil {
		return Result{}, fmt.Errorf("load project: %w", err)
	}

	var domainTracking *tracking.Result
	if cmd.Tracking() != nil {
		t, err := evidencedto.TrackingToDomain(*cmd.Tracking())
		if err != nil {
			return Result{}, fmt.Errorf("tracking: %w", err)
		}
		domainTracking = &t
	}

	verdictResult, err := caseFile.Judge(project.Gate(), domainTracking, time.Now().UTC(), cmd.DeltaInput())
	if err != nil {
		return Result{}, fmt.Errorf("execute judge: %w", err)
	}

	events := caseFile.Events()
	caseFile.ClearEvents()

	if err := h.caseFiles.Save(ctx, caseFile); err != nil {
		return Result{}, fmt.Errorf("save case file: %w", err)
	}

	return Result{
		CaseFileID: cmd.CaseFileID(),
		Verdict:    toVerdictView(verdictResult),
		Events:     event.EventsFromDomain(events),
	}, nil
}

func toVerdictView(verdictResult verdict.Result) VerdictView {
	rulings := verdictResult.Rulings()
	out := make([]RulingView, 0, len(rulings))
	for _, r := range rulings {
		out = append(out, RulingView{
			Subtype: r.Subtype().String(),
			Passed:  r.Passed(),
			Detail:  r.Detail(),
		})
	}
	return VerdictView{
		Outcome:     verdictResult.Outcome().String(),
		Rulings:     out,
		EvaluatedAt: verdictResult.EvaluatedAt(),
	}
}
