package qualitygate

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidStrategy = failure.New("invalid strategy", failure.Validation)
	ErrInvalidRule     = failure.New("invalid quality gate rule", failure.Validation)
	ErrInvalidGate     = failure.New("invalid quality gate", failure.Validation)
)
