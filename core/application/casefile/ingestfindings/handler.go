package ingestfindings

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

type Handler struct {
	parsers map[string]Parser
	now     func() time.Time
}

func NewHandler(parsers map[string]Parser) *Handler {
	if len(parsers) == 0 {
		panic("findings: at least one parser must be registered")
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

	findings, err := toDomainFindings(cmd.Source(), parsed)
	if err != nil {
		return Result{}, fmt.Errorf("build findings: %w", err)
	}

	content, err := finding.NewContent(cmd.Subtype(), findings)
	if err != nil {
		return Result{}, fmt.Errorf("build findings content: %w", err)
	}

	ev, err := evidence.NewEvidence(cmd.Subtype(), cmd.Source(), content, h.now())
	if err != nil {
		return Result{}, fmt.Errorf("build evidence: %w", err)
	}

	return Result{Evidence: evidencedto.EvidenceFromDomain(ev)}, nil
}

func toDomainFindings(tool string, parsed []Parsed) ([]finding.Finding, error) {
	findings := make([]finding.Finding, 0, len(parsed))
	for i, p := range parsed {
		f, err := finding.NewFinding(tool, p.RuleID, p.Severity, p.FilePath, p.Line, p.Message, p.FingerprintID)
		if err != nil {
			return nil, fmt.Errorf("finding %d: %w", i, err)
		}
		findings = append(findings, f)
	}
	return findings, nil
}
