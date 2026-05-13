package ingestevidence

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	EvidenceIDs []string
	Events      []event.Event
}
