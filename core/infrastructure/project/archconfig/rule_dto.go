package archconfig

type ruleDTO struct {
	Name   string   `yaml:"name"`
	Source string   `yaml:"source"`
	Deny   []string `yaml:"deny"`
}
