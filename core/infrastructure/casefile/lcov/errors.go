package lcov

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrScanLCOV    = failure.New("scan lcov", failure.Validation)
	ErrInvalidLine = failure.New("invalid lcov line", failure.Validation)
)
