package updatequalitygate

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func toDomain(input Input) (qualitygate.Gate, error) {
	rules := make([]qualitygate.Rule, 0, len(input.Rules))
	for i, r := range input.Rules {
		rule, err := toDomainRule(r)
		if err != nil {
			return qualitygate.Gate{}, fmt.Errorf("rules[%d]: %w", i, err)
		}
		rules = append(rules, rule)
	}
	return qualitygate.NewGate(rules)
}

func toDomainRule(input RuleInput) (qualitygate.Rule, error) {
	subtype, err := evidence.NewSubtype(input.Subtype)
	if err != nil {
		return qualitygate.Rule{}, fmt.Errorf("subtype: %w", err)
	}
	strategy, err := toDomainStrategy(input.Strategy)
	if err != nil {
		return qualitygate.Rule{}, fmt.Errorf("strategy: %w", err)
	}
	var opts []qualitygate.RuleOption
	if input.MinResolved != nil {
		opts = append(opts, qualitygate.WithMinResolved(*input.MinResolved))
	}
	if input.MinDelta != nil {
		opts = append(opts, qualitygate.WithMinDelta(*input.MinDelta))
	}
	return qualitygate.NewRule(subtype, strategy, opts...)
}

func toDomainStrategy(input StrategyInput) (qualitygate.Strategy, error) {
	switch input.Type {
	case StrategyTypeCountBySeverity:
		if input.CountBySeverity == nil {
			return nil, fmt.Errorf("%w: %s requires countBySeverity payload", ErrInvalidCommand, input.Type)
		}
		return qualitygate.NewCountBySeverity(
			input.CountBySeverity.MaxError,
			input.CountBySeverity.MaxWarning,
			input.CountBySeverity.MaxNote,
		)
	case StrategyTypeZeroTolerance:
		return qualitygate.NewZeroTolerance(), nil
	case StrategyTypeMinPercentage:
		if input.MinPercentage == nil {
			return nil, fmt.Errorf("%w: %s requires minPercentage payload", ErrInvalidCommand, input.Type)
		}
		return qualitygate.NewMinPercentage(input.MinPercentage.Min)
	case StrategyTypeForbiddenList:
		if input.ForbiddenList == nil {
			return nil, fmt.Errorf("%w: %s requires forbiddenList payload", ErrInvalidCommand, input.Type)
		}
		return qualitygate.NewForbiddenList(input.ForbiddenList.Forbidden)
	case StrategyTypeMaxViolations:
		if input.MaxViolations == nil {
			return nil, fmt.Errorf("%w: %s requires maxViolations payload", ErrInvalidCommand, input.Type)
		}
		return qualitygate.NewMaxViolations(input.MaxViolations.Max)
	case StrategyTypeMinNewCodeCoverage:
		if input.MinNewCodeCoverage == nil {
			return nil, fmt.Errorf("%w: %s requires minNewCodeCoverage payload", ErrInvalidCommand, input.Type)
		}
		return qualitygate.NewMinNewCodeCoverage(input.MinNewCodeCoverage.Min)
	default:
		return nil, fmt.Errorf("%w: unknown strategy type %q", ErrInvalidCommand, input.Type)
	}
}
