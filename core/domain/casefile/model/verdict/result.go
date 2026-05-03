package verdict

import (
	"fmt"
	"time"
)

type Result struct {
	outcome     Outcome
	rulings     []Ruling
	evaluatedAt time.Time
}

func Compose(rulings []Ruling, evaluatedAt time.Time) (Result, error) {
	if evaluatedAt.IsZero() {
		return Result{}, fmt.Errorf("%w: evaluatedAt must not be zero", ErrInvalidVerdict)
	}
	copied := make([]Ruling, len(rulings))
	copy(copied, rulings)

	return Result{
		outcome:     determineOutcome(copied),
		rulings:     copied,
		evaluatedAt: evaluatedAt,
	}, nil
}

func determineOutcome(rulings []Ruling) Outcome {
	for _, r := range rulings {
		if !r.Passed() {
			return OutcomeFail
		}
	}
	return OutcomePass
}

func ReconstituteResult(outcomeStr string, rulings []Ruling, evaluatedAt time.Time) (Result, error) {
	outcome, err := NewOutcome(outcomeStr)
	if err != nil {
		return Result{}, fmt.Errorf("reconstitute verdict: %w", err)
	}
	if evaluatedAt.IsZero() {
		return Result{}, fmt.Errorf("%w: evaluatedAt must not be zero", ErrInvalidVerdict)
	}
	copied := make([]Ruling, len(rulings))
	copy(copied, rulings)
	return Result{
		outcome:     outcome,
		rulings:     copied,
		evaluatedAt: evaluatedAt,
	}, nil
}

func (v Result) Outcome() Outcome { return v.outcome }

func (v Result) Rulings() []Ruling {
	copied := make([]Ruling, len(v.rulings))
	copy(copied, v.rulings)
	return copied
}

func (v Result) EvaluatedAt() time.Time { return v.evaluatedAt }
