package gavelconfig

type qualityGateDTO struct {
	Findings        *findingsRuleDTO        `yaml:"findings,omitempty"`
	Coverage        *coverageRuleDTO        `yaml:"coverage,omitempty"`
	NewCodeCoverage *newCodeCoverageRuleDTO `yaml:"new_code_coverage,omitempty"`
	Violations      *violationsRuleDTO      `yaml:"architecture_violations,omitempty"`
	License         *licenseRuleDTO         `yaml:"license,omitempty"`
}
