package archconfig

type configDTO struct {
	Version      int                 `yaml:"version"`
	Module       string              `yaml:"module,omitempty"`
	Layers       map[string][]string `yaml:"layers"`
	Rules        []ruleDTO           `yaml:"rules,omitempty"`
	DetectCycles bool                `yaml:"detect_cycles,omitempty"`
	Generic      *genericDTO         `yaml:"generic,omitempty"`
}
