package resolve

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/pleading/model"
	pleadingservice "github.com/usegavel/gavel/core/domain/pleading/service"
)

type Handler struct {
	pleadings pleadingservice.PleadingRepository
}

func NewHandler(pleadings pleadingservice.PleadingRepository) *Handler {
	if pleadings == nil {
		panic("resolve: pleadings repository must not be nil")
	}
	return &Handler{pleadings: pleadings}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	tenantID, err := tenant.ParseTenantID(cmd.TenantID())
	if err != nil {
		return Result{}, fmt.Errorf("resolve pleading: %w", err)
	}

	id, err := model.ParsePleadingID(cmd.PleadingID())
	if err != nil {
		return Result{}, fmt.Errorf("resolve pleading: %w", err)
	}

	pleading, err := h.pleadings.FindByID(ctx, tenantID, id)
	if err != nil {
		return Result{}, fmt.Errorf("load pleading: %w", err)
	}

	target, err := model.NewStatus(cmd.Outcome())
	if err != nil {
		return Result{}, fmt.Errorf("resolve pleading: %w", err)
	}

	if pleading.Status().Equal(target) {
		return Result{Changed: false, Status: pleading.Status().String()}, nil
	}

	if err := applyTransition(&pleading, target, time.Now().UTC()); err != nil {
		return Result{}, fmt.Errorf("resolve pleading: %w", err)
	}

	events := pleading.Events()
	pleading.ClearEvents()

	if err := h.pleadings.Save(ctx, pleading); err != nil {
		return Result{}, fmt.Errorf("save pleading: %w", err)
	}

	return Result{
		Changed: true,
		Status:  pleading.Status().String(),
		Events:  event.EventsFromDomain(events),
	}, nil
}

func applyTransition(p *model.Pleading, target model.Status, occurredAt time.Time) error {
	switch target {
	case model.StatusMerged:
		return p.MarkMerged(occurredAt)
	case model.StatusClosed:
		return p.MarkClosed(occurredAt)
	default:
		return fmt.Errorf("%w: unsupported target status %q", model.ErrInvalidTransition, target)
	}
}
