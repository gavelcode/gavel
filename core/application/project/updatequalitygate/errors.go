package updatequalitygate

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid update quality gate command", failure.Validation)
