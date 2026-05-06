package model

import (
	"fmt"

	"github.com/google/uuid"
)

type ProjectID struct {
	value uuid.UUID
}

func NewProjectID(value uuid.UUID) ProjectID {
	return ProjectID{value: value}
}

func ParseProjectID(s string) (ProjectID, error) {
	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return ProjectID{}, fmt.Errorf("%w: %v", ErrInvalidProject, err)
	}
	if parsedUUID == uuid.Nil {
		return ProjectID{}, fmt.Errorf("%w: id must not be nil uuid", ErrInvalidProject)
	}
	return ProjectID{value: parsedUUID}, nil
}

func (id ProjectID) UUID() uuid.UUID { return id.value }

func (id ProjectID) String() string { return id.value.String() }

func (id ProjectID) Equal(other ProjectID) bool { return id.value == other.value }
