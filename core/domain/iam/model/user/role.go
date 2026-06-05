package user

import (
	"fmt"
	"strings"
)

type Role struct {
	value string
}

var (
	RoleAdmin      = Role{value: "admin"}
	RoleMaintainer = Role{value: "maintainer"}
	RoleViewer     = Role{value: "viewer"}
)

func NewRole(raw string) (Role, error) {
	normalised := strings.ToLower(strings.TrimSpace(raw))
	switch normalised {
	case RoleAdmin.value:
		return RoleAdmin, nil
	case RoleMaintainer.value:
		return RoleMaintainer, nil
	case RoleViewer.value:
		return RoleViewer, nil
	case "":
		return Role{}, fmt.Errorf("%w: role must not be empty", ErrInvalidUser)
	default:
		return Role{}, fmt.Errorf("%w: unknown role %q", ErrInvalidUser, raw)
	}
}

func (r Role) String() string { return r.value }

func (r Role) Equal(other Role) bool { return r.value == other.value }

func (r Role) IsAdmin() bool { return r.value == RoleAdmin.value }
