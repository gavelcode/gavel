package issuetoken

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	TokenID     string
	UserID      string
	TenantID    string
	Name        string
	TokenPrefix string
	Scopes      []string
	PlainSecret string
	Events      []event.Event
}
