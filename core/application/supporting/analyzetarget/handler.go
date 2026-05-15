package analyzetarget

import (
	"context"
	"fmt"
	"time"
)

type Handler struct {
	analyzer TargetAnalyzer
}

func NewHandler(analyzer TargetAnalyzer) *Handler {
	if analyzer == nil {
		panic("analyzetarget: analyzer must not be nil")
	}
	return &Handler{analyzer: analyzer}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	start := time.Now()
	findings, err := h.analyzer.Analyze(ctx, cmd.Workspace(), cmd.Target(), cmd.Languages())
	if err != nil {
		return Result{}, fmt.Errorf("analyze %s: %w", cmd.Target(), err)
	}
	return Result{
		Target:   cmd.Target(),
		Findings: findings,
		Duration: time.Since(start),
	}, nil
}
