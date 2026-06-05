package session

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type SessionID struct {
	value uuid.UUID
}

func NewSessionID(value uuid.UUID) SessionID {
	return SessionID{value: value}
}

func ParseSessionID(s string) (SessionID, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return SessionID{}, fmt.Errorf("%w: session id must not be empty", ErrInvalid)
	}
	parsedUUID, err := uuid.Parse(trimmed)
	if err != nil {
		return SessionID{}, fmt.Errorf("%w: session id must be a uuid", ErrInvalid)
	}
	if parsedUUID == uuid.Nil {
		return SessionID{}, fmt.Errorf("%w: session id must not be the nil uuid", ErrInvalid)
	}
	return SessionID{value: parsedUUID}, nil
}

func (id SessionID) UUID() uuid.UUID        { return id.value }
func (id SessionID) String() string         { return id.value.String() }
func (id SessionID) Equal(o SessionID) bool { return id.value == o.value }
