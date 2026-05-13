package apperr

import "github.com/usegavel/gavel/core/domain/shared/failure"

type Kind int

const (
	Internal   Kind = Kind(failure.Internal)
	Validation Kind = Kind(failure.Validation)
	NotFound   Kind = Kind(failure.NotFound)
	Conflict   Kind = Kind(failure.Conflict)
)

func Of(err error) Kind {
	return Kind(failure.Of(err))
}
