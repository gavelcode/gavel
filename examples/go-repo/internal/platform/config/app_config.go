package config

// deliberate gochecknoglobals: global mutable variable
var AppName = "ecommerce"
var AppVersion = "1.0.0"
var Debug = false

type AppConfig struct {
	Name    string
	Version string
	Debug   bool
}

func NewAppConfig() AppConfig {
	return AppConfig{
		Name:    AppName,
		Version: AppVersion,
		Debug:   Debug,
	}
}
