package event

import (
	"time"

	"github.com/usegavel/gavel/core/domain/shared/event"
)

type Event struct {
	Name       string
	OccurredAt time.Time
}

func EventsFromDomain(events []event.DomainEvent) []Event {
	out := make([]Event, 0, len(events))
	for _, e := range events {
		out = append(out, Event{
			Name:       e.EventName(),
			OccurredAt: e.OccurredAt(),
		})
	}
	return out
}
