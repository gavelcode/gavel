package session

import (
	"time"

	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

const (
	EventNameCreated = "iam.session_created"
	EventNameRevoked = "iam.session_revoked"
)

type Created struct {
	sessionID  SessionID
	tokenHash  TokenHash
	userID     user.UserID
	expiresAt  time.Time
	occurredAt time.Time
}

func NewCreated(sessionID SessionID, tokenHash TokenHash, userID user.UserID, expiresAt, occurredAt time.Time) Created {
	return Created{sessionID: sessionID, tokenHash: tokenHash, userID: userID, expiresAt: expiresAt, occurredAt: occurredAt}
}

func (e Created) EventName() string     { return EventNameCreated }
func (e Created) OccurredAt() time.Time { return e.occurredAt }
func (e Created) SessionID() SessionID  { return e.sessionID }
func (e Created) TokenHash() TokenHash  { return e.tokenHash }
func (e Created) UserID() user.UserID   { return e.userID }
func (e Created) ExpiresAt() time.Time  { return e.expiresAt }

type Revoked struct {
	sessionID  SessionID
	tokenHash  TokenHash
	occurredAt time.Time
}

func NewRevoked(sessionID SessionID, tokenHash TokenHash, occurredAt time.Time) Revoked {
	return Revoked{sessionID: sessionID, tokenHash: tokenHash, occurredAt: occurredAt}
}

func (e Revoked) EventName() string     { return EventNameRevoked }
func (e Revoked) OccurredAt() time.Time { return e.occurredAt }
func (e Revoked) SessionID() SessionID  { return e.sessionID }
func (e Revoked) TokenHash() TokenHash  { return e.tokenHash }
