package collectevidence

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/classifyarch"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/application/casefile/ingestncc"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

const percentageMultiplier = 100

type Handler struct {
	findings      FindingsCollector
	coverage      CoverageCollector
	architecture  ArchitectureCollector
	ingestFind    *ingestfind.Handler
	ingestCov     *ingestcov.Handler
	classifyArch  *classifyarch.Handler
	ingestNCC     *ingestncc.Handler
	changedLines  ChangedLinesSource
	perLine       ingestncc.PerLineParser
	toolExecution ToolExecutionParser
}

type HandlerOption func(*Handler)

func WithChangedLinesSource(cls ChangedLinesSource) HandlerOption {
	return func(h *Handler) { h.changedLines = cls }
}

func WithPerLineParser(p ingestncc.PerLineParser) HandlerOption {
	return func(h *Handler) { h.perLine = p }
}

func WithToolExecutionParser(p ToolExecutionParser) HandlerOption {
	return func(h *Handler) { h.toolExecution = p }
}

func NewHandler(
	findings FindingsCollector,
	coverage CoverageCollector,
	architecture ArchitectureCollector,
	ingestFind *ingestfind.Handler,
	ingestCov *ingestcov.Handler,
	classifyArch *classifyarch.Handler,
	ingestNCC *ingestncc.Handler,
	opts ...HandlerOption,
) *Handler {
	handler := &Handler{
		findings:     findings,
		coverage:     coverage,
		architecture: architecture,
		ingestFind:   ingestFind,
		ingestCov:    ingestCov,
		classifyArch: classifyArch,
		ingestNCC:    ingestNCC,
	}
	for _, o := range opts {
		o(handler)
	}
	return handler
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	targets := cmd.ScopedTargets()
	if len(targets) == 0 {
		targets = []string{cmd.TargetPattern()}
		for _, exclude := range cmd.ExcludePatterns() {
			targets = append(targets, "-"+exclude)
		}
	}

	findingsEvidences, rawSARIF, buildWarning, err := h.collectFindings(ctx, cmd, targets)
	if err != nil {
		return Result{}, fmt.Errorf("findings: %w", err)
	}

	allFindings := evidencedto.ExtractFindings(findingsEvidences)
	findingsCount := len(allFindings)

	var sarifDocs [][]byte
	for _, rf := range rawSARIF {
		sarifDocs = append(sarifDocs, rf.Data)
	}

	var evidences []evidencedto.Evidence
	evidences = append(evidences, findingsEvidences...)
	evidences = h.appendToolExecutionEvidence(evidences, rawSARIF)

	var covPercent float64
	var rawLCOV []byte
	var archEvidence *evidencedto.Evidence
	var archCount int
	var archDelta classifyarch.Result

	if !cmd.Quick() {
		covPercent, rawLCOV, err = h.collectCoverage(ctx, cmd, targets, &evidences)
		if err != nil {
			return Result{}, err
		}

		archEvidence, archCount, archDelta, err = h.collectArchitecture(ctx, cmd, targets, &evidences)
		if err != nil {
			return Result{}, err
		}
	}

	coverageByFile, err := h.coverageByFile(rawLCOV)
	if err != nil {
		return Result{}, err
	}

	archIDs := evidencedto.ExtractArchIDs(evidencedto.ExtractViolations(archEvidence))

	var nccPercent float64
	if !cmd.Quick() && rawLCOV != nil && h.ingestNCC != nil && h.changedLines != nil {
		cl, clErr := h.changedLines.ChangedLines(ctx, cmd.Workspace(), cmd.DefaultBranch())
		if clErr == nil && len(cl) > 0 {
			nccCmd, cmdErr := ingestncc.NewCommand(rawLCOV, cl)
			if cmdErr == nil {
				nccRes, nccErr := h.ingestNCC.Execute(ctx, nccCmd)
				if nccErr == nil {
					nccPercent = nccRes.Percent
					evidences = append(evidences, nccRes.Evidence)
				}
			}
		}
	}

	return Result{
		Evidences:       evidences,
		FindingsCount:   findingsCount,
		ViolationsCount: archCount,
		CovPercent:      covPercent,
		NCCPercent:      nccPercent,
		CoverageByFile:  coverageByFile,
		RawSARIF:        rawSARIF,
		RawLCOV:         rawLCOV,
		SARIFDocs:       sarifDocs,
		Findings:        allFindings,
		Fingerprints:    evidencedto.ExtractFingerprints(allFindings),
		Violations:      evidencedto.ExtractViolations(archEvidence),
		ArchIDs:         archIDs,
		ArchDelta:       archDelta,
		BuildWarning:    buildWarning,
	}, nil
}

// appendToolExecutionEvidence reads the analyzer invocations from the collected
// SARIF and, if any analyzer did not complete, adds a tool_execution evidence
// carrying the failures. CaseFile.Judge turns that into a failing verdict.
func (h *Handler) appendToolExecutionEvidence(evidences []evidencedto.Evidence, rawSARIF []RawFile) []evidencedto.Evidence {
	if h.toolExecution == nil {
		return evidences
	}
	var failures []evidencedto.ToolFailure
	for _, raw := range rawSARIF {
		parsed, err := h.toolExecution.ParseToolExecutions(raw.Data)
		if err != nil {
			continue
		}
		failures = append(failures, parsed...)
	}
	if len(failures) == 0 {
		return evidences
	}
	return append(evidences, evidencedto.Evidence{
		Subtype:       evidence.SubtypeToolExecution.String(),
		Source:        "sarif",
		CollectedAt:   time.Now(),
		ToolExecution: &evidencedto.ToolExecution{Failures: failures},
	})
}

func (h *Handler) collectFindings(ctx context.Context, cmd Command, targets []string) ([]evidencedto.Evidence, []RawFile, string, error) {
	if h.findings == nil {
		return nil, nil, "", nil
	}
	return h.findings.CollectFindings(ctx, cmd.Workspace(), targets, cmd.ToolSelection())
}

func (h *Handler) collectCoverage(ctx context.Context, cmd Command, targets []string, evidences *[]evidencedto.Evidence) (float64, []byte, error) {
	if h.coverage == nil {
		return 0, nil, nil
	}
	data, err := h.coverage.CollectCoverage(ctx, cmd.Workspace(), targets, cmd.Languages())
	if err != nil {
		return 0, nil, fmt.Errorf("coverage: %w", err)
	}
	if data == nil {
		return 0, nil, nil
	}

	covCmd, err := ingestcov.NewCommand(data, "lcov", "bazel")
	if err != nil {
		return 0, nil, err
	}
	covRes, err := h.ingestCov.Execute(ctx, covCmd)
	if err != nil {
		return 0, nil, err
	}

	var percent float64
	if covRes.Evidence.Coverage != nil {
		total := covRes.Evidence.Coverage.TotalLines
		covered := covRes.Evidence.Coverage.CoveredLines
		if total > 0 {
			percent = float64(covered) / float64(total) * percentageMultiplier
		}
	}
	*evidences = append(*evidences, covRes.Evidence)
	return percent, data, nil
}

func (h *Handler) coverageByFile(rawLCOV []byte) ([]evidencedto.FileCoverage, error) {
	if h.perLine == nil || rawLCOV == nil {
		return nil, nil
	}
	perLine, err := h.perLine.ParsePerLine(rawLCOV)
	if err != nil {
		return nil, fmt.Errorf("coverage by file: %w", err)
	}
	return evidencedto.FileCoverageFromPerLine(perLine), nil
}

func (h *Handler) collectArchitecture(ctx context.Context, cmd Command, targets []string, evidences *[]evidencedto.Evidence) (*evidencedto.Evidence, int, classifyarch.Result, error) {
	if h.architecture == nil {
		return nil, 0, classifyarch.Result{}, nil
	}
	archEv, archDocs, err := h.architecture.CollectViolations(ctx, cmd.Workspace(), targets, cmd.ToolSelection())
	if err != nil {
		return nil, 0, classifyarch.Result{}, fmt.Errorf("architecture: %w", err)
	}

	_ = archDocs

	if archEv == nil {
		return nil, 0, classifyarch.Result{}, nil
	}

	violations := evidencedto.ExtractViolations(archEv)
	archIDs := evidencedto.ExtractArchIDs(violations)
	archCount := len(violations)

	if !cmd.Absolute() {
		archCmd := classifyarch.NewCommand(archIDs, cmd.BaselineArchIDs())
		classified, _ := h.classifyArch.Execute(ctx, archCmd)
		archEv = evidencedto.FilterNewViolations(archEv, classified.NewIDs)
		archCount = len(classified.NewIDs)
		*evidences = append(*evidences, *archEv)
		return archEv, archCount, classified, nil
	}

	*evidences = append(*evidences, *archEv)
	return archEv, archCount, classifyarch.Result{}, nil
}
