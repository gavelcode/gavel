package apitoken

import (
	"time"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

const (
	EventNameIssued  = "iam.token_issued"
	EventNameRevoked = "iam.token_revoked"
)

type Issued struct {
	tokenID    APITokenID
	tenantID   tenant.TenantID
	userID     user.UserID
	name       string
	occurredAt time.Time
}

func NewIssued(tokenID APITokenID, tenantID tenant.TenantID, userID user.UserID, name string, occurredAt time.Time) Issued {
	return Issued{tokenID: tokenID, tenantID: tenantID, userID: userID, name: name, occurredAt: occurredAt}
}

func (e Issued) EventName() string         { return EventNameIssued }
func (e Issued) OccurredAt() time.Time     { return e.occurredAt }
func (e Issued) TokenID() APITokenID       { return e.tokenID }
func (e Issued) TenantID() tenant.TenantID { return e.tenantID }
func (e Issued) UserID() user.UserID       { return e.userID }
func (e Issued) Name() string              { return e.name }

type Revoked struct {
	tokenID    APITokenID
	occurredAt time.Time
}

func NewRevoked(tokenID APITokenID, occurredAt time.Time) Revoked {
	return Revoked{tokenID: tokenID, occurredAt: occurredAt}
}

func (e Revoked) EventName() string     { return EventNameRevoked }
func (e Revoked) OccurredAt() time.Time { return e.occurredAt }
func (e Revoked) TokenID() APITokenID   { return e.tokenID }
