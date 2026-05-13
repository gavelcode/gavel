package classifyarch

import (
	"context"

	"github.com/usegavel/gavel/core/domain/casefile/model/tracking"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Execute(_ context.Context, cmd Command) (Result, error) {
	classified := tracking.ClassifyIdentifiers(cmd.CurrentIDs(), cmd.PreviousIDs())
	return Result{
		NewCount:      classified.NewCount(),
		FixedCount:    classified.ResolvedCount(),
		ExistingCount: classified.ExistingCount(),
		NewIDs:        classified.NewIdentifiers(),
	}, nil
}
