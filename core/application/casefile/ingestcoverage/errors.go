package ingestcoverage

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidCommand = failure.New("invalid ingest coverage command", failure.Validation)
	ErrUnknownFormat  = failure.New("unknown coverage format", failure.Validation)
	ErrParseFailed    = failure.New("coverage parse failed", failure.Validation)
)
