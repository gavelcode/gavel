package evidencedto

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrIncompatibleEvidence = failure.New("incompatible evidence dto", failure.Validation)
