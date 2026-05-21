package collector

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

type AnalysisRunner interface {
	RunAnalysis(ctx context.Context, config runner.AnalysisConfig) (*runner.AnalysisResult, error)
}

type FindingsParser interface {
	Execute(ctx context.Context, cmd ingestfind.Command) (ingestfind.Result, error)
}

type BazelFindingsCollector struct {
	runner AnalysisRunner
	parser FindingsParser
}

func NewBazelFindingsCollector(r AnalysisRunner, p FindingsParser) *BazelFindingsCollector {
	return &BazelFindingsCollector{runner: r, parser: p}
}

func (c *BazelFindingsCollector) CollectFindings(ctx context.Context, workspace string, targets []string, languages []string) ([]evidencedto.Evidence, []collectevidence.RawFile, string, error) {
	lintAspects := catalog.LintAspectsForLanguages(languages)
	exclusiveAspects := catalog.GavelExclusiveLintAspects(languages)

	aspects := append(lintAspects, exclusiveAspects...)
	if len(aspects) == 0 {
		return nil, nil, "", nil
	}

	config := runner.AnalysisConfig{
		Workspace: workspace,
		Targets:   targets,
		Aspects:   lintAspects,
	}

	result, err := c.runner.RunAnalysis(ctx, config)
	if err != nil {
		return nil, nil, "", fmt.Errorf("run analysis: %w", err)
	}

	reports := runner.SARIFReportsFromResult(result, lintAspects)

	var evidences []evidencedto.Evidence
	var rawFiles []collectevidence.RawFile
	for _, report := range reports {
		cmd, err := ingestfind.NewCommand(report.Data, "sarif", report.Source, "code_quality")
		if err != nil {
			return nil, nil, "", err
		}
		res, err := c.parser.Execute(ctx, cmd)
		if err != nil {
			return nil, nil, "", err
		}
		evidences = append(evidences, res.Evidence)
		rawFiles = append(rawFiles, collectevidence.RawFile{
			Format: "sarif",
			Source: report.Source + ".sarif",
			Data:   report.Data,
		})
	}

	var buildWarning string
	if result.BuildWarning != nil {
		buildWarning = result.BuildWarning.Error()
	}
	return evidences, rawFiles, buildWarning, nil
}
