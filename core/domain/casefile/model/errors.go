package model

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidCaseFile = failure.New("invalid case file", failure.Validation)
	ErrAlreadyJudged   = failure.New("case file already judged", failure.Conflict)
)
