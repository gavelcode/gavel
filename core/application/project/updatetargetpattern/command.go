package updatetargetpattern

import (
	"fmt"
	"strings"
)

type Command struct {
	projectID     string
	targetPattern string
}

func NewCommand(projectID, targetPattern string) (Command, error) {
	if strings.TrimSpace(projectID) == "" {
		return Command{}, fmt.Errorf("%w: projectID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(targetPattern) == "" {
		return Command{}, fmt.Errorf("%w: targetPattern must not be empty", ErrInvalidCommand)
	}
	return Command{
		projectID:     projectID,
		targetPattern: targetPattern,
	}, nil
}

func (c Command) ProjectID() string { return c.projectID }

func (c Command) TargetPattern() string { return c.targetPattern }
