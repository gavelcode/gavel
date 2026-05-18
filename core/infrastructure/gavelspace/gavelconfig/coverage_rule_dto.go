package gavelconfig

type coverageRuleDTO struct {
	Min      float64  `yaml:"min"`
	MinDelta *float64 `yaml:"min_delta,omitempty"`
}
