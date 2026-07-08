package resolve

import (
	"fmt"
	"strings"
)

type Command struct {
	tenantID   string
	pleadingID string
	outcome    string
}

var validOutcomes = map[string]struct{}{
	"merged": {},
	"closed": {},
}

func NewCommand(tenantID, pleadingID, outcome string) (Command, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(pleadingID) == "" {
		return Command{}, fmt.Errorf("%w: pleadingID must not be empty", ErrInvalidCommand)
	}
	if _, ok := validOutcomes[outcome]; !ok {
		return Command{}, fmt.Errorf("%w: outcome must be merged or closed (got %q)", ErrInvalidCommand, outcome)
	}
	return Command{tenantID: tenantID, pleadingID: pleadingID, outcome: outcome}, nil
}

func (c Command) TenantID() string   { return c.tenantID }
func (c Command) PleadingID() string { return c.pleadingID }
func (c Command) Outcome() string    { return c.outcome }
