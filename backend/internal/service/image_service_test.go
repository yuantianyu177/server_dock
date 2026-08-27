package service

import (
	"testing"

	"serverdock/internal/dto"
	"serverdock/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupImageService(t *testing.T) (*ImageService, *MockSSHService) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&model.Server{}, &model.Image{})
	mockSSH := &MockSSHService{}
	servers := NewServerService(db, mockSSH.TestConnection, mockSSH.ExecuteCommand, testEncryptKey)
	servers.Create(&dto.CreateServerRequest{
		Host: "Test", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "pass",
	})
	return NewImageService(db, servers), mockSSH
}

func TestImageService_CRUD(t *testing.T) {
	service, _ := setupImageService(t)
	created, err := service.Create(&dto.CreateImageRequest{
		ServerID: 1, ImageID: "sha256:abc", Name: "Ubuntu", ImageAddress: "ubuntu:22.04",
	})
	if err != nil || created.Name != "Ubuntu" {
		t.Fatalf("Create failed: %#v, %v", created, err)
	}
	images, err := service.List(nil)
	if err != nil || len(images) != 1 {
		t.Fatalf("List failed: %#v, %v", images, err)
	}
	if err := service.Delete(created.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestImageService_CreateInvalidServer(t *testing.T) {
	service, _ := setupImageService(t)
	_, err := service.Create(&dto.CreateImageRequest{
		ServerID: 999, ImageID: "sha256:abc", Name: "Ubuntu", ImageAddress: "ubuntu:22.04",
	})
	if err == nil {
		t.Fatal("expected error for non-existent server")
	}
}

func TestImageService_RemoveRemoteImage(t *testing.T) {
	service, _ := setupImageService(t)
	service.Create(&dto.CreateImageRequest{
		ServerID: 1, ImageID: "sha256:abc", Name: "Ubuntu", ImageAddress: "ubuntu:22.04",
	})
	if err := service.RemoveRemoteImage(1, "sha256:abc"); err == nil {
		t.Fatal("expected registered image removal to be blocked")
	}
	if err := service.RemoveRemoteImage(1, "sha256:xyz"); err != nil {
		t.Fatalf("expected unregistered image removal to succeed: %v", err)
	}
}
