package create

import (
	"fmt"
	"strings"
)

type Command struct {
	key           string
	name          string
	targetPattern string
}

func NewCommand(key, name, targetPattern string) (Command, error) {
	if strings.TrimSpace(key) == "" {
		return Command{}, fmt.Errorf("%w: key must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(name) == "" {
		return Command{}, fmt.Errorf("%w: name must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(targetPattern) == "" {
		return Command{}, fmt.Errorf("%w: targetPattern must not be empty", ErrInvalidCommand)
	}
	return Command{key: key, name: name, targetPattern: targetPattern}, nil
}

func (c Command) Key() string           { return c.key }
func (c Command) Name() string          { return c.name }
func (c Command) TargetPattern() string { return c.targetPattern }
