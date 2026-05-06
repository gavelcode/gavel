package qualitygate

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

type MinPercentage struct {
	min float64
}

func NewMinPercentage(min float64) (MinPercentage, error) {
	if min < 0 || min > 100 {
		return MinPercentage{}, fmt.Errorf("%w: min must be between 0 and 100", ErrInvalidStrategy)
	}
	return MinPercentage{min: min}, nil
}

func (m MinPercentage) Min() float64 { return m.min }

func (m MinPercentage) Equal(other Strategy) bool {
	o, ok := other.(MinPercentage)
	if !ok {
		return false
	}
	return m.min == o.min
}

func (m MinPercentage) Evaluate(content evidence.Content) Outcome {
	cc, ok := content.(coverage.Content)
	if !ok {
		return NewOutcome(true, "")
	}
	pct := cc.Percent()
	if pct >= m.min {
		return NewOutcome(true, "")
	}
	return NewOutcome(false, fmt.Sprintf("%.1f%% coverage (min %.1f%%)", pct, m.min))
}
