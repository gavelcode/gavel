package apitoken

import (
	"fmt"
	"strings"
)

type Scope struct {
	value string
}

var (
	ScopeIngest      = Scope{value: "ingest"}
	ScopeRead        = Scope{value: "read"}
	ScopeAdmin       = Scope{value: "admin"}
	ScopeProjectSync = Scope{value: "project_sync"}
)

func NewScope(raw string) (Scope, error) {
	normalised := strings.ToLower(strings.TrimSpace(raw))
	switch normalised {
	case ScopeIngest.value:
		return ScopeIngest, nil
	case ScopeRead.value:
		return ScopeRead, nil
	case ScopeAdmin.value:
		return ScopeAdmin, nil
	case ScopeProjectSync.value:
		return ScopeProjectSync, nil
	case "":
		return Scope{}, fmt.Errorf("%w: scope must not be empty", ErrInvalid)
	default:
		return Scope{}, fmt.Errorf("%w: unknown scope %q", ErrInvalid, raw)
	}
}

func (s Scope) String() string { return s.value }

func (s Scope) Equal(other Scope) bool { return s.value == other.value }
