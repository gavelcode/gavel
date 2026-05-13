package createcasefile

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	CaseFileID string
	Events     []event.Event
}
