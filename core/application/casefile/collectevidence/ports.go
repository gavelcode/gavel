package collectevidence

import (
	"context"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

type FindingsCollector interface {
	CollectFindings(ctx context.Context, workspace string, targets []string, selection map[string][]string, bazelConfigs []string) ([]evidencedto.Evidence, []RawFile, string, []string, error)
}

type CoverageCollector interface {
	CollectCoverage(ctx context.Context, workspace string, targets []string, languages []string, bazelConfigs []string) ([]byte, error)
}

type ArchitectureCollector interface {
	CollectViolations(ctx context.Context, workspace string, targets []string, selection map[string][]string, bazelConfigs []string) (*evidencedto.Evidence, [][]byte, error)
}

type ChangedLinesSource interface {
	ChangedLines(ctx context.Context, workspace, baseRef string) (map[string][]int, error)
}

type ToolExecutionParser interface {
	ParseToolExecutions(data []byte) ([]evidencedto.ToolFailure, error)
}
