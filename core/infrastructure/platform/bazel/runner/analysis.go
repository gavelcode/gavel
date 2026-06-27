package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

const defaultTestSizeFilters = "small,medium"

type AnalysisConfig struct {
	Workspace       string
	Targets         []string
	Aspects         []catalog.Aspect
	IncludeCoverage bool
	TestSizeFilters string
	TestTagFilters  string
}

type AnalysisResult struct {
	SARIFFiles      map[string][][]byte
	CoverageData    []byte
	CoverageWarning error
	BuildWarning    error
}

func RunAnalysis(ctx context.Context, config AnalysisConfig) (*AnalysisResult, error) {
	return runAnalysis(ctx, &ExecRunner{}, config)
}

func runAnalysis(ctx context.Context, cmd CommandRunner, config AnalysisConfig) (*AnalysisResult, error) {
	args := buildBazelArgs(config)

	stdout, stderr, runErr := cmd.Run(ctx, config.Workspace, "bazel", args...)
	output := string(stdout) + string(stderr)

	binDir, err := bazelBinDir(ctx, cmd, config.Workspace)
	if err != nil {
		if runErr != nil {
			return nil, fmt.Errorf("analysis: bazel %s: %w\n%s", strings.Join(args, " "), runErr, output)
		}
		return nil, err
	}

	scopedDir := scopeBinDir(binDir, config.Targets)
	sarifFiles, err := collectAllSARIF(scopedDir, config.Aspects)
	if err != nil {
		return nil, fmt.Errorf("collect sarif: %w", err)
	}

	result := &AnalysisResult{SARIFFiles: sarifFiles}

	if config.IncludeCoverage {
		result.CoverageData, result.CoverageWarning = collectCoverageDataWith(ctx, cmd, config.Workspace, runErr)
	}

	if runErr != nil && len(sarifFiles) == 0 && result.CoverageData == nil {
		return nil, fmt.Errorf("analysis: bazel %s: %w\n%s", strings.Join(args, " "), runErr, output)
	}

	if runErr != nil {
		result.BuildWarning = fmt.Errorf("bazel build had failures (partial results collected): %w", runErr)
	}

	return result, nil
}

func buildBazelArgs(config AnalysisConfig) []string {
	var args []string
	if config.IncludeCoverage {
		args = append(args, "coverage")
	} else {
		args = append(args, "build")
	}
	args = append(args,
		"--aspects="+catalog.AspectPaths(config.Aspects),
		"--output_groups=gavel_submissions",
		"--keep_going",
	)

	if config.IncludeCoverage {
		sizeFilters := config.TestSizeFilters
		if sizeFilters == "" {
			sizeFilters = defaultTestSizeFilters
		}
		args = append(args,
			"--test_size_filters="+sizeFilters,
			"--combined_report=lcov",
		)
	}
	if config.TestTagFilters != "" {
		args = append(args, "--test_tag_filters="+config.TestTagFilters)
	}

	args = append(args, "--")
	args = append(args, config.Targets...)
	return args
}

func scopeBinDir(binDir string, targets []string) string {
	if len(targets) == 0 {
		return binDir
	}
	pkg := extractPackagePath(targets[0])
	if pkg == "" {
		return binDir
	}
	scoped := filepath.Join(binDir, pkg)
	if info, err := os.Stat(scoped); err == nil && info.IsDir() {
		return scoped
	}
	return binDir
}

func extractPackagePath(target string) string {
	target = strings.TrimPrefix(target, "//")
	if idx := strings.LastIndex(target, ":"); idx >= 0 {
		target = target[:idx]
	}
	target = strings.TrimSuffix(target, "...")
	target = strings.TrimSuffix(target, "/")
	return target
}

func collectAllSARIF(binDir string, aspects []catalog.Aspect) (map[string][][]byte, error) {
	result := make(map[string][][]byte)
	for _, asp := range aspects {
		files, err := CollectSARIFFiles(binDir, asp.SARIFSuffix)
		if err != nil {
			return nil, err
		}
		if len(files) > 0 {
			result[asp.Name] = files
		}
	}
	return result, nil
}

func SARIFReportsFromResult(result *AnalysisResult, aspects []catalog.Aspect) []SARIFReport {
	var reports []SARIFReport
	for _, asp := range aspects {
		files, ok := result.SARIFFiles[asp.Name]
		if !ok {
			continue
		}
		for _, data := range files {
			reports = append(reports, SARIFReport{Data: data, Source: ExtractToolName(data, asp.Name)})
		}
	}
	return reports
}

func collectCoverageDataWith(ctx context.Context, cmd CommandRunner, workspace string, runErr error) ([]byte, error) {
	reportPath, err := findCombinedReportWith(ctx, cmd, workspace)
	if err != nil {
		return nil, fmt.Errorf("find combined report: %w", err)
	}

	data, err := os.ReadFile(reportPath)
	if err == nil && len(data) > 0 {
		if runErr != nil {
			return data, fmt.Errorf("bazel coverage had partial failures (report still collected): %w", runErr)
		}
		return data, nil
	}

	fallbackData, count, fallbackErr := collectIndividualCoverageFiles(ctx, workspace)
	if fallbackErr != nil {
		if runErr != nil {
			return nil, fmt.Errorf("bazel coverage failed and fallback failed: %w", runErr)
		}
		return nil, fmt.Errorf("coverage fallback: %w", fallbackErr)
	}
	if count > 0 {
		return fallbackData, fmt.Errorf("combined report unavailable; collected %d individual coverage files", count)
	}

	if runErr != nil {
		return nil, fmt.Errorf("bazel coverage failed and produced no report: %w", runErr)
	}
	return nil, nil
}
