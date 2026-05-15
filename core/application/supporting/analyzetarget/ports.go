package analyzetarget

import "context"

type TargetAnalyzer interface {
	Analyze(ctx context.Context, workspace, target string, languages []string) ([]Finding, error)
}

type TargetResolver interface {
	FindAffectedTargets(ctx context.Context, workspace string, changedFiles []string, scope string) ([]string, error)
}
