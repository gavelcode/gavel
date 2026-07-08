package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var lookPathFunc = exec.LookPath

func RunJSCoverage(ctx context.Context, workspace, targetPattern string, log *slog.Logger) ([]byte, error) {
	return runJSCoverage(ctx, &ExecRunner{}, workspace, targetPattern, log)
}

func runJSCoverage(ctx context.Context, cmd CommandRunner, workspace, targetPattern string, log *slog.Logger) ([]byte, error) {
	projectDir := resolveProjectDir(workspace, targetPattern)
	if projectDir == "" {
		return nil, nil
	}

	if !hasPackageJSON(projectDir) {
		return nil, nil
	}

	npx, err := lookPathFunc("npx")
	if err != nil {
		return nil, fmt.Errorf("npx not found: %w", err)
	}

	coverageDir, err := os.MkdirTemp("", "gavel-jscoverage-*")
	if err != nil {
		return nil, fmt.Errorf("create coverage dir: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(coverageDir); err != nil {
			log.Warn("failed to remove coverage dir", "path", coverageDir, "error", err)
		}
	}()

	args := []string{
		"vitest", "run",
		"--coverage.enabled",
		"--coverage.provider=v8",
		"--coverage.reporter=lcov",
		"--coverage.reportsDirectory=" + coverageDir,
	}

	_, stderr, err := cmd.Run(ctx, projectDir, npx, args...)
	if err != nil {
		// vitest exits 1 on test failures; coverage is still valid, so only other errors are fatal.
		exitErr, ok := err.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 1 {
			return nil, fmt.Errorf("vitest coverage: %w\n%s", err, string(stderr))
		}
	}

	lcovPath := filepath.Join(coverageDir, "lcov.info")
	data, err := os.ReadFile(lcovPath)
	if err != nil {
		return nil, nil
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data, nil
}

func resolveProjectDir(workspace, targetPattern string) string {
	pattern := strings.TrimPrefix(targetPattern, "//")
	pattern = strings.TrimSuffix(pattern, "...")
	pattern = strings.TrimSuffix(pattern, "/")
	pattern = strings.TrimSuffix(pattern, ":")
	if pattern == "" {
		return workspace
	}
	return filepath.Join(workspace, pattern)
}

func hasPackageJSON(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "package.json"))
	return err == nil
}
