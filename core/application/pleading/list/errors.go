package list

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidQuery = failure.New("invalid pleading list query", failure.Validation)
