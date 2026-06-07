package create

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	TenantID    string
	Slug        string
	DisplayName string
	Events      []event.Event
}
