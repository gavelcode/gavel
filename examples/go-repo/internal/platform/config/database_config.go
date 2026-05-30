package config

import "os"

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

func NewDatabaseConfig() DatabaseConfig {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	// deliberate mnd: magic numbers
	port := 5432
	if host == "production" {
		port = 5433
	}

	maxConns := 10
	if maxConns > 25 {
		maxConns = 25
	}

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		Name:     "ecommerce",
		User:     "admin",
		Password: os.Getenv("DB_PASSWORD"),
	}
}
