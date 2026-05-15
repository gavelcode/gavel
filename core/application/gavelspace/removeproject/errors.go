package removeproject

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid remove project command", failure.Validation)
