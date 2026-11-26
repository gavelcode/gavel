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

// BridgeSpec is a project-declared linter aspect (see project model Bridge).
type BridgeSpec struct {
	Name        string
	Aspect      string
	SARIFSuffix string
	BuildFlags  []string
}

// BridgeResult is what running a project's bridges produced. Findings and tool
// failures are already marked as bridge provenance; RawFiles are the validated
// SARIF documents (tagged for emission) and carry no failures of their own.
type BridgeResult struct {
	Evidences []evidencedto.Evidence
	RawFiles  []RawFile
}

// BridgeCollector builds and judges the project's bridges. A bridge that
// breaks the contract yields a failed tool execution, never an error.
type BridgeCollector interface {
	CollectBridges(ctx context.Context, workspace string, targets []string, bridges []BridgeSpec, bazelConfigs []string) (BridgeResult, error)
}
