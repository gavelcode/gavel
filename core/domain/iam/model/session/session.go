package session

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/shared/event"
)

type Session struct {
	id         SessionID
	tokenHash  TokenHash
	userID     user.UserID
	userAgent  string
	ipAddress  string
	createdAt  time.Time
	expiresAt  time.Time
	lastSeenAt time.Time
	isRevoked  bool
	events     []event.DomainEvent
}

func NewSession(token Token, userID user.UserID, userAgent, ipAddress string, createdAt, expiresAt time.Time) (Session, error) {
	if createdAt.IsZero() {
		return Session{}, fmt.Errorf("%w: createdAt must not be zero", ErrInvalid)
	}
	if expiresAt.IsZero() {
		return Session{}, fmt.Errorf("%w: expiresAt must not be zero", ErrInvalid)
	}
	if !expiresAt.After(createdAt) {
		return Session{}, fmt.Errorf("%w: expiresAt must be strictly after createdAt", ErrInvalid)
	}
	hash := HashToken(token)
	sessionID := NewSessionID(uuid.New())
	return Session{
		id:         sessionID,
		tokenHash:  hash,
		userID:     userID,
		userAgent:  userAgent,
		ipAddress:  ipAddress,
		createdAt:  createdAt,
		expiresAt:  expiresAt,
		lastSeenAt: createdAt,
		events:     []event.DomainEvent{NewCreated(sessionID, hash, userID, expiresAt, createdAt)},
	}, nil
}

func ReconstituteSession(sessionID SessionID, tokenHash TokenHash, userID user.UserID, userAgent, ipAddress string, createdAt, expiresAt, lastSeenAt time.Time, isRevoked bool) (Session, error) {
	if createdAt.IsZero() {
		return Session{}, fmt.Errorf("%w: createdAt must not be zero", ErrInvalid)
	}
	if expiresAt.IsZero() {
		return Session{}, fmt.Errorf("%w: expiresAt must not be zero", ErrInvalid)
	}
	if lastSeenAt.IsZero() {
		return Session{}, fmt.Errorf("%w: lastSeenAt must not be zero", ErrInvalid)
	}
	return Session{
		id:         sessionID,
		tokenHash:  tokenHash,
		userID:     userID,
		userAgent:  userAgent,
		ipAddress:  ipAddress,
		createdAt:  createdAt,
		expiresAt:  expiresAt,
		lastSeenAt: lastSeenAt,
		isRevoked:  isRevoked,
	}, nil
}

func (s *Session) ID() SessionID         { return s.id }
func (s *Session) TokenHash() TokenHash  { return s.tokenHash }
func (s *Session) UserID() user.UserID   { return s.userID }
func (s *Session) UserAgent() string     { return s.userAgent }
func (s *Session) IPAddress() string     { return s.ipAddress }
func (s *Session) CreatedAt() time.Time  { return s.createdAt }
func (s *Session) ExpiresAt() time.Time  { return s.expiresAt }
func (s *Session) LastSeenAt() time.Time { return s.lastSeenAt }
func (s *Session) IsRevoked() bool       { return s.isRevoked }

func (s *Session) IsExpired(referenceTime time.Time) bool {
	if s.isRevoked {
		return true
	}
	return !referenceTime.Before(s.expiresAt)
}

func (s *Session) Touch(at time.Time) error {
	if at.IsZero() {
		return fmt.Errorf("%w: touch timestamp must not be zero", ErrInvalid)
	}
	s.lastSeenAt = at
	return nil
}

func (s *Session) Revoke(occurredAt time.Time) error {
	if s.isRevoked {
		return fmt.Errorf("%w: session is already revoked", ErrInvalid)
	}
	if occurredAt.IsZero() {
		return fmt.Errorf("%w: occurredAt must not be zero", ErrInvalid)
	}
	s.isRevoked = true
	s.events = append(s.events, NewRevoked(s.id, s.tokenHash, occurredAt))
	return nil
}

func (s *Session) Events() []event.DomainEvent {
	copied := make([]event.DomainEvent, len(s.events))
	copy(copied, s.events)
	return copied
}

func (s *Session) ClearEvents() {
	s.events = nil
}
