package qualitygate

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

type Rule struct {
	subtype     evidence.Subtype
	strategy    Strategy
	minResolved *int
	minDelta    *float64
}

type RuleOption func(*Rule) error

func WithMinResolved(n int) RuleOption {
	return func(r *Rule) error {
		if n < 0 {
			return fmt.Errorf("%w: minResolved must not be negative", ErrInvalidRule)
		}
		r.minResolved = &n
		return nil
	}
}

func WithMinDelta(d float64) RuleOption {
	return func(r *Rule) error {
		r.minDelta = &d
		return nil
	}
}

func NewRule(subtype evidence.Subtype, strategy Strategy, opts ...RuleOption) (Rule, error) {
	if strategy == nil {
		return Rule{}, fmt.Errorf("%w: strategy must not be nil", ErrInvalidRule)
	}
	if err := validateCompatibility(subtype, strategy); err != nil {
		return Rule{}, err
	}
	rule := Rule{
		subtype:  subtype,
		strategy: strategy,
	}
	for _, opt := range opts {
		if err := opt(&rule); err != nil {
			return Rule{}, err
		}
	}
	return rule, nil
}

func validateCompatibility(subtype evidence.Subtype, strategy Strategy) error {
	switch strategy.(type) {
	case CountBySeverity, ZeroTolerance:
		if !evidence.IsSubtypeFindingBased(subtype) {
			return fmt.Errorf("%w: strategy requires finding-based subtype", ErrInvalidRule)
		}
	case MinPercentage:
		if subtype != evidence.SubtypeCoverage {
			return fmt.Errorf("%w: MinPercentage requires coverage subtype", ErrInvalidRule)
		}
	case ForbiddenList:
		if subtype != evidence.SubtypeLicense {
			return fmt.Errorf("%w: ForbiddenList requires license subtype", ErrInvalidRule)
		}
	case MaxViolations:
		if subtype != evidence.SubtypeArchitecture {
			return fmt.Errorf("%w: MaxViolations requires architecture subtype", ErrInvalidRule)
		}
	case MinNewCodeCoverage:
		if subtype != evidence.SubtypeNewCodeCoverage {
			return fmt.Errorf("%w: MinNewCodeCoverage requires new_code_coverage subtype", ErrInvalidRule)
		}
	default:
		return fmt.Errorf("%w: unknown strategy type", ErrInvalidRule)
	}
	return nil
}

func (r Rule) Subtype() evidence.Subtype {
	return r.subtype
}

func (r Rule) Strategy() Strategy {
	return r.strategy
}

func (r Rule) MinResolved() *int {
	if r.minResolved == nil {
		return nil
	}
	v := *r.minResolved
	return &v
}

func (r Rule) MinDelta() *float64 {
	if r.minDelta == nil {
		return nil
	}
	v := *r.minDelta
	return &v
}

func (r Rule) Equal(other Rule) bool {
	if r.subtype != other.subtype {
		return false
	}
	if !r.strategy.Equal(other.strategy) {
		return false
	}
	if !ptrEqual(r.minResolved, other.minResolved) {
		return false
	}
	if !ptrEqual(r.minDelta, other.minDelta) {
		return false
	}
	return true
}

func ptrEqual[T comparable](current, other *T) bool {
	if current == nil && other == nil {
		return true
	}
	if current == nil || other == nil {
		return false
	}
	return *current == *other
}
