package analyzetarget

import (
	"fmt"
	"strings"
)

type Command struct {
	workspace string
	target    string
	languages []string
}

func NewCommand(workspace, target string, languages []string) (Command, error) {
	if strings.TrimSpace(workspace) == "" {
		return Command{}, fmt.Errorf("%w: workspace must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(target) == "" {
		return Command{}, fmt.Errorf("%w: target must not be empty", ErrInvalidCommand)
	}
	return Command{workspace: workspace, target: target, languages: languages}, nil
}

func (c Command) Workspace() string  { return c.workspace }
func (c Command) Target() string     { return c.target }
func (c Command) Languages() []string { return c.languages }
