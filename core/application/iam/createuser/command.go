package createuser

import (
	"fmt"
	"strings"
	"time"
)

const DefaultRole = "viewer"

const minPasswordLength = 8

type Command struct {
	tenantID           string
	email              string
	displayName        string
	role               string
	plainPassword      string
	mustChangePassword bool
	occurredAt         time.Time
}

func NewCommand(tenantID, email, displayName, role, plainPassword string, mustChangePassword bool, occurredAt time.Time) (Command, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Command{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(email) == "" {
		return Command{}, fmt.Errorf("%w: email must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(displayName) == "" {
		return Command{}, fmt.Errorf("%w: displayName must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(role) == "" {
		return Command{}, fmt.Errorf("%w: role must not be empty", ErrInvalidCommand)
	}
	if len(plainPassword) < minPasswordLength {
		return Command{}, fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidCommand)
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	return Command{
		tenantID:           tenantID,
		email:              email,
		displayName:        displayName,
		role:               role,
		plainPassword:      plainPassword,
		mustChangePassword: mustChangePassword,
		occurredAt:         occurredAt,
	}, nil
}

func (c Command) TenantID() string         { return c.tenantID }
func (c Command) Email() string            { return c.email }
func (c Command) DisplayName() string      { return c.displayName }
func (c Command) Role() string             { return c.role }
func (c Command) PlainPassword() string    { return c.plainPassword }
func (c Command) MustChangePassword() bool { return c.mustChangePassword }
func (c Command) OccurredAt() time.Time    { return c.occurredAt }
