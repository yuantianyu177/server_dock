package repository

import (
	"serverdock/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupConfigTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.SystemConfig{})
	return db
}

func TestConfigRepo_SetAndGet(t *testing.T) {
	db := setupConfigTestDB(t)
	repo := NewConfigRepo(db)

	if err := repo.Set("port_range_start", "20000"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := repo.Get("port_range_start")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "20000" {
		t.Fatalf("Expected '20000', got %s", val)
	}

	// Upsert
	repo.Set("port_range_start", "25000")
	val, _ = repo.Get("port_range_start")
	if val != "25000" {
		t.Fatalf("Expected '25000', got %s", val)
	}
}

func TestConfigRepo_GetNotFound(t *testing.T) {
	db := setupConfigTestDB(t)
	repo := NewConfigRepo(db)

	_, err := repo.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent key")
	}
}

func TestConfigRepo_GetAll(t *testing.T) {
	db := setupConfigTestDB(t)
	repo := NewConfigRepo(db)

	repo.Set("key1", "val1")
	repo.Set("key2", "val2")

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("Expected 2, got %d", len(all))
	}
}
