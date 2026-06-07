package login

import (
	"time"

	"github.com/usegavel/gavel/core/application/shared/event"
)

type Result struct {
	UserID             string
	TenantID           string
	Email              string
	DisplayName        string
	Role               string
	MustChangePassword bool
	SessionToken       string
	SessionExpiresAt   time.Time
	Events             []event.Event
}
