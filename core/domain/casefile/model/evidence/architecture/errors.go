package architecture

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidViolation = failure.New("invalid architecture violation", failure.Validation)
)
