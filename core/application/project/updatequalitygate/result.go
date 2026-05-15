package updatequalitygate

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	Changed bool
	Events  []event.Event
}
