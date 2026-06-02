package resolve

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid resolve pleading command", failure.Validation)
