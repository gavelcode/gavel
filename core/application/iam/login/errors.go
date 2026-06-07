package login

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidCommand     = failure.New("invalid login command", failure.Validation)
	ErrInvalidCredentials = failure.New("invalid credentials", failure.Validation)
)
