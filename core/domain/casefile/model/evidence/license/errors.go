package license

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidDependency = failure.New("invalid dependency license", failure.Validation)
)
