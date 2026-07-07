package provision

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	slug             string
	displayName      string
	adminEmail       string
	adminDisplayName string
	adminPassword    string
	occurredAt       time.Time
}

func NewCommand(slug, displayName, adminEmail, adminDisplayName, adminPassword string, occurredAt time.Time) (Command, error) {
	required := []struct{ name, value string }{
		{"slug", slug},
		{"displayName", displayName},
		{"adminEmail", adminEmail},
		{"adminDisplayName", adminDisplayName},
		{"adminPassword", adminPassword},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return Command{}, fmt.Errorf("%w: %s must not be empty", ErrInvalidCommand, field.name)
		}
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	return Command{
		slug:             slug,
		displayName:      displayName,
		adminEmail:       adminEmail,
		adminDisplayName: adminDisplayName,
		adminPassword:    adminPassword,
		occurredAt:       occurredAt,
	}, nil
}

func (c Command) Slug() string             { return c.slug }
func (c Command) DisplayName() string      { return c.displayName }
func (c Command) AdminEmail() string       { return c.adminEmail }
func (c Command) AdminDisplayName() string { return c.adminDisplayName }
func (c Command) AdminPassword() string    { return c.adminPassword }
func (c Command) OccurredAt() time.Time    { return c.occurredAt }
