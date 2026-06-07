package issuetoken

import (
	"fmt"
	"strings"
	"time"
)

const (
	ScopeAdmin = "admin"
	RoleAdmin  = "admin"
)

type Command struct {
	userID     string
	name       string
	scopes     []string
	occurredAt time.Time
	expiresAt  time.Time
}

func NewCommand(userID, name string, scopes []string, occurredAt, expiresAt time.Time) (Command, error) {
	if strings.TrimSpace(userID) == "" {
		return Command{}, fmt.Errorf("%w: userID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(name) == "" {
		return Command{}, fmt.Errorf("%w: name must not be empty", ErrInvalidCommand)
	}
	if len(scopes) == 0 {
		return Command{}, fmt.Errorf("%w: scopes must not be empty", ErrInvalidCommand)
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	scopeCopy := make([]string, len(scopes))
	copy(scopeCopy, scopes)
	return Command{
		userID:     userID,
		name:       name,
		scopes:     scopeCopy,
		occurredAt: occurredAt,
		expiresAt:  expiresAt,
	}, nil
}

func (c Command) UserID() string        { return c.userID }
func (c Command) Name() string          { return c.name }
func (c Command) OccurredAt() time.Time { return c.occurredAt }
func (c Command) ExpiresAt() time.Time  { return c.expiresAt }

func (c Command) Scopes() []string {
	out := make([]string, len(c.scopes))
	copy(out, c.scopes)
	return out
}
