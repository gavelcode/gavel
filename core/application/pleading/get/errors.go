package get

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidQuery = failure.New("invalid pleading get query", failure.Validation)
