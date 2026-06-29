package toolexecution

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidFailure = failure.New("invalid tool execution failure", failure.Validation)
)
