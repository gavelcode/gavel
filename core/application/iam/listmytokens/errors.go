package listmytokens

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidQuery = failure.New("invalid list my tokens query", failure.Validation)
