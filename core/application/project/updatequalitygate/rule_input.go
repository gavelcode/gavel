package updatequalitygate

type RuleInput struct {
	Subtype     string
	Strategy    StrategyInput
	MinResolved *int
	MinDelta    *float64
}
