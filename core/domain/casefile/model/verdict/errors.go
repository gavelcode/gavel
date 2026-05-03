package verdict

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidVerdict = failure.New("invalid verdict", failure.Validation)
