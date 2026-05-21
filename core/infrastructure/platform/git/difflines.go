package git

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var hunkHeaderRegexp = regexp.MustCompile(`^@@ .* \+(\d+)(?:,(\d+))? @@`)

func (sc *SourceContext) ChangedLines(ctx context.Context, workspace, baseRef string) (map[string][]int, error) {
	mergeBaseOut, err := exec.CommandContext(ctx, "git", "-C", workspace, "merge-base", baseRef, "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("git merge-base: %w", err)
	}
	mergeBase := strings.TrimSpace(string(mergeBaseOut))

	diffOut, err := exec.CommandContext(ctx, "git", "-C", workspace, "diff", "--unified=0", mergeBase+"...HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}

	return parseDiffOutput(string(diffOut))
}

func parseDiffOutput(output string) (map[string][]int, error) {
	result := make(map[string][]int)
	var currentFile string

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			continue
		}

		matches := hunkHeaderRegexp.FindStringSubmatch(line)
		if matches == nil || currentFile == "" {
			continue
		}

		start, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse hunk start in %q: %w", line, err)
		}
		count := 1
		if matches[2] != "" {
			count, err = strconv.Atoi(matches[2])
			if err != nil {
				return nil, fmt.Errorf("parse hunk count in %q: %w", line, err)
			}
		}

		for i := 0; i < count; i++ {
			result[currentFile] = append(result[currentFile], start+i)
		}
	}
	return result, nil
}
