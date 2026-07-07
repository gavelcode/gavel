package activate

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid activate tenant command", failure.Validation)
