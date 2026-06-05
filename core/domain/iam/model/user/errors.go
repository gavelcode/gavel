package user

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidUser       = failure.New("invalid user", failure.Validation)
	ErrInvalidEmail      = failure.New("invalid email", failure.Validation)
	ErrUserNotFound      = failure.New("user not found", failure.NotFound)
	ErrEmailAlreadyInUse = failure.New("email already in use", failure.Conflict)
)
