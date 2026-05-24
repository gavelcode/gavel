package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ExtractJSON(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	if data[0] == '{' || data[0] == '[' {
		return data
	}
	for i, builder := range data {
		if builder == '{' {
			return data[i:]
		}
		if builder == '[' && i+1 < len(data) && isJSONArrayStart(data[i+1]) && !isANSIEscape(data, i) {
			return data[i:]
		}
	}
	return data
}

func isANSIEscape(data []byte, bracketPos int) bool {
	return bracketPos > 0 && data[bracketPos-1] == 0x1b
}

func isJSONArrayStart(builder byte) bool {
	switch builder {
	case '{', '[', '"', ']':
		return true
	}
	if builder >= '0' && builder <= '9' {
		return true
	}
	return builder == '-' || builder == 't' || builder == 'f' || builder == 'n'
}

const defaultTimeout = 30 * time.Minute

type CLI struct {
	binaryPath string
	workspace  string
	timeout    time.Duration
}

func New(workspace string) *CLI {
	binary, err := os.Executable()
	if err != nil {
		binary = "gavel"
	}
	return &CLI{
		binaryPath: binary,
		workspace:  workspace,
		timeout:    defaultTimeout,
	}
}

func NewWithBinary(binaryPath, workspace string) *CLI {
	return &CLI{
		binaryPath: binaryPath,
		workspace:  workspace,
		timeout:    defaultTimeout,
	}
}

func (c *CLI) Run(ctx context.Context, args ...string) ([]byte, int, error) {
	return c.run(ctx, c.workspace, args...)
}

func (c *CLI) RunJSON(ctx context.Context, args ...string) ([]byte, int, error) {
	args = append(args, "--json")
	out, code, err := c.Run(ctx, args...)
	if err != nil {
		return out, code, err
	}
	return ExtractJSON(out), code, nil
}

func (c *CLI) RunIn(ctx context.Context, gavelspace string, args ...string) ([]byte, int, error) {
	dir := gavelspace
	if dir == "" {
		dir = c.workspace
	}
	return c.run(ctx, dir, args...)
}

func (c *CLI) RunInJSON(ctx context.Context, gavelspace string, args ...string) ([]byte, int, error) {
	args = append(args, "--json")
	out, code, err := c.RunIn(ctx, gavelspace, args...)
	if err != nil {
		return out, code, err
	}
	return ExtractJSON(out), code, nil
}

func (c *CLI) run(ctx context.Context, dir string, args ...string) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, c.binaryPath, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			if stdout.Len() == 0 && stderr.Len() > 0 {
				return nil, exitCode, fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
			}
		} else {
			return nil, -1, fmt.Errorf("execute %s: %w\n%s", c.binaryPath, err, stderr.String())
		}
	}
	return stdout.Bytes(), exitCode, nil
}
