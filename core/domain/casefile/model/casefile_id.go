package model

import (
	"fmt"

	"github.com/google/uuid"
)

type CaseFileID struct {
	value uuid.UUID
}

func NewCaseFileID(value uuid.UUID) CaseFileID {
	return CaseFileID{value: value}
}

func ParseCaseFileID(s string) (CaseFileID, error) {
	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return CaseFileID{}, fmt.Errorf("%w: %v", ErrInvalidCaseFile, err)
	}
	if parsedUUID == uuid.Nil {
		return CaseFileID{}, fmt.Errorf("%w: id must not be nil uuid", ErrInvalidCaseFile)
	}
	return CaseFileID{value: parsedUUID}, nil
}

func (id CaseFileID) UUID() uuid.UUID { return id.value }

func (id CaseFileID) String() string { return id.value.String() }

func (id CaseFileID) Equal(other CaseFileID) bool { return id.value == other.value }
