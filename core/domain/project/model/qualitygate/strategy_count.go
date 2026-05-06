package qualitygate

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

type CountBySeverity struct {
	maxError   int
	maxWarning int
	maxNote    int
}

func NewCountBySeverity(maxError, maxWarning, maxNote int) (CountBySeverity, error) {
	if maxError < 0 || maxWarning < 0 || maxNote < 0 {
		return CountBySeverity{}, fmt.Errorf("%w: thresholds must be non-negative", ErrInvalidStrategy)
	}
	return CountBySeverity{
		maxError:   maxError,
		maxWarning: maxWarning,
		maxNote:    maxNote,
	}, nil
}

func (c CountBySeverity) Evaluate(content evidence.Content) Outcome {
	fc, ok := content.(finding.Content)
	if !ok {
		return NewOutcome(true, "")
	}
	counts := countBySeverity(fc.Findings())
	return c.checkThresholds(counts)
}

func countBySeverity(findings []finding.Finding) map[finding.Severity]int {
	counts := make(map[finding.Severity]int)
	for _, f := range findings {
		counts[f.Severity()]++
	}
	return counts
}

func (c CountBySeverity) MaxError() int   { return c.maxError }
func (c CountBySeverity) MaxWarning() int { return c.maxWarning }
func (c CountBySeverity) MaxNote() int    { return c.maxNote }

func (c CountBySeverity) Equal(other Strategy) bool {
	o, ok := other.(CountBySeverity)
	if !ok {
		return false
	}
	return c.maxError == o.maxError && c.maxWarning == o.maxWarning && c.maxNote == o.maxNote
}

func (c CountBySeverity) checkThresholds(counts map[finding.Severity]int) Outcome {
	if n := counts[finding.SeverityError]; n > c.maxError {
		return NewOutcome(false, fmt.Sprintf("%d errors (max %d)", n, c.maxError))
	}
	if n := counts[finding.SeverityWarning]; n > c.maxWarning {
		return NewOutcome(false, fmt.Sprintf("%d warnings (max %d)", n, c.maxWarning))
	}
	if n := counts[finding.SeverityNote]; n > c.maxNote {
		return NewOutcome(false, fmt.Sprintf("%d notes (max %d)", n, c.maxNote))
	}
	return NewOutcome(true, "")
}
