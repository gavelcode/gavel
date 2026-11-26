package collector

import (
	"context"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
)

type BazelRunner struct{}

func NewBazelRunner() *BazelRunner { return &BazelRunner{} }

func (r *BazelRunner) RunAnalysis(ctx context.Context, config runner.AnalysisConfig) (*runner.AnalysisResult, error) {
	return runner.RunAnalysis(ctx, config)
}

func (r *BazelRunner) RunBridge(ctx context.Context, config runner.BridgeConfig) (*runner.BridgeResult, error) {
	return runner.RunBridge(ctx, config)
}
