package qualitygate

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

type ZeroTolerance struct{}

func NewZeroTolerance() ZeroTolerance {
	return ZeroTolerance{}
}

func (z ZeroTolerance) Equal(other Strategy) bool {
	_, ok := other.(ZeroTolerance)
	return ok
}

func (z ZeroTolerance) Evaluate(content evidence.Content) Outcome {
	fc, ok := content.(finding.Content)
	if !ok {
		return NewOutcome(true, "")
	}
	findings := fc.Findings()
	if len(findings) == 0 {
		return NewOutcome(true, "")
	}
	return NewOutcome(false, fmt.Sprintf("%d findings (max 0)", len(findings)))
}
