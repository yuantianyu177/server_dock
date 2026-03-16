package config

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Clear env vars
	os.Unsetenv("SECRET_KEY")
	os.Unsetenv("SSH_CREDENTIAL_KEY")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("DEFAULT_ADMIN_USERNAME")
	os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
	os.Unsetenv("DEBUG")
	os.Unsetenv("PORT")

	cfg := Load()

	if cfg.Port != "8000" {
		t.Fatalf("Expected default port 8000, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "data/serverdock.db" {
		t.Fatalf("Expected default database URL, got %s", cfg.DatabaseURL)
	}
	if cfg.DefaultAdminUsername != "admin" {
		t.Fatalf("Expected default admin username 'admin', got %s", cfg.DefaultAdminUsername)
	}
	if cfg.DefaultAdminPassword != "admin123" {
		t.Fatalf("Expected default admin password 'admin123', got %s", cfg.DefaultAdminPassword)
	}
	if cfg.Debug != false {
		t.Fatal("Expected debug to be false by default")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("SECRET_KEY", "my-secret")
	os.Setenv("SSH_CREDENTIAL_KEY", "my-ssh-key-32bytes-padded-here!")
	os.Setenv("PORT", "9090")
	os.Setenv("DEBUG", "true")
	defer func() {
		os.Unsetenv("SECRET_KEY")
		os.Unsetenv("SSH_CREDENTIAL_KEY")
		os.Unsetenv("PORT")
		os.Unsetenv("DEBUG")
	}()

	cfg := Load()

	if cfg.SecretKey != "my-secret" {
		t.Fatalf("Expected SECRET_KEY 'my-secret', got %s", cfg.SecretKey)
	}
	if cfg.SSHCredentialKey != "my-ssh-key-32bytes-padded-here!" {
		t.Fatalf("Expected SSH_CREDENTIAL_KEY to be set, got %s", cfg.SSHCredentialKey)
	}
	if cfg.Port != "9090" {
		t.Fatalf("Expected port 9090, got %s", cfg.Port)
	}
	if cfg.Debug != true {
		t.Fatal("Expected debug to be true")
	}
}
