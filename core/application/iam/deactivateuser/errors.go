package deactivateuser

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid deactivate user command", failure.Validation)
