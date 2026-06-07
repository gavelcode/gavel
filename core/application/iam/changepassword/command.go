package changepassword

import (
	"fmt"
	"strings"
	"time"
)

const minPasswordLength = 8

type Command struct {
	userID          string
	currentPassword string
	newPassword     string
	occurredAt      time.Time
}

func NewCommand(userID, currentPassword, newPassword string, occurredAt time.Time) (Command, error) {
	if strings.TrimSpace(userID) == "" {
		return Command{}, fmt.Errorf("%w: userID must not be empty", ErrInvalidCommand)
	}
	if currentPassword == "" {
		return Command{}, fmt.Errorf("%w: currentPassword must not be empty", ErrInvalidCommand)
	}
	if len(newPassword) < minPasswordLength {
		return Command{}, fmt.Errorf("%w: newPassword must be at least 8 characters", ErrInvalidCommand)
	}
	if currentPassword == newPassword {
		return Command{}, fmt.Errorf("%w: newPassword must differ from currentPassword", ErrInvalidCommand)
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	return Command{
		userID:          userID,
		currentPassword: currentPassword,
		newPassword:     newPassword,
		occurredAt:      occurredAt,
	}, nil
}

func (c Command) UserID() string          { return c.userID }
func (c Command) CurrentPassword() string { return c.currentPassword }
func (c Command) NewPassword() string     { return c.newPassword }
func (c Command) OccurredAt() time.Time   { return c.occurredAt }
