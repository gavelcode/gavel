package create

import (
	"fmt"
	"strings"
)

type Command struct {
	tenantID      string
	key           string
	name          string
	targetPattern string
}

func NewCommand(tenantID, key, name, targetPattern string) (Command, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(key) == "" {
		return Command{}, fmt.Errorf("%w: key must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(name) == "" {
		return Command{}, fmt.Errorf("%w: name must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(targetPattern) == "" {
		return Command{}, fmt.Errorf("%w: targetPattern must not be empty", ErrInvalidCommand)
	}
	return Command{tenantID: tenantID, key: key, name: name, targetPattern: targetPattern}, nil
}

func (c Command) TenantID() string      { return c.tenantID }
func (c Command) Key() string           { return c.key }
func (c Command) Name() string          { return c.name }
func (c Command) TargetPattern() string { return c.targetPattern }
