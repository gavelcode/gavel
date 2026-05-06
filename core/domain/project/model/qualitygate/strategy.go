package qualitygate

import "github.com/usegavel/gavel/core/domain/casefile/model/evidence"

type Strategy interface {
	Evaluate(content evidence.Content) Outcome

	Equal(other Strategy) bool
}
