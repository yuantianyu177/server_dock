package service

import (
	"serverdock/internal/dto"
	"serverdock/internal/model"
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

	mockSSH := &MockSSHService{}
	serverSvc := NewServerService(db, mockSSH.TestConnection, mockSSH.ExecuteCommand, testEncryptKey)

	serverSvc.Create(&dto.CreateServerRequest{
		Host: "Test", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "pass",
	})

	return NewTerminalService(serverSvc)
}

func TestTerminalService_CreateSessionInvalidServer(t *testing.T) {
	svc := setupTerminalService(t)

	_, err := svc.CreateSession(999, "")
	if err == nil {
		t.Fatal("Expected error for non-existent server")
	}
}
