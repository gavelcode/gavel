package model

type ServerConfig struct {
	url   string
	token string
}

func NewServerConfig(url, token string) ServerConfig {
	return ServerConfig{url: url, token: token}
}

func (s ServerConfig) URL() string   { return s.url }
func (s ServerConfig) Token() string { return s.token }
func (s ServerConfig) IsConfigured() bool { return s.url != "" }
