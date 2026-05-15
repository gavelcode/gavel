package get

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidQuery = failure.New("invalid project get query", failure.Validation)
