package suspend

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid suspend tenant command", failure.Validation)
