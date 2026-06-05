package apitoken

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalid  = failure.New("invalid api token", failure.Validation)
	ErrNotFound = failure.New("api token not found", failure.NotFound)
)
