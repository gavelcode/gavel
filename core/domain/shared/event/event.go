package event

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}
