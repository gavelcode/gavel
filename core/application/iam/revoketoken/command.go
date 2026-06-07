package revoketoken

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	tokenID      string
	callerUserID string
	occurredAt   time.Time
}

func NewCommand(tokenID, callerUserID string, occurredAt time.Time) (Command, error) {
	if strings.TrimSpace(tokenID) == "" {
		return Command{}, fmt.Errorf("%w: tokenID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(callerUserID) == "" {
		return Command{}, fmt.Errorf("%w: callerUserID must not be empty", ErrInvalidCommand)
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	return Command{tokenID: tokenID, callerUserID: callerUserID, occurredAt: occurredAt}, nil
}

func (c Command) TokenID() string       { return c.tokenID }
func (c Command) CallerUserID() string  { return c.callerUserID }
func (c Command) OccurredAt() time.Time { return c.occurredAt }
