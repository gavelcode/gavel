package revoketoken

import "github.com/usegavel/gavel/core/application/shared/event"

type Result struct {
	TokenID string
	Events  []event.Event
}
