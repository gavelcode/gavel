package classifyarch

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid classifyarch command", failure.Validation)
