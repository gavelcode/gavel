package updatelanguages

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid update languages command", failure.Validation)
