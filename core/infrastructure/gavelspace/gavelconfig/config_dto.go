package gavelconfig

type configDTO struct {
	Name           string       `yaml:"name"`
	Projects       []projectDTO `yaml:"projects"`
	Server         serverDTO    `yaml:"server,omitempty"`
	FindingsSource string       `yaml:"findings_source,omitempty"`
}
