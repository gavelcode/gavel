package create

import (
	"fmt"
	"strings"
)

type Command struct {
	tenantID string
	name     string
}

func NewCommand(tenantID, name string) (Command, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(name) == "" {
		return Command{}, fmt.Errorf("%w: name must not be empty", ErrInvalidCommand)
	}
	return Command{tenantID: tenantID, name: name}, nil
}

func (c Command) TenantID() string { return c.tenantID }
func (c Command) Name() string     { return c.name }
