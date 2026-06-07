package revoketoken

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidCommand = failure.New("invalid revoke token command", failure.Validation)
	ErrUnauthorized   = failure.New("token does not belong to caller", failure.Conflict)
)
