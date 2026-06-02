package resolve

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	Changed bool
	Status  string
	Events  []event.Event
}
