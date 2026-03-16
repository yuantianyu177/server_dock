package repository

import (
	"serverdock/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Admin{})
	return db
}

func TestAdminRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAdminRepo(db)

	admin := &model.Admin{Username: "admin", PasswordHash: "hash123"}
	if err := repo.Create(admin); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if admin.ID == 0 {
		t.Fatal("Expected ID to be set after create")
	}
}

func TestAdminRepo_FindByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAdminRepo(db)

	repo.Create(&model.Admin{Username: "admin", PasswordHash: "hash123"})

	found, err := repo.FindByUsername("admin")
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	if found.Username != "admin" {
		t.Fatalf("Expected username 'admin', got %s", found.Username)
	}
}

func TestAdminRepo_FindByUsernameNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAdminRepo(db)

	_, err := repo.FindByUsername("nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}
}

func TestAdminRepo_UpdatePassword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAdminRepo(db)

	admin := &model.Admin{Username: "admin", PasswordHash: "old-hash"}
	repo.Create(admin)

	if err := repo.UpdatePassword(admin.ID, "new-hash"); err != nil {
		t.Fatalf("UpdatePassword failed: %v", err)
	}

	found, _ := repo.FindByUsername("admin")
	if found.PasswordHash != "new-hash" {
		t.Fatalf("Expected new-hash, got %s", found.PasswordHash)
	}
}

func TestAdminRepo_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAdminRepo(db)

	count, _ := repo.Count()
	if count != 0 {
		t.Fatalf("Expected 0, got %d", count)
	}

	repo.Create(&model.Admin{Username: "admin", PasswordHash: "hash"})
	count, _ = repo.Count()
	if count != 1 {
		t.Fatalf("Expected 1, got %d", count)
	}
}
