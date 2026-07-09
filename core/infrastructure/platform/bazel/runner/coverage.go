package runner

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type CoverageResult struct {
	Data    []byte
	Warning error
}

func RunCoverage(ctx context.Context, workspace string, targets []string) (*CoverageResult, error) {
	return runCoverage(ctx, &ExecRunner{}, workspace, targets)
}

func runCoverage(ctx context.Context, cmd CommandRunner, workspace string, targets []string) (*CoverageResult, error) {
	args := []string{"coverage"}
	args = append(args, "--combined_report=lcov", "--keep_going")
	args = append(args, "--")
	args = append(args, targets...)

	_, _, runErr := cmd.Run(ctx, workspace, "bazel", args...)

	reportPath, err := findCombinedReportWith(ctx, cmd, workspace)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		if os.IsNotExist(err) {
			if runErr != nil {
				return &CoverageResult{Warning: fmt.Errorf("bazel coverage failed and produced no report: %w", runErr)}, nil
			}
			return &CoverageResult{}, nil
		}
		return nil, fmt.Errorf("read coverage report: %w", err)
	}

	result := &CoverageResult{Data: data}
	if runErr != nil {
		result.Warning = fmt.Errorf("bazel coverage had partial failures (report still collected): %w", runErr)
	}
	return result, nil
}

func findCombinedReportWith(ctx context.Context, cmd CommandRunner, workspace string) (string, error) {
	outputPath, err := bazelOutputPathWith(ctx, cmd, workspace)
	if err != nil {
		return "", err
	}
	return filepath.Join(outputPath, "_coverage", "_coverage_report.dat"), nil
}

func bazelOutputPathWith(ctx context.Context, cmd CommandRunner, workspace string) (string, error) {
	stdout, stderr, err := cmd.Run(ctx, workspace, "bazel", "info", "output_path")
	if err != nil {
		return "", fmt.Errorf("bazel info output_path: %w\n%s", err, string(stderr))
	}
	return strings.TrimSpace(string(stdout)), nil
}

func collectIndividualCoverageFiles(ctx context.Context, workspace string) ([]byte, int, error) {
	testlogsDir, err := resolveTestlogsDir(ctx, workspace)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve testlogs dir: %w", err)
	}
	if resolved, evalErr := filepath.EvalSymlinks(testlogsDir); evalErr == nil {
		testlogsDir = resolved
	}

	var merged bytes.Buffer
	var count int
	walkErr := filepath.WalkDir(testlogsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || d.Name() != "coverage.dat" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if len(data) == 0 {
			return nil
		}
		merged.Write(data)
		if data[len(data)-1] != '\n' {
			merged.WriteByte('\n')
		}
		count++
		return nil
	})
	if walkErr != nil {
		return nil, 0, fmt.Errorf("walk testlogs: %w", walkErr)
	}
	if count == 0 {
		return nil, 0, nil
	}
	return merged.Bytes(), count, nil
}

func resolveTestlogsDir(ctx context.Context, workspace string) (string, error) {
	return resolveTestlogsDirWith(ctx, &ExecRunner{}, workspace)
}

func resolveTestlogsDirWith(ctx context.Context, cmd CommandRunner, workspace string) (string, error) {
	symlink := filepath.Join(workspace, "bazel-testlogs")
	if info, err := os.Stat(symlink); err == nil && info.IsDir() {
		return symlink, nil
	}
	outputPath, err := bazelOutputPathWith(ctx, cmd, workspace)
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(outputPath), "testlogs"), nil
}
