package collector

import (
	"context"
	"log/slog"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
)

type VitestCoverageCollector struct {
	log *slog.Logger
}

func NewVitestCoverageCollector(log *slog.Logger) *VitestCoverageCollector {
	return &VitestCoverageCollector{log: log}
}

func (c *VitestCoverageCollector) CollectCoverage(ctx context.Context, workspace string, targets []string, _ []string) ([]byte, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	return runner.RunJSCoverage(ctx, workspace, targets[0], c.log)
}
