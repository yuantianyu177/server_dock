package config

import (
	"os"
)

// Config holds all application configuration.
type Config struct {
	SecretKey            string
	SSHCredentialKey     string
	PublicURL            string
	DatabaseURL          string
	DefaultAdminUsername string
	DefaultAdminPassword string
	Debug                bool
	Port                 string
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		SecretKey:            getEnv("SECRET_KEY", ""),
		SSHCredentialKey:     getEnv("SSH_CREDENTIAL_KEY", ""),
		PublicURL:            getEnv("PUBLIC_URL", ""),
		DatabaseURL:          getEnv("DATABASE_URL", "data/serverdock.db"),
		DefaultAdminUsername: getEnv("DEFAULT_ADMIN_USERNAME", "admin"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "admin123"),
		Debug:                getEnv("DEBUG", "false") == "true",
		Port:                 getEnv("PORT", "8000"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
