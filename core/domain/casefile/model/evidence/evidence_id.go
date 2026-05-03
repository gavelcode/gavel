package evidence

import (
	"fmt"

	"github.com/google/uuid"
)

type EvidenceID struct {
	value uuid.UUID
}

func NewEvidenceID(value uuid.UUID) EvidenceID {
	return EvidenceID{value: value}
}

func ParseEvidenceID(s string) (EvidenceID, error) {
	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return EvidenceID{}, fmt.Errorf("%w: %v", ErrInvalidEvidence, err)
	}
	if parsedUUID == uuid.Nil {
		return EvidenceID{}, fmt.Errorf("%w: id must not be nil uuid", ErrInvalidEvidence)
	}
	return EvidenceID{value: parsedUUID}, nil
}

func (id EvidenceID) UUID() uuid.UUID { return id.value }

func (id EvidenceID) String() string { return id.value.String() }

func (id EvidenceID) Equal(other EvidenceID) bool { return id.value == other.value }
