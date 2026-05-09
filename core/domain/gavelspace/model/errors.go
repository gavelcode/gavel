package model

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidGavelspace      = failure.New("invalid gavelspace", failure.Validation)
	ErrDuplicateTargetPattern = failure.New("duplicate target pattern", failure.Conflict)
	ErrProjectNotFound        = failure.New("project not found", failure.NotFound)
)
