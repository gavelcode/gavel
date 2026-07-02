package gavelconfig

type projectDTO struct {
	Name            string              `yaml:"name"`
	Pattern         string              `yaml:"pattern"`
	Exclude         []string            `yaml:"exclude,omitempty"`
	Tooling         map[string][]string `yaml:"tooling"`
	Gate            qualityGateDTO      `yaml:"quality_gate,omitempty"`
	CoverageOptions *coverageOptionsDTO `yaml:"coverage_options,omitempty"`
}
