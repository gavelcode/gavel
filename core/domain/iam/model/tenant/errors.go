package tenant

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidTenant  = failure.New("invalid tenant", failure.Validation)
	ErrTenantNotFound = failure.New("tenant not found", failure.NotFound)
	ErrSlugTaken      = failure.New("tenant slug already taken", failure.Conflict)
)
