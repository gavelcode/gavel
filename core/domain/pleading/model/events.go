package model

import "time"

const (
	EventNameMerged = "pleading.merged"
	EventNameClosed = "pleading.closed"
)

type Merged struct {
	pleadingID PleadingID
	occurredAt time.Time
}

func NewMerged(pleadingID PleadingID, occurredAt time.Time) Merged {
	return Merged{pleadingID: pleadingID, occurredAt: occurredAt}
}

func (e Merged) EventName() string      { return EventNameMerged }
func (e Merged) OccurredAt() time.Time  { return e.occurredAt }
func (e Merged) PleadingID() PleadingID { return e.pleadingID }

type Closed struct {
	pleadingID PleadingID
	occurredAt time.Time
}

func NewClosed(pleadingID PleadingID, occurredAt time.Time) Closed {
	return Closed{pleadingID: pleadingID, occurredAt: occurredAt}
}

func (e Closed) EventName() string      { return EventNameClosed }
func (e Closed) OccurredAt() time.Time  { return e.occurredAt }
func (e Closed) PleadingID() PleadingID { return e.pleadingID }
