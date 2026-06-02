package model

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidPleading = failure.New("invalid pleading", failure.Validation)

var ErrInvalidStatus = failure.New("invalid pleading status", failure.Validation)

var ErrInvalidTransition = failure.New("invalid pleading status transition", failure.Conflict)
