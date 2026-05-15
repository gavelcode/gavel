package preparebaseline

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid preparebaseline command", failure.Validation)
