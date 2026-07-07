package activate

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	TenantID string
	Events   []event.Event
}
