package preparebaseline

import (
	"fmt"
	"strings"
)

type ProjectInput struct {
	Name          string
	DefaultBranch string
}

type Command struct {
	projects []ProjectInput
}

func NewCommand(projects []ProjectInput) (Command, error) {
	if len(projects) == 0 {
		return Command{}, fmt.Errorf("%w: at least one project is required", ErrInvalidCommand)
	}
	for _, p := range projects {
		if strings.TrimSpace(p.Name) == "" {
			return Command{}, fmt.Errorf("%w: project name must not be empty", ErrInvalidCommand)
		}
	}
	cp := make([]ProjectInput, len(projects))
	copy(cp, projects)
	return Command{projects: cp}, nil
}

func (c Command) Projects() []ProjectInput { return c.projects }
