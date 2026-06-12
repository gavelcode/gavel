package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultSessionCookie  = "gavel_session"
	defaultAPITokenPrefix = "gav_"
	dirPermissions        = 0o755
)

type Config struct {
	Addr string

	DatabaseURL string
	DataDir     string
	SARIFDir    string

	SessionTTL    time.Duration
	SessionCookie string
	SecureCookies bool

	APITokenPrefix string
}

func Load() (*Config, error) {
	dataDir := getenv("GAVEL_DATA_DIR", "./data")
	if err := os.MkdirAll(dataDir, dirPermissions); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	sarifDir := dataDir + "/analyses"
	if err := os.MkdirAll(sarifDir, dirPermissions); err != nil {
		return nil, fmt.Errorf("create sarif dir: %w", err)
	}

	ttlHours, _ := strconv.Atoi(getenv("GAVEL_SESSION_TTL_HOURS", "168"))
	secure, _ := strconv.ParseBool(getenv("GAVEL_SECURE_COOKIES", "false"))

	return &Config{
		Addr:           getenv("GAVEL_ADDR", ":8080"),
		DatabaseURL:    getenv("GAVEL_DATABASE_URL", "postgres://localhost:5432/gavel?sslmode=disable"),
		DataDir:        dataDir,
		SARIFDir:       sarifDir,
		SessionTTL:     time.Duration(ttlHours) * time.Hour,
		SessionCookie:  defaultSessionCookie,
		SecureCookies:  secure,
		APITokenPrefix: defaultAPITokenPrefix,
	}, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
