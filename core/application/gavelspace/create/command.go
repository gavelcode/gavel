package create

import (
	"fmt"
	"strings"
)

type Command struct {
	name string
}

func NewCommand(name string) (Command, error) {
	if strings.TrimSpace(name) == "" {
		return Command{}, fmt.Errorf("%w: name must not be empty", ErrInvalidCommand)
	}
	return Command{name: name}, nil
}

func (c Command) Name() string { return c.name }
