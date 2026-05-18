package gavelconfig

type violationsRuleDTO struct {
	Max         int  `yaml:"max"`
	MinResolved *int `yaml:"min_resolved,omitempty"`
}
