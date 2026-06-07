package resolveprincipal

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidQuery    = failure.New("invalid resolve principal query", failure.Validation)
	ErrUnauthenticated = failure.New("unauthenticated", failure.Validation)
)
