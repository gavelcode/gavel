package updatequalitygate

const (
	StrategyTypeCountBySeverity    = "count_by_severity"
	StrategyTypeZeroTolerance      = "zero_tolerance"
	StrategyTypeMinPercentage      = "min_percentage"
	StrategyTypeForbiddenList      = "forbidden_list"
	StrategyTypeMaxViolations      = "max_violations"
	StrategyTypeMinNewCodeCoverage = "min_new_code_coverage"
)

type StrategyInput struct {
	Type               string
	CountBySeverity    *CountBySeverity
	MinPercentage      *MinPercentage
	ForbiddenList      *ForbiddenList
	MaxViolations      *MaxViolations
	MinNewCodeCoverage *MinNewCodeCoverage
}
