package create

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	slug        string
	displayName string
	occurredAt  time.Time
}

func NewCommand(slug, displayName string, occurredAt time.Time) (Command, error) {
	if strings.TrimSpace(slug) == "" {
		return Command{}, fmt.Errorf("%w: slug must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(displayName) == "" {
		return Command{}, fmt.Errorf("%w: display name must not be empty", ErrInvalidCommand)
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	return Command{slug: slug, displayName: displayName, occurredAt: occurredAt}, nil
}

func (c Command) Slug() string          { return c.slug }
func (c Command) DisplayName() string   { return c.displayName }
func (c Command) OccurredAt() time.Time { return c.occurredAt }
