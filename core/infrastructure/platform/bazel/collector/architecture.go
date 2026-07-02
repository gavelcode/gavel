package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	coresarif "github.com/usegavel/gavel/core/infrastructure/casefile/sarif"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
)

type BazelArchitectureCollector struct {
	runner AnalysisRunner
}

func NewBazelArchitectureCollector(r AnalysisRunner) *BazelArchitectureCollector {
	return &BazelArchitectureCollector{runner: r}
}

func (c *BazelArchitectureCollector) CollectViolations(ctx context.Context, workspace string, targets []string, selection map[string][]string) (*evidencedto.Evidence, [][]byte, error) {
	selected, err := catalog.SelectedAspects(selection)
	if err != nil {
		return nil, nil, err
	}
	archAspects := make([]catalog.Aspect, 0, len(selected))
	for _, aspect := range selected {
		if catalog.IsArchtestAspect(aspect.Name) {
			archAspects = append(archAspects, aspect)
		}
	}
	if len(archAspects) == 0 {
		return nil, nil, nil
	}

	config := runner.AnalysisConfig{
		Workspace: workspace,
		Targets:   targets,
		Aspects:   archAspects,
	}

	result, err := c.runner.RunAnalysis(ctx, config)
	if err != nil {
		return nil, nil, fmt.Errorf("run archtest: %w", err)
	}

	var allViolations []evidencedto.Violation
	var sarifDocs [][]byte
	for _, asp := range archAspects {
		files, ok := result.SARIFFiles[asp.Name]
		if !ok {
			continue
		}
		for _, data := range files {
			violations, err := coresarif.ParseArchitectureViolations(data)
			if err != nil {
				return nil, nil, fmt.Errorf("parse archtest SARIF: %w", err)
			}
			allViolations = append(allViolations, violations...)
			sarifDocs = append(sarifDocs, data)
		}
	}

	if len(allViolations) == 0 {
		return nil, sarifDocs, nil
	}

	evidence := evidencedto.Evidence{
		Subtype:      "architecture",
		Source:       "archtest",
		CollectedAt:  time.Now().UTC(),
		Architecture: &evidencedto.Architecture{Violations: allViolations},
	}
	return &evidence, sarifDocs, nil
}
