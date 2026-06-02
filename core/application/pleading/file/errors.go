package file

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid file pleading command", failure.Validation)
