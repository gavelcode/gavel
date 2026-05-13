package classify

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/domain/casefile/model/tracking"
	caseservice "github.com/usegavel/gavel/core/domain/casefile/service"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type Handler struct {
	caseFiles caseservice.CaseFileRepository
}

func NewHandler(caseFiles caseservice.CaseFileRepository) *Handler {
	if caseFiles == nil {
		panic("classify: caseFiles repository must not be nil")
	}
	return &Handler{caseFiles: caseFiles}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	projectID, err := projectmodel.ParseProjectID(cmd.ProjectID())
	if err != nil {
		return Result{}, fmt.Errorf("project id: %w", err)
	}

	previous, err := h.caseFiles.FindFingerprintIDsByBranch(ctx, projectID, cmd.Branch())
	if err != nil {
		return Result{}, fmt.Errorf("load previous fingerprints: %w", err)
	}

	tracking := tracking.ClassifyFindings(cmd.Findings(), previous)
	return Result{Tracking: evidencedto.TrackingFromDomain(tracking)}, nil
}
