package archpolicy

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidLayer    = failure.New("invalid layer", failure.Validation)
	ErrInvalidDenyRule = failure.New("invalid deny rule", failure.Validation)
	ErrInvalidPolicy   = failure.New("invalid architecture policy", failure.Validation)
)
