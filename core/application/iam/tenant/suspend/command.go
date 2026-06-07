package suspend

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	tenantID   string
	occurredAt time.Time
}

func NewCommand(tenantID string, occurredAt time.Time) (Command, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	return Command{tenantID: tenantID, occurredAt: occurredAt}, nil
}

func (c Command) TenantID() string      { return c.tenantID }
func (c Command) OccurredAt() time.Time { return c.occurredAt }
