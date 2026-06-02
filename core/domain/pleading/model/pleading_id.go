package model

import (
	"fmt"

	"github.com/google/uuid"
)

type PleadingID struct {
	value uuid.UUID
}

func NewPleadingID(value uuid.UUID) PleadingID {
	return PleadingID{value: value}
}

func ParsePleadingID(s string) (PleadingID, error) {
	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return PleadingID{}, fmt.Errorf("%w: %v", ErrInvalidPleading, err)
	}
	if parsedUUID == uuid.Nil {
		return PleadingID{}, fmt.Errorf("%w: id must not be nil uuid", ErrInvalidPleading)
	}
	return PleadingID{value: parsedUUID}, nil
}

func (id PleadingID) UUID() uuid.UUID { return id.value }

func (id PleadingID) String() string { return id.value.String() }

func (id PleadingID) Equal(other PleadingID) bool { return id.value == other.value }
