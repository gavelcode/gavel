package updatetargetpattern

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid update target pattern command", failure.Validation)
