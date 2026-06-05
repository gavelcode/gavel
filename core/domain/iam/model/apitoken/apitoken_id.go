package apitoken

import (
	"fmt"

	"github.com/google/uuid"
)

type APITokenID struct {
	value uuid.UUID
}

func NewAPITokenID(value uuid.UUID) APITokenID {
	return APITokenID{value: value}
}

func ParseAPITokenID(s string) (APITokenID, error) {
	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return APITokenID{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if parsedUUID == uuid.Nil {
		return APITokenID{}, fmt.Errorf("%w: id must not be nil uuid", ErrInvalid)
	}
	return APITokenID{value: parsedUUID}, nil
}

func (id APITokenID) UUID() uuid.UUID { return id.value }

func (id APITokenID) String() string { return id.value.String() }

func (id APITokenID) Equal(other APITokenID) bool { return id.value == other.value }
