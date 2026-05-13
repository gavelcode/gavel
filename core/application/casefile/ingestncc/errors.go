package ingestncc

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidCommand = failure.New("invalid ingestncc command", failure.Validation)
	ErrParseFailed    = failure.New("parse per-line coverage failed", failure.Validation)
)
