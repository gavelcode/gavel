package provision

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	TenantID    string
	AdminUserID string
	Events      []event.Event
}
