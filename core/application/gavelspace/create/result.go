package create

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	Name   string
	Events []event.Event
}
