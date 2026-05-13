package ingestfindings

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidCommand = failure.New("invalid ingest findings command", failure.Validation)
	ErrUnknownFormat  = failure.New("unknown findings format", failure.Validation)
	ErrParseFailed    = failure.New("findings parse failed", failure.Validation)
)
