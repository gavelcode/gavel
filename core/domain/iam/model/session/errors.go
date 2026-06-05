package session

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalid  = failure.New("invalid session", failure.Validation)
	ErrNotFound = failure.New("session not found", failure.NotFound)
)
