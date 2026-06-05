package user

import (
	"fmt"

	"github.com/google/uuid"
)

type UserID struct {
	value uuid.UUID
}

func NewUserID(value uuid.UUID) UserID {
	return UserID{value: value}
}

func ParseUserID(s string) (UserID, error) {
	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, fmt.Errorf("%w: %v", ErrInvalidUser, err)
	}
	if parsedUUID == uuid.Nil {
		return UserID{}, fmt.Errorf("%w: id must not be nil uuid", ErrInvalidUser)
	}
	return UserID{value: parsedUUID}, nil
}

func (id UserID) UUID() uuid.UUID { return id.value }

func (id UserID) String() string { return id.value.String() }

func (id UserID) Equal(other UserID) bool { return id.value == other.value }
