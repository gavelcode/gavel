package ingestncc

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

type Handler struct {
	parser PerLineParser
	now    func() time.Time
}

func NewHandler(parser PerLineParser) *Handler {
	if parser == nil {
		panic("ingestncc: parser must not be nil")
	}
	return &Handler{parser: parser, now: func() time.Time { return time.Now().UTC() }}
}

func (h *Handler) Execute(_ context.Context, cmd Command) (Result, error) {
	perLine, err := h.parser.ParsePerLine(cmd.RawLCOV())
	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrParseFailed, err)
	}

	patch, err := coverage.NewPatchContentFromChanges(cmd.ChangedLines(), perLine)
	if err != nil {
		return Result{}, fmt.Errorf("compute patch coverage: %w", err)
	}

	evidence := evidencedto.Evidence{
		Subtype:     "new_code_coverage",
		Source:      "gavel",
		CollectedAt: h.now(),
		NewCodeCoverage: &evidencedto.NewCodeCoverage{
			CoveredLines:   patch.CoveredLines(),
			CoverableLines: patch.CoverableLines(),
		},
	}
	return Result{Evidence: evidence, Percent: patch.Percent()}, nil
}
