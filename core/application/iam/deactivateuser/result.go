package deactivateuser

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	UserID string
	Events []event.Event
}
