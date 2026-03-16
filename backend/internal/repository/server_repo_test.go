package repository

import (
	"serverdock/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Server{})
	return db
}

func TestServerRepo_CRUD(t *testing.T) {
	db := setupServerTestDB(t)
	repo := NewServerRepo(db)

	// Create
	server := &model.Server{
		Host:       "Test Server",
		Hostname:   "192.168.1.1",
		Port:       22,
		User:       "root",
		AuthType:   "password",
		Credential: "encrypted-data",
	}
	if err := repo.Create(server); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if server.ID == 0 {
		t.Fatal("Expected ID to be set")
	}

	// FindByID
	found, err := repo.FindByID(server.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Host != "Test Server" {
		t.Fatalf("Expected 'Test Server', got %s", found.Host)
	}

	// Update
	found.Host = "Updated Server"
	if err := repo.Update(found); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	updated, _ := repo.FindByID(server.ID)
	if updated.Host != "Updated Server" {
		t.Fatalf("Expected 'Updated Server', got %s", updated.Host)
	}

	// List
	servers, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(servers))
	}

	// Delete
	if err := repo.Delete(server.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = repo.FindByID(server.ID)
	if err == nil {
		t.Fatal("Expected error after delete")
	}
}

func TestServerRepo_FindByIDNotFound(t *testing.T) {
	db := setupServerTestDB(t)
	repo := NewServerRepo(db)

	_, err := repo.FindByID(999)
	if err == nil {
		t.Fatal("Expected error for non-existent server")
	}
}
