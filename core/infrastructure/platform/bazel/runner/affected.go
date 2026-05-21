package runner

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const bazelQueryTimeout = 2 * time.Minute

type BazelTargetResolver struct {
	runner CommandRunner
}

func NewBazelTargetResolver(cr CommandRunner) *BazelTargetResolver {
	return &BazelTargetResolver{runner: cr}
}

func (r *BazelTargetResolver) FindAffectedTargets(ctx context.Context, workspace string, changedFiles []string, scope string) ([]string, error) {
	if len(changedFiles) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, bazelQueryTimeout)
	defer cancel()

	fileSet := buildAffectedFileSet(workspace, changedFiles)
	query := buildRdepsQuery(fileSet, scope)

	stdout, stderr, err := r.runner.Run(ctx, workspace, "bazel", "query", query, "--output=label", "--keep_going")
	if err != nil {
		labels := parseAffectedLabels(string(stdout))
		if len(labels) > 0 {
			return labels, nil
		}
		return nil, fmt.Errorf("bazel query: %w\n%s", err, string(stderr))
	}

	return parseAffectedLabels(string(stdout)), nil
}

func buildAffectedFileSet(workspace string, changedFiles []string) string {
	labels := make([]string, 0, len(changedFiles))
	for _, f := range changedFiles {
		rel := strings.TrimPrefix(f, workspace+"/")
		labels = append(labels, rel)
	}
	return strings.Join(labels, " ")
}

func buildRdepsQuery(fileSet, scope string) string {
	universe := "//..."
	if scope != "" {
		universe = scope
	}
	return fmt.Sprintf("rdeps(%s, set(%s))", universe, fileSet)
}

func (r *BazelTargetResolver) FindOwnerTarget(ctx context.Context, workspace, file string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, bazelQueryTimeout)
	defer cancel()

	stdout, stderr, err := r.runner.Run(ctx, workspace, "bazel", "query", file, "--output=package")
	if err != nil {
		return "", fmt.Errorf("bazel query owner: %w\n%s", err, string(stderr))
	}

	pkg := strings.TrimSpace(string(stdout))
	if pkg == "" {
		return "", fmt.Errorf("no package found for file %s", file)
	}
	return buildOwnerTarget(pkg), nil
}

func buildOwnerTarget(pkg string) string {
	return "//" + pkg + ":all"
}

func parseAffectedLabels(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var labels []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.HasPrefix(line, "//") {
			labels = append(labels, line)
		}
	}
	return labels
}
