package repository

import (
	"serverdock/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupImageTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Server{}, &model.Image{})

	// Create a test server
	db.Create(&model.Server{Host: "Test", Hostname: "1.1.1.1", Port: 22, User: "root", AuthType: "password", Credential: "enc"})
	return db
}

func TestImageRepo_CRUD(t *testing.T) {
	db := setupImageTestDB(t)
	repo := NewImageRepo(db)

	// Create
	img := &model.Image{ServerID: 1, DockerImageID: "sha256:abc", Name: "Ubuntu", ImageAddress: "ubuntu:22.04"}
	if err := repo.Create(img); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// FindByID
	found, err := repo.FindByID(img.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "Ubuntu" {
		t.Fatalf("Expected 'Ubuntu', got %s", found.Name)
	}

	// Update
	found.Name = "Ubuntu Updated"
	repo.Update(found)
	updated, _ := repo.FindByID(img.ID)
	if updated.Name != "Ubuntu Updated" {
		t.Fatalf("Expected 'Ubuntu Updated', got %s", updated.Name)
	}

	// List
	images, _ := repo.List(nil)
	if len(images) != 1 {
		t.Fatalf("Expected 1, got %d", len(images))
	}

	// List by server
	sid := uint(1)
	images, _ = repo.List(&sid)
	if len(images) != 1 {
		t.Fatalf("Expected 1, got %d", len(images))
	}

	sid2 := uint(999)
	images, _ = repo.List(&sid2)
	if len(images) != 0 {
		t.Fatalf("Expected 0, got %d", len(images))
	}

	// Delete
	repo.Delete(img.ID)
	_, err = repo.FindByID(img.ID)
	if err == nil {
		t.Fatal("Expected error after delete")
	}
}

func TestImageRepo_FindByImageIDAndServerID(t *testing.T) {
	db := setupImageTestDB(t)
	repo := NewImageRepo(db)

	repo.Create(&model.Image{ServerID: 1, DockerImageID: "sha256:abc", Name: "Ubuntu", ImageAddress: "ubuntu:22.04"})

	found, err := repo.FindByImageIDAndServerID("sha256:abc", 1)
	if err != nil {
		t.Fatalf("FindByImageIDAndServerID failed: %v", err)
	}
	if found.Name != "Ubuntu" {
		t.Fatalf("Expected 'Ubuntu', got %s", found.Name)
	}

	_, err = repo.FindByImageIDAndServerID("sha256:xyz", 1)
	if err == nil {
		t.Fatal("Expected error for non-existent image")
	}
}
