package analyzetarget

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid analyzetarget command", failure.Validation)
