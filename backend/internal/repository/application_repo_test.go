package repository

import (
	"serverdock/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAppTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Server{}, &model.Image{}, &model.Application{})

	db.Create(&model.Server{Host: "Test", Hostname: "1.1.1.1", Port: 22, User: "root", AuthType: "password", Credential: "enc"})
	db.Create(&model.Image{ServerID: 1, DockerImageID: "sha256:abc", Name: "Ubuntu", ImageAddress: "ubuntu:22.04"})
	return db
}

func TestApplicationRepo_CRUD(t *testing.T) {
	db := setupAppTestDB(t)
	repo := NewApplicationRepo(db)

	app := &model.Application{
		ApplicantName:  "Zhang San",
		ApplicantEmail: "zhang@example.com",
		ServerID:       1,
		ImageID:        1,
		Status:         "pending",
	}
	if err := repo.Create(app); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(app.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.ApplicantName != "Zhang San" {
		t.Fatalf("Expected 'Zhang San', got %s", found.ApplicantName)
	}
	if found.Server.Host != "Test" {
		t.Fatalf("Expected preloaded server 'Test', got %s", found.Server.Host)
	}

	// List
	apps, _ := repo.List("")
	if len(apps) != 1 {
		t.Fatalf("Expected 1, got %d", len(apps))
	}

	// List by status
	apps, _ = repo.List("pending")
	if len(apps) != 1 {
		t.Fatalf("Expected 1 pending, got %d", len(apps))
	}
	apps, _ = repo.List("approved")
	if len(apps) != 0 {
		t.Fatalf("Expected 0 approved, got %d", len(apps))
	}

	// Update
	found.Status = "approved"
	repo.Update(found)
	updated, _ := repo.FindByID(app.ID)
	if updated.Status != "approved" {
		t.Fatalf("Expected 'approved', got %s", updated.Status)
	}
}
