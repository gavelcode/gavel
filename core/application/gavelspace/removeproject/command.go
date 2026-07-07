package removeproject

import (
	"fmt"
	"strings"
)

type Command struct {
	tenantID       string
	gavelspaceName string
	projectID      string
}

func NewCommand(tenantID, gavelspaceName, projectID string) (Command, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(gavelspaceName) == "" {
		return Command{}, fmt.Errorf("%w: gavelspaceName must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(projectID) == "" {
		return Command{}, fmt.Errorf("%w: projectID must not be empty", ErrInvalidCommand)
	}
	return Command{
		tenantID:       tenantID,
		gavelspaceName: gavelspaceName,
		projectID:      projectID,
	}, nil
}

func (c Command) TenantID() string     { return c.tenantID }
func (c Command) GavelspaceID() string { return c.gavelspaceName }
func (c Command) ProjectID() string    { return c.projectID }
