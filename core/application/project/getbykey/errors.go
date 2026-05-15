package getbykey

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidQuery = failure.New("invalid project getbykey query", failure.Validation)
