package finding

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidFinding       = failure.New("invalid finding", failure.Validation)
	ErrInvalidSeverity      = failure.New("invalid severity", failure.Validation)
	ErrInvalidFingerprintID = failure.New("invalid fingerprint", failure.Validation)
	ErrInvalidContent       = failure.New("invalid findings content", failure.Validation)
)
