package sarif

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrDecodeSARIF   = failure.New("decode sarif", failure.Validation)
	ErrInvalidResult = failure.New("invalid sarif result", failure.Validation)
)
