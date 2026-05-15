package loadgavelspace

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidQuery = failure.New("invalid loadgavelspace query", failure.Validation)
