package login

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	tenantSlug    string
	email         string
	plainPassword string
	userAgent     string
	ipAddress     string
	occurredAt    time.Time
	sessionTTL    time.Duration
}

func NewCommand(tenantSlug, email, plainPassword, userAgent, ipAddress string, occurredAt time.Time, sessionTTL time.Duration) (Command, error) {
	if strings.TrimSpace(tenantSlug) == "" {
		return Command{}, fmt.Errorf("%w: tenantSlug must not be empty", ErrInvalidCommand)
	}
	if strings.TrimSpace(email) == "" {
		return Command{}, fmt.Errorf("%w: email must not be empty", ErrInvalidCommand)
	}
	if plainPassword == "" {
		return Command{}, fmt.Errorf("%w: password must not be empty", ErrInvalidCommand)
	}
	if occurredAt.IsZero() {
		return Command{}, fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidCommand)
	}
	if sessionTTL <= 0 {
		return Command{}, fmt.Errorf("%w: sessionTTL must be positive", ErrInvalidCommand)
	}
	return Command{
		tenantSlug:    tenantSlug,
		email:         email,
		plainPassword: plainPassword,
		userAgent:     userAgent,
		ipAddress:     ipAddress,
		occurredAt:    occurredAt,
		sessionTTL:    sessionTTL,
	}, nil
}

func (c Command) TenantSlug() string        { return c.tenantSlug }
func (c Command) Email() string             { return c.email }
func (c Command) PlainPassword() string     { return c.plainPassword }
func (c Command) UserAgent() string         { return c.userAgent }
func (c Command) IPAddress() string         { return c.ipAddress }
func (c Command) OccurredAt() time.Time     { return c.occurredAt }
func (c Command) SessionTTL() time.Duration { return c.sessionTTL }
