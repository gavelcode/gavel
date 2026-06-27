package runner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

func RunAspect(ctx context.Context, workspace string, targets []string, asp catalog.Aspect) ([][]byte, error) {
	return runAspect(ctx, &ExecRunner{}, workspace, targets, asp)
}

func runAspect(ctx context.Context, cmd CommandRunner, workspace string, targets []string, asp catalog.Aspect) ([][]byte, error) {
	args := []string{"build"}
	args = append(args,
		"--aspects="+asp.Path,
		"--output_groups=gavel_submissions",
		"--keep_going",
	)
	args = append(args, "--")
	args = append(args, targets...)

	stdout, stderr, runErr := cmd.Run(ctx, workspace, "bazel", args...)
	output := string(stdout) + string(stderr)

	binDir, err := bazelBinDir(ctx, cmd, workspace)
	if err != nil {
		if runErr != nil {
			return nil, fmt.Errorf("aspect %s: bazel %s: %w\n%s", asp.Name, strings.Join(args, " "), runErr, output)
		}
		return nil, err
	}

	results, err := CollectSARIFFiles(binDir, asp.SARIFSuffix)
	if err != nil {
		return nil, err
	}
	if runErr != nil && len(results) == 0 {
		return nil, fmt.Errorf("aspect %s: bazel %s: %w\n%s", asp.Name, strings.Join(args, " "), runErr, output)
	}
	return results, nil
}

func CollectSARIFFiles(dir, suffix string) ([][]byte, error) {
	if suffix == "" {
		suffix = ".sarif"
	}
	var results [][]byte
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), suffix) {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read sarif %s: %w", path, readErr)
		}
		results = append(results, data)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan bazel-bin for sarif: %w", err)
	}
	return results, nil
}
