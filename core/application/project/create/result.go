package create

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	ProjectID string
	Events    []event.Event
}
