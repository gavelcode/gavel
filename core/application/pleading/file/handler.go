package file

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/pleading/model"
	pleadingservice "github.com/usegavel/gavel/core/domain/pleading/service"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type Handler struct {
	pleadings pleadingservice.PleadingRepository
}

func NewHandler(pleadings pleadingservice.PleadingRepository) *Handler {
	if pleadings == nil {
		panic("file: pleadings repository must not be nil")
	}
	return &Handler{pleadings: pleadings}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	tenantID, err := tenant.ParseTenantID(cmd.TenantID())
	if err != nil {
		return Result{}, fmt.Errorf("file pleading: %w", err)
	}

	projectID, err := projectmodel.ParseProjectID(cmd.ProjectID())
	if err != nil {
		return Result{}, fmt.Errorf("file pleading: %w", err)
	}

	pleading, err := model.FilePleading(
		tenantID, projectID, cmd.Number(), cmd.Title(), cmd.Petitioner(),
		cmd.SourceBranch(), cmd.TargetBranch(), cmd.CommitSHA(),
	)
	if err != nil {
		return Result{}, fmt.Errorf("file pleading: %w", err)
	}

	events := pleading.Events()
	pleading.ClearEvents()

	if err := h.pleadings.Save(ctx, pleading); err != nil {
		return Result{}, fmt.Errorf("save pleading: %w", err)
	}

	return Result{
		PleadingID: pleading.ID().String(),
		Status:     pleading.Status().String(),
		Events:     event.EventsFromDomain(events),
	}, nil
}
