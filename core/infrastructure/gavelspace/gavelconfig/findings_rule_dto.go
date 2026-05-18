package gavelconfig

type findingsRuleDTO struct {
	MaxError    *int `yaml:"max_error,omitempty"`
	MaxWarning  *int `yaml:"max_warning,omitempty"`
	MaxNote     *int `yaml:"max_note,omitempty"`
	MinResolved *int `yaml:"min_resolved,omitempty"`
}
