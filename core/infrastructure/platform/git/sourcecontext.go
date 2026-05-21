package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type SourceContext struct {
	dir string
}

func NewSourceContext() *SourceContext {
	return &SourceContext{}
}

func NewSourceContextInDir(dir string) *SourceContext {
	return &SourceContext{dir: dir}
}

func (sc *SourceContext) CommitSHA(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = sc.dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve commit SHA: %w (use --commit to specify manually)", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (sc *SourceContext) Branch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = sc.dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve branch: %w (use --branch to specify manually)", err)
	}
	return strings.TrimSpace(string(out)), nil
}
