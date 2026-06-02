package file

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	PleadingID string
	Status     string
	Events     []event.Event
}
