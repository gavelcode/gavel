package collector

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
)

type BazelCoverageCollector struct {
	runner AnalysisRunner
}

func NewBazelCoverageCollector(r AnalysisRunner) *BazelCoverageCollector {
	return &BazelCoverageCollector{runner: r}
}

func (c *BazelCoverageCollector) CollectCoverage(ctx context.Context, workspace string, targets []string, _ []string) ([]byte, error) {
	config := runner.AnalysisConfig{
		Workspace:       workspace,
		Targets:         targets,
		IncludeCoverage: true,
	}

	result, err := c.runner.RunAnalysis(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("run coverage: %w", err)
	}
	return result.CoverageData, nil
}
