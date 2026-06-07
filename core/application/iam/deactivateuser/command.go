package deactivateuser

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	userID     string
	occurredAt time.Time
}

func NewCommand(userID string, occurredAt time.Time) (Command, error) {
	if strings.TrimSpace(userID) == "" {
		return Command{}, fmt.Errorf("%w: userID must not be empty", ErrInvalidCommand)
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	return Command{userID: userID, occurredAt: occurredAt}, nil
}

func (c Command) UserID() string        { return c.userID }
func (c Command) OccurredAt() time.Time { return c.occurredAt }
