package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/usegavel/gavel/core/application/supporting/analyzetarget"
	"github.com/usegavel/gavel/core/infrastructure/casefile/sarif"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

const bazelBuildTimeout = 10 * time.Minute

type BazelTargetAnalyzer struct {
	runner CommandRunner
}

func NewBazelTargetAnalyzer(cr CommandRunner) *BazelTargetAnalyzer {
	return &BazelTargetAnalyzer{runner: cr}
}

func (a *BazelTargetAnalyzer) Analyze(ctx context.Context, workspace, target string, languages []string) ([]analyzetarget.Finding, error) {
	aspects := catalog.LintAspectsForLanguages(languages)
	if err := a.runBazelBuildTarget(ctx, workspace, target, aspects); err != nil {
		return nil, err
	}

	binDir, err := bazelBinDir(ctx, a.runner, workspace)
	if err != nil {
		return nil, err
	}

	return collectFindings(ctx, binDir, aspects)
}

func (a *BazelTargetAnalyzer) runBazelBuildTarget(ctx context.Context, workspace, target string, aspects []catalog.Aspect) error {
	ctx, cancel := context.WithTimeout(ctx, bazelBuildTimeout)
	defer cancel()

	args := []string{
		"build", target,
		"--aspects=" + catalog.AspectPaths(aspects),
		"--output_groups=gavel_submissions",
		"--keep_going",
	}

	stdout, stderr, err := a.runner.Run(ctx, workspace, "bazel", args...)
	if err != nil {
		output := string(stdout) + string(stderr)
		return fmt.Errorf("bazel %s: %w\n%s", strings.Join(args, " "), err, output)
	}
	return nil
}

func collectFindings(ctx context.Context, binDir string, aspects []catalog.Aspect) ([]analyzetarget.Finding, error) {
	parser := sarif.NewParser()
	var findings []analyzetarget.Finding

	for _, asp := range aspects {
		files, err := CollectSARIFFiles(binDir, asp.SARIFSuffix)
		if err != nil {
			continue
		}
		for _, data := range files {
			parsed, err := parser.Parse(ctx, data)
			if err != nil {
				continue
			}
			for _, parsedFinding := range parsed {
				findings = append(findings, analyzetarget.Finding{
					Tool:        asp.Name,
					RuleID:      parsedFinding.RuleID,
					Severity:    parsedFinding.Severity.String(),
					FilePath:    parsedFinding.FilePath,
					Line:        parsedFinding.Line,
					Message:     parsedFinding.Message,
					Fingerprint: parsedFinding.FingerprintID.Value(),
				})
			}
		}
	}
	return findings, nil
}
