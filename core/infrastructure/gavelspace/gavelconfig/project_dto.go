package gavelconfig

type projectDTO struct {
	Name            string              `yaml:"name"`
	Pattern         string              `yaml:"pattern"`
	Tooling         []string            `yaml:"tooling"`
	Gate            qualityGateDTO      `yaml:"quality_gate,omitempty"`
	CoverageOptions *coverageOptionsDTO `yaml:"coverage_options,omitempty"`
}
