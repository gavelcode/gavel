package collectevidence

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid collectevidence command", failure.Validation)
