package updatequalitygate

import (
	"fmt"
	"strings"

	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

type Command struct {
	projectID   string
	qualityGate qualitygate.Gate
}

func NewCommand(projectID string, input Input) (Command, error) {
	if strings.TrimSpace(projectID) == "" {
		return Command{}, fmt.Errorf("%w: projectID must not be empty", ErrInvalidCommand)
	}
	qualityGate, err := toDomain(input)
	if err != nil {
		return Command{}, fmt.Errorf("%w: %s", ErrInvalidCommand, err.Error())
	}
	return Command{
		projectID:   projectID,
		qualityGate: qualityGate,
	}, nil
}

func (c Command) ProjectID() string      { return c.projectID }
func (c Command) Gate() qualitygate.Gate { return c.qualityGate }
