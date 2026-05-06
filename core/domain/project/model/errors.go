package model

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidProject = failure.New("invalid project", failure.Validation)

var ErrDuplicateProjectKey = failure.New("duplicate project key", failure.Conflict)
