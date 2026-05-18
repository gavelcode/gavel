package gavelconfig

type ServerConfig struct {
	url   string
	token string
}

func (s ServerConfig) URL() string   { return s.url }
func (s ServerConfig) Token() string { return s.token }
func (s ServerConfig) IsZero() bool  { return s.url == "" && s.token == "" }
