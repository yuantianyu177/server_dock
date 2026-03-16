package service

import (
	"serverdock/internal/model"
	"serverdock/internal/repository"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupConfigService(t *testing.T) *ConfigService {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.SystemConfig{})
	repo := repository.NewConfigRepo(db)
	return NewConfigService(repo)
}

func TestConfigService_GetWithDefault(t *testing.T) {
	svc := setupConfigService(t)

	val := svc.Get("port_range_start")
	if val != "20000" {
		t.Fatalf("Expected default '20000', got %s", val)
	}
}

func TestConfigService_SetAndGet(t *testing.T) {
	svc := setupConfigService(t)

	svc.Set("port_range_start", "25000")
	val := svc.Get("port_range_start")
	if val != "25000" {
		t.Fatalf("Expected '25000', got %s", val)
	}
}

func TestConfigService_GetInt(t *testing.T) {
	svc := setupConfigService(t)

	n := svc.GetInt("port_range_start")
	if n != 20000 {
		t.Fatalf("Expected 20000, got %d", n)
	}
}

func TestConfigService_EnsureDefaults(t *testing.T) {
	svc := setupConfigService(t)

	if err := svc.EnsureDefaults(); err != nil {
		t.Fatalf("EnsureDefaults failed: %v", err)
	}

	items, _ := svc.List()
	if len(items) < 10 {
		t.Fatalf("Expected at least 10 config items, got %d", len(items))
	}
}

func TestConfigService_GetAllAsMap(t *testing.T) {
	svc := setupConfigService(t)
	svc.EnsureDefaults()

	m, err := svc.GetAllAsMap()
	if err != nil {
		t.Fatalf("GetAllAsMap failed: %v", err)
	}

	if m["port_range_start"] != "20000" {
		t.Fatalf("Expected '20000', got %s", m["port_range_start"])
	}
}
