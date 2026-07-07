package provision

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid provision tenant command", failure.Validation)
