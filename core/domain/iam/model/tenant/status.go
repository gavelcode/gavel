package tenant

import "fmt"

type Status struct {
	value string
}

var (
	StatusActive    = Status{value: "active"}
	StatusSuspended = Status{value: "suspended"}
)

var validTenantStatuses = map[string]Status{
	StatusActive.value:    StatusActive,
	StatusSuspended.value: StatusSuspended,
}

func NewStatus(raw string) (Status, error) {
	s, ok := validTenantStatuses[raw]
	if !ok {
		return Status{}, fmt.Errorf("%w: unknown tenant status %q", ErrInvalidTenant, raw)
	}
	return s, nil
}

func (s Status) String() string { return s.value }

func (s Status) Equal(other Status) bool { return s.value == other.value }

func (s Status) IsActive() bool { return s.value == StatusActive.value }
