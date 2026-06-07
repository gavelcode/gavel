package changepassword

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidCommand       = failure.New("invalid change password command", failure.Validation)
	ErrCurrentPasswordWrong = failure.New("current password incorrect", failure.Validation)
)
