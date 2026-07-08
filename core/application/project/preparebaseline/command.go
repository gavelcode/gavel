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
	tenantID string
	projects []ProjectInput
}

func NewCommand(tenantID string, projects []ProjectInput) (Command, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
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
	return Command{tenantID: tenantID, projects: cp}, nil
}

func (c Command) TenantID() string         { return c.tenantID }
func (c Command) Projects() []ProjectInput { return c.projects }
