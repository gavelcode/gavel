package collectevidence

import (
	"fmt"
	"strings"
)

type Command struct {
	workspace       string
	targetPattern   string
	languages       []string
	defaultBranch   string
	projectName     string
	quick           bool
	absolute        bool
	baselineArchIDs []string
	scopedTargets   []string
	excludePatterns []string
	toolSelection   map[string][]string
}

type CommandOption func(*Command)

func WithToolSelection(selection map[string][]string) CommandOption {
	return func(c *Command) {
		c.toolSelection = copyToolSelection(selection)
	}
}

func copyToolSelection(selection map[string][]string) map[string][]string {
	if selection == nil {
		return nil
	}
	copied := make(map[string][]string, len(selection))
	for language, tools := range selection {
		toolsCopy := make([]string, len(tools))
		copy(toolsCopy, tools)
		copied[language] = toolsCopy
	}
	return copied
}

func WithScopedTargets(targets []string) CommandOption {
	return func(c *Command) {
		c.scopedTargets = append([]string(nil), targets...)
	}
}

func WithExcludePatterns(patterns []string) CommandOption {
	return func(c *Command) {
		c.excludePatterns = append([]string(nil), patterns...)
	}
}

func NewCommand(workspace, targetPattern, projectName, defaultBranch string, languages []string, quick, absolute bool, baselineArchIDs []string, opts ...CommandOption) (Command, error) {
	if strings.TrimSpace(workspace) == "" {
		return Command{}, fmt.Errorf("%w: workspace must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(targetPattern) == "" {
		return Command{}, fmt.Errorf("%w: targetPattern must not be empty", ErrInvalidCommand)
	}
	cmd := Command{
		workspace:       workspace,
		targetPattern:   targetPattern,
		languages:       languages,
		defaultBranch:   defaultBranch,
		projectName:     projectName,
		quick:           quick,
		absolute:        absolute,
		baselineArchIDs: baselineArchIDs,
	}
	for _, o := range opts {
		o(&cmd)
	}
	return cmd, nil
}

func (c Command) Workspace() string         { return c.workspace }
func (c Command) TargetPattern() string     { return c.targetPattern }
func (c Command) Languages() []string       { return c.languages }
func (c Command) DefaultBranch() string     { return c.defaultBranch }
func (c Command) ProjectName() string       { return c.projectName }
func (c Command) Quick() bool               { return c.quick }
func (c Command) Absolute() bool            { return c.absolute }
func (c Command) BaselineArchIDs() []string { return c.baselineArchIDs }
func (c Command) ScopedTargets() []string   { return c.scopedTargets }
func (c Command) ExcludePatterns() []string { return c.excludePatterns }

func (c Command) ToolSelection() map[string][]string { return copyToolSelection(c.toolSelection) }
