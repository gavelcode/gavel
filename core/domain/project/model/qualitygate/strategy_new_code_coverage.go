package qualitygate

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

type MinNewCodeCoverage struct {
	min float64
}

func NewMinNewCodeCoverage(min float64) (MinNewCodeCoverage, error) {
	if min < 0 || min > 100 {
		return MinNewCodeCoverage{}, fmt.Errorf("%w: min must be between 0 and 100", ErrInvalidStrategy)
	}
	return MinNewCodeCoverage{min: min}, nil
}

func (m MinNewCodeCoverage) Min() float64 { return m.min }

func (m MinNewCodeCoverage) Equal(other Strategy) bool {
	o, ok := other.(MinNewCodeCoverage)
	if !ok {
		return false
	}
	return m.min == o.min
}

func (m MinNewCodeCoverage) Evaluate(content evidence.Content) Outcome {
	ncc, ok := content.(coverage.PatchContent)
	if !ok {
		return NewOutcome(true, "")
	}
	if ncc.CoverableLines() == 0 {
		return NewOutcome(true, "")
	}
	pct := ncc.Percent()
	if pct >= m.min {
		return NewOutcome(true, "")
	}
	return NewOutcome(false, fmt.Sprintf("%.1f%% new code coverage (min %.1f%%)", pct, m.min))
}
