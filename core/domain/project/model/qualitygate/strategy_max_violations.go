package qualitygate

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
)

type MaxViolations struct {
	max int
}

func NewMaxViolations(max int) (MaxViolations, error) {
	if max < 0 {
		return MaxViolations{}, fmt.Errorf("%w: max must be non-negative", ErrInvalidStrategy)
	}
	return MaxViolations{max: max}, nil
}

func (m MaxViolations) Max() int { return m.max }

func (m MaxViolations) Equal(other Strategy) bool {
	o, ok := other.(MaxViolations)
	if !ok {
		return false
	}
	return m.max == o.max
}

func (m MaxViolations) Evaluate(content evidence.Content) Outcome {
	ac, ok := content.(architecture.Content)
	if !ok {
		return NewOutcome(true, "")
	}
	count := len(ac.Violations())
	if count > m.max {
		return NewOutcome(false, fmt.Sprintf("%d violations (max %d)", count, m.max))
	}
	return NewOutcome(true, "")
}
