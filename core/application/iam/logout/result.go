package logout

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	SessionID string
	Events    []event.Event
}
