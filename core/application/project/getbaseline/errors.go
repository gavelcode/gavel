package getbaseline

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidQuery = failure.New("invalid project getbaseline query", failure.Validation)
