package event_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	appevent "github.com/usegavel/gavel/core/application/shared/event"
	domainevent "github.com/usegavel/gavel/core/domain/shared/event"
)

type fakeEvent struct {
	name       string
	occurredAt time.Time
}

func (f fakeEvent) EventName() string    { return f.name }
func (f fakeEvent) OccurredAt() time.Time { return f.occurredAt }

func TestEventsFromDomain(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	events := appevent.EventsFromDomain([]domainevent.DomainEvent{
		fakeEvent{name: "CaseFileCreated", occurredAt: now},
		fakeEvent{name: "VerdictReached", occurredAt: now.Add(time.Second)},
	})

	assert.Len(t, events, 2)
	assert.Equal(t, "CaseFileCreated", events[0].Name)
	assert.Equal(t, now, events[0].OccurredAt)
	assert.Equal(t, "VerdictReached", events[1].Name)
	assert.Equal(t, now.Add(time.Second), events[1].OccurredAt)
}

func TestEventsFromDomain_EmptySlice(t *testing.T) {
	events := appevent.EventsFromDomain(nil)
	assert.Empty(t, events)
}

func TestEventsFromDomain_NilSliceReturnsEmpty(t *testing.T) {
	var domainEvents []domainevent.DomainEvent
	events := appevent.EventsFromDomain(domainEvents)
	assert.NotNil(t, events)
	assert.Empty(t, events)
}
