package logout

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	sessionToken string
	occurredAt   time.Time
}

func NewCommand(sessionToken string, occurredAt time.Time) (Command, error) {
	if strings.TrimSpace(sessionToken) == "" {
		return Command{}, fmt.Errorf("%w: sessionToken must not be empty", ErrInvalidCommand)
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	return Command{sessionToken: sessionToken, occurredAt: occurredAt}, nil
}

func (c Command) SessionToken() string  { return c.sessionToken }
func (c Command) OccurredAt() time.Time { return c.occurredAt }
