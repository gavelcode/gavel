package gavelconfig

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrReadConfig  = failure.New("read gavel config", failure.Validation)
	ErrParseConfig = failure.New("parse gavel config", failure.Validation)
)
