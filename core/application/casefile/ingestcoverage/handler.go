package ingestcoverage

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

type Handler struct {
	parsers map[string]Parser
	now     func() time.Time
}

func NewHandler(parsers map[string]Parser) *Handler {
	if len(parsers) == 0 {
		panic("coverage: at least one parser must be registered")
	}
	copied := make(map[string]Parser, len(parsers))
	for k, v := range parsers {
		copied[k] = v
	}
	return &Handler{parsers: copied, now: func() time.Time { return time.Now().UTC() }}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	parser, ok := h.parsers[cmd.Format()]
	if !ok {
		return Result{}, fmt.Errorf("%w: %q", ErrUnknownFormat, cmd.Format())
	}

	parsed, err := parser.Parse(ctx, cmd.Data())
	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrParseFailed, err)
	}

	content, err := coverage.NewContent(parsed.TotalLines, parsed.CoveredLines, parsed.ByLanguage)
	if err != nil {
		return Result{}, fmt.Errorf("build coverage content: %w", err)
	}

	ev, err := evidence.NewEvidence(evidence.SubtypeCoverage, cmd.Source(), content, h.now())
	if err != nil {
		return Result{}, fmt.Errorf("build evidence: %w", err)
	}

	evDTO := evidencedto.EvidenceFromDomain(ev)
	if evDTO.Coverage != nil {
		evDTO.Coverage.ByFile = parsed.ByFile
	}

	return Result{Evidence: evDTO}, nil
}
