package apitoken

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/shared/event"
)

type APIToken struct {
	id          APITokenID
	tenantID    tenant.TenantID
	userID      user.UserID
	name        string
	tokenHash   SecretHash
	tokenPrefix string
	scopes      Scopes
	createdAt   time.Time
	expiresAt   time.Time
	lastUsedAt  time.Time
	isRevoked   bool
	events      []event.DomainEvent
}

func NewAPIToken(
	secret Secret,
	tenantID tenant.TenantID,
	userID user.UserID,
	name string,
	scopes Scopes,
	createdAt time.Time,
	expiresAt time.Time,
) (APIToken, error) {
	if err := validateAPITokenFields(tenantID, userID, name, scopes, createdAt, expiresAt); err != nil {
		return APIToken{}, err
	}
	apiTokenID := NewAPITokenID(uuid.New())
	trimmedName := strings.TrimSpace(name)
	hash := HashSecret(secret)
	scopeCopy := make(Scopes, len(scopes))
	copy(scopeCopy, scopes)
	return APIToken{
		id:          apiTokenID,
		tenantID:    tenantID,
		userID:      userID,
		name:        trimmedName,
		tokenHash:   hash,
		tokenPrefix: secret.Prefix(),
		scopes:      scopeCopy,
		createdAt:   createdAt,
		expiresAt:   expiresAt,
		events:      []event.DomainEvent{NewIssued(apiTokenID, tenantID, userID, trimmedName, createdAt)},
	}, nil
}

func ReconstituteAPIToken(
	apiTokenID APITokenID,
	tenantID tenant.TenantID,
	userID user.UserID,
	name string,
	tokenHash SecretHash,
	tokenPrefix string,
	scopes Scopes,
	createdAt time.Time,
	expiresAt time.Time,
	lastUsedAt time.Time,
	isRevoked bool,
) (APIToken, error) {
	if strings.TrimSpace(tokenPrefix) == "" {
		return APIToken{}, fmt.Errorf("%w: tokenPrefix must not be empty", ErrInvalid)
	}
	if err := validateAPITokenFields(tenantID, userID, name, scopes, createdAt, expiresAt); err != nil {
		return APIToken{}, err
	}
	scopeCopy := make(Scopes, len(scopes))
	copy(scopeCopy, scopes)
	return APIToken{
		id:          apiTokenID,
		tenantID:    tenantID,
		userID:      userID,
		name:        strings.TrimSpace(name),
		tokenHash:   tokenHash,
		tokenPrefix: tokenPrefix,
		scopes:      scopeCopy,
		createdAt:   createdAt,
		expiresAt:   expiresAt,
		lastUsedAt:  lastUsedAt,
		isRevoked:   isRevoked,
	}, nil
}

func validateAPITokenFields(tenantID tenant.TenantID, userID user.UserID, name string, scopes Scopes, createdAt, expiresAt time.Time) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name must not be empty", ErrInvalid)
	}
	if len(scopes) == 0 {
		return fmt.Errorf("%w: scopes must not be empty", ErrInvalid)
	}
	if createdAt.IsZero() {
		return fmt.Errorf("%w: createdAt must not be zero", ErrInvalid)
	}
	if !expiresAt.IsZero() && !expiresAt.After(createdAt) {
		return fmt.Errorf("%w: expiresAt must be zero (never expires) or strictly after createdAt", ErrInvalid)
	}
	return nil
}

func (token *APIToken) ID() APITokenID            { return token.id }
func (token *APIToken) TenantID() tenant.TenantID { return token.tenantID }
func (token *APIToken) UserID() user.UserID       { return token.userID }
func (token *APIToken) Name() string              { return token.name }
func (token *APIToken) TokenHash() SecretHash     { return token.tokenHash }
func (token *APIToken) TokenPrefix() string       { return token.tokenPrefix }

func (token *APIToken) Scopes() Scopes {
	copied := make(Scopes, len(token.scopes))
	copy(copied, token.scopes)
	return copied
}

func (token *APIToken) CreatedAt() time.Time  { return token.createdAt }
func (token *APIToken) ExpiresAt() time.Time  { return token.expiresAt }
func (token *APIToken) LastUsedAt() time.Time { return token.lastUsedAt }
func (token *APIToken) IsRevoked() bool       { return token.isRevoked }

func (token *APIToken) IsExpired(referenceTime time.Time) bool {
	if token.isRevoked {
		return true
	}
	if token.expiresAt.IsZero() {
		return false
	}
	return !referenceTime.Before(token.expiresAt)
}

func (token *APIToken) HasScope(s Scope) bool {
	return token.scopes.Contains(s)
}

func (token *APIToken) TouchUsed(at time.Time) error {
	if at.IsZero() {
		return fmt.Errorf("%w: usage timestamp must not be zero", ErrInvalid)
	}
	token.lastUsedAt = at
	return nil
}

func (token *APIToken) Revoke(occurredAt time.Time) error {
	if token.isRevoked {
		return fmt.Errorf("%w: token is already revoked", ErrInvalid)
	}
	if occurredAt.IsZero() {
		return fmt.Errorf("%w: occurredAt must not be zero", ErrInvalid)
	}
	token.isRevoked = true
	token.events = append(token.events, NewRevoked(token.id, occurredAt))
	return nil
}

func (token *APIToken) Events() []event.DomainEvent {
	copied := make([]event.DomainEvent, len(token.events))
	copy(copied, token.events)
	return copied
}

func (token *APIToken) ClearEvents() {
	token.events = nil
}
