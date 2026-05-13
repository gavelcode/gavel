package ingestevidence

import "github.com/usegavel/gavel/core/domain/shared/failure"

var ErrInvalidCommand = failure.New("invalid ingest evidence command", failure.Validation)
