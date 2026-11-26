package gavelconfig

import "gopkg.in/yaml.v3"

type configDTO struct {
	Name     string       `yaml:"name"`
	Projects []projectDTO `yaml:"projects"`
	Server   serverDTO    `yaml:"server,omitempty"`
	// RemovedFindingsSource only detects the removed findings_source key so
	// Parse can fail loudly instead of silently ignoring it.
	RemovedFindingsSource yaml.Node `yaml:"findings_source,omitempty"`
	BazelConfig           []string  `yaml:"bazel_config,omitempty"`
}
