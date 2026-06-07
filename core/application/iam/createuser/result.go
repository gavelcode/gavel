package createuser

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	UserID      string
	TenantID    string
	Email       string
	DisplayName string
	Role        string
	Events      []event.Event
}
