package registerproject

import (
	"fmt"
	"strings"
)

type Command struct {
	gavelspaceName string
	projectID      string
	targetPattern  string
}

func NewCommand(gavelspaceName, projectID, targetPattern string) (Command, error) {
	if strings.TrimSpace(gavelspaceName) == "" {
		return Command{}, fmt.Errorf("%w: gavelspaceName must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(projectID) == "" {
		return Command{}, fmt.Errorf("%w: projectID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(targetPattern) == "" {
		return Command{}, fmt.Errorf("%w: targetPattern must not be empty", ErrInvalidCommand)
	}
	return Command{
		gavelspaceName: gavelspaceName,
		projectID:      projectID,
		targetPattern:  targetPattern,
	}, nil
}

func (c Command) GavelspaceID() string  { return c.gavelspaceName }
func (c Command) ProjectID() string     { return c.projectID }
func (c Command) TargetPattern() string { return c.targetPattern }
