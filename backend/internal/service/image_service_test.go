package service

import (
	"serverdock/internal/dto"
	"serverdock/internal/model"
	"serverdock/internal/repository"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockDockerService struct {
	ListImagesFn      func() ([]dto.RemoteImage, error)
	PullImageFn       func(image string) (string, error)
	RemoveImageFn     func(imageID string) error
	CreateContainerFn func(cmd string) (string, error)
	ExecuteCommandFn  func(command string) (string, error)
	CreateVolumeFn    func(name string) error
}

func (m *MockDockerService) ListImages(hostname string, port int, user, authType, credential string) ([]dto.RemoteImage, error) {
	if m.ListImagesFn != nil {
		return m.ListImagesFn()
	}
	return nil, nil
}

func (m *MockDockerService) PullImage(hostname string, port int, user, authType, credential string, image string) (string, error) {
	if m.PullImageFn != nil {
		return m.PullImageFn(image)
	}
	return "pulled", nil
}

func (m *MockDockerService) RemoveImage(hostname string, port int, user, authType, credential string, imageID string) error {
	if m.RemoveImageFn != nil {
		return m.RemoveImageFn(imageID)
	}
	return nil
}

func (m *MockDockerService) ListContainers(hostname string, port int, user, authType, credential string) ([]map[string]string, error) {
	return nil, nil
}

func (m *MockDockerService) ContainerAction(hostname string, port int, user, authType, credential string, name, action string) error {
	return nil
}

func (m *MockDockerService) GetContainerLogs(hostname string, port int, user, authType, credential string, name string, tail int) (string, error) {
	return "", nil
}

func (m *MockDockerService) CreateContainer(hostname string, port int, user, authType, credential string, cmd string) (string, error) {
	if m.CreateContainerFn != nil {
		return m.CreateContainerFn(cmd)
	}
	return "", nil
}

func (m *MockDockerService) ExecuteCommand(hostname string, port int, user, authType, credential string, command string) (string, error) {
	if m.ExecuteCommandFn != nil {
		return m.ExecuteCommandFn(command)
	}
	return "", nil
}

func (m *MockDockerService) ListVolumes(hostname string, port int, user, authType, credential string) ([]map[string]string, error) {
	return nil, nil
}

func (m *MockDockerService) CreateVolume(hostname string, port int, user, authType, credential string, name string) error {
	if m.CreateVolumeFn != nil {
		return m.CreateVolumeFn(name)
	}
	return nil
}

func (m *MockDockerService) RemoveVolume(hostname string, port int, user, authType, credential string, name string) error {
	return nil
}

func setupImageTestDB2(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Server{}, &model.Image{})
	return db
}

func setupImageService(t *testing.T) (*ImageService, *MockDockerService) {
	db := setupImageTestDB2(t)

	serverRepo := repository.NewServerRepo(db)
	imageRepo := repository.NewImageRepo(db)
	mockSSH := &MockSSHService{}
	mockDocker := &MockDockerService{}

	serverSvc := NewServerService(serverRepo, mockSSH, testEncryptKey)
	imageSvc := NewImageService(imageRepo, serverSvc, mockDocker)

	// Create a test server
	serverSvc.Create(&dto.CreateServerRequest{
		Host: "Test", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "pass",
	})

	return imageSvc, mockDocker
}

func TestImageService_CRUD(t *testing.T) {
	svc, _ := setupImageService(t)

	// Create
	resp, err := svc.Create(&dto.CreateImageRequest{
		ServerID: 1, ImageID: "sha256:abc", Name: "Ubuntu", ImageAddress: "ubuntu:22.04",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.Name != "Ubuntu" {
		t.Fatalf("Expected 'Ubuntu', got %s", resp.Name)
	}

	// Get
	got, err := svc.GetByID(resp.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ImageAddress != "ubuntu:22.04" {
		t.Fatalf("Expected 'ubuntu:22.04', got %s", got.ImageAddress)
	}

	// List
	list, _ := svc.List(nil)
	if len(list) != 1 {
		t.Fatalf("Expected 1, got %d", len(list))
	}

	// Update
	updated, _ := svc.Update(resp.ID, &dto.UpdateImageRequest{Name: "Ubuntu Updated"})
	if updated.Name != "Ubuntu Updated" {
		t.Fatalf("Expected 'Ubuntu Updated', got %s", updated.Name)
	}

	// Delete
	if err := svc.Delete(resp.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestImageService_CreateInvalidServer(t *testing.T) {
	svc, _ := setupImageService(t)

	_, err := svc.Create(&dto.CreateImageRequest{
		ServerID: 999, ImageID: "sha256:abc", Name: "Ubuntu", ImageAddress: "ubuntu:22.04",
	})
	if err == nil {
		t.Fatal("Expected error for non-existent server")
	}
}

func TestImageService_RemoveRemoteImage_BlockedByDBRecord(t *testing.T) {
	svc, _ := setupImageService(t)

	// Create DB record referencing this image
	svc.Create(&dto.CreateImageRequest{
		ServerID: 1, ImageID: "sha256:abc", Name: "Ubuntu", ImageAddress: "ubuntu:22.04",
	})

	err := svc.RemoveRemoteImage(1, "sha256:abc")
	if err == nil {
		t.Fatal("Expected error when removing image referenced by DB record")
	}
}

func TestImageService_RemoveRemoteImage_Allowed(t *testing.T) {
	svc, _ := setupImageService(t)

	// No DB record for this image
	err := svc.RemoveRemoteImage(1, "sha256:xyz")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}
