package service

import (
	"serverdock/internal/dto"
	"serverdock/internal/model"
	"serverdock/internal/repository"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTerminalService(t *testing.T) *TerminalService {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Server{})

	serverRepo := repository.NewServerRepo(db)
	mockSSH := &MockSSHService{}
	serverSvc := NewServerService(serverRepo, mockSSH, testEncryptKey)

	serverSvc.Create(&dto.CreateServerRequest{
		Host: "Test", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "pass",
	})

	return NewTerminalService(serverSvc)
}

func TestTerminalService_SessionManagement(t *testing.T) {
	svc := setupTerminalService(t)

	// CloseSession on non-existent session should not panic
	svc.CloseSession("nonexistent")

	// Verify sessions map is empty
	svc.mu.RLock()
	count := len(svc.sessions)
	svc.mu.RUnlock()

	if count != 0 {
		t.Fatalf("Expected 0 sessions, got %d", count)
	}
}

func TestTerminalService_CreateSessionInvalidServer(t *testing.T) {
	svc := setupTerminalService(t)

	_, err := svc.CreateSession("test-session", 999, "")
	if err == nil {
		t.Fatal("Expected error for non-existent server")
	}
}
