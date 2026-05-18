package archconfig

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrReadConfig  = failure.New("read architecture config", failure.Validation)
	ErrParseConfig = failure.New("parse architecture config", failure.Validation)
)
