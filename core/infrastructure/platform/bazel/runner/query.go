package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type BazelTargetQuery struct {
	runner CommandRunner
}

func NewBazelTargetQuery(cr CommandRunner) BazelTargetQuery {
	return BazelTargetQuery{runner: cr}
}

func (q BazelTargetQuery) QueryTargetsOfKind(ctx context.Context, workspace, pattern string, kinds []string) ([]string, error) {
	cmd := q.runner
	if cmd == nil {
		cmd = &ExecRunner{}
	}
	return queryTargetsOfKind(ctx, cmd, workspace, pattern, kinds)
}

func QueryTargetsOfKind(ctx context.Context, workspace, pattern string, kinds []string) ([]string, error) {
	return queryTargetsOfKind(ctx, &ExecRunner{}, workspace, pattern, kinds)
}

func queryTargetsOfKind(ctx context.Context, cmd CommandRunner, workspace, pattern string, kinds []string) ([]string, error) {
	if len(kinds) == 0 || strings.TrimSpace(pattern) == "" {
		return nil, nil
	}

	expr := fmt.Sprintf(`kind("^(%s)$", %s)`, strings.Join(kinds, "|"), pattern)
	stdout, stderr, err := cmd.Run(ctx, workspace, "bazel", "query", expr, "--output=label", "--keep_going")
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 && len(stdout) > 0 {
			return parseQueryLabels(stdout), nil
		}
		return nil, fmt.Errorf("bazel query: %w\n%s", err, string(stderr))
	}
	return parseQueryLabels(stdout), nil
}

func parseQueryLabels(out []byte) []string {
	lines := strings.Split(string(out), "\n")
	labels := make([]string, 0, len(lines))
	for _, line := range lines {
		label := strings.TrimSpace(line)
		if label == "" {
			continue
		}
		labels = append(labels, label)
	}
	return labels
}
