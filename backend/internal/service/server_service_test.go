package service

import (
	"errors"
	"serverdock/internal/dto"
	"serverdock/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockSSHService for testing
type MockSSHService struct {
	TestConnectionFn func(hostname string, port int, user, authType, credential string) error
	ExecuteCommandFn func(hostname string, port int, user, authType, credential string, command string) (string, error)
}

func (m *MockSSHService) TestConnection(hostname string, port int, user, authType, credential string) error {
	if m.TestConnectionFn != nil {
		return m.TestConnectionFn(hostname, port, user, authType, credential)
	}
	return nil
}

func (m *MockSSHService) ExecuteCommand(hostname string, port int, user, authType, credential string, command string) (string, error) {
	if m.ExecuteCommandFn != nil {
		return m.ExecuteCommandFn(hostname, port, user, authType, credential, command)
	}
	return "", nil
}

const testEncryptKey = "0123456789abcdef0123456789abcdef"

func setupServerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Server{})
	return db
}

func setupServerService(t *testing.T) (*ServerService, *MockSSHService) {
	db := setupServerTestDB(t)
	mockSSH := &MockSSHService{}
	svc := NewServerService(db, mockSSH.TestConnection, mockSSH.ExecuteCommand, testEncryptKey)
	return svc, mockSSH
}

func TestServerService_CreateAndGet(t *testing.T) {
	svc, _ := setupServerService(t)

	resp, err := svc.Create(&dto.CreateServerRequest{
		Host:       "Test Server",
		Hostname:   "192.168.1.1",
		Port:       22,
		User:       "root",
		AuthType:   "password",
		Credential: "mypassword",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.Host != "Test Server" {
		t.Fatalf("Expected 'Test Server', got %s", resp.Host)
	}

	// Verify credential is encrypted at rest (not plaintext)
	raw, err := svc.find(resp.ID)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if raw.Credential == "mypassword" {
		t.Fatal("Credential should be encrypted")
	}
}

func TestServerService_CreateDefaultPort(t *testing.T) {
	svc, _ := setupServerService(t)

	server, _ := svc.Create(&dto.CreateServerRequest{
		Host:       "Test",
		Hostname:   "1.2.3.4",
		User:       "root",
		AuthType:   "password",
		Credential: "pass",
	})

	if server.Port != 22 {
		t.Fatalf("Expected default port 22, got %d", server.Port)
	}
}

func TestServerService_List(t *testing.T) {
	svc, _ := setupServerService(t)

	svc.Create(&dto.CreateServerRequest{Host: "S1", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "p"})
	svc.Create(&dto.CreateServerRequest{Host: "S2", Hostname: "2.2.2.2", User: "root", AuthType: "password", Credential: "p"})

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("Expected 2 servers, got %d", len(list))
	}
}

func TestServerService_Update(t *testing.T) {
	svc, _ := setupServerService(t)

	server, _ := svc.Create(&dto.CreateServerRequest{
		Host: "Old", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "p",
	})

	resp, err := svc.Update(server.ID, &dto.UpdateServerRequest{Host: "New"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Host != "New" {
		t.Fatalf("Expected 'New', got %s", resp.Host)
	}
}

func TestServerService_UpdateNotFound(t *testing.T) {
	svc, _ := setupServerService(t)
	_, err := svc.Update(999, &dto.UpdateServerRequest{Host: "New"})
	if err == nil {
		t.Fatal("Expected error for non-existent server")
	}
}

func TestServerService_Delete(t *testing.T) {
	svc, _ := setupServerService(t)

	server, _ := svc.Create(&dto.CreateServerRequest{
		Host: "S1", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "p",
	})

	if err := svc.Delete(server.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := svc.GetByID(server.ID)
	if err == nil {
		t.Fatal("Expected error after delete")
	}
}

func TestServerService_TestConnection_Success(t *testing.T) {
	svc, mockSSH := setupServerService(t)
	mockSSH.TestConnectionFn = func(hostname string, port int, user, authType, credential string) error {
		return nil
	}

	server, _ := svc.Create(&dto.CreateServerRequest{
		Host: "S1", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "mypass",
	})

	if err := svc.TestConnection(server.ID); err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}
}

func TestServerService_TestConnection_Fail(t *testing.T) {
	svc, mockSSH := setupServerService(t)
	mockSSH.TestConnectionFn = func(hostname string, port int, user, authType, credential string) error {
		return errors.New("connection refused")
	}

	server, _ := svc.Create(&dto.CreateServerRequest{
		Host: "S1", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "mypass",
	})

	err := svc.TestConnection(server.ID)
	if err == nil {
		t.Fatal("Expected error for failed connection")
	}
}

func TestServerService_TestConnection_DecryptsCredential(t *testing.T) {
	svc, mockSSH := setupServerService(t)

	var receivedCred string
	mockSSH.TestConnectionFn = func(hostname string, port int, user, authType, credential string) error {
		receivedCred = credential
		return nil
	}

	svc.Create(&dto.CreateServerRequest{
		Host: "S1", Hostname: "1.1.1.1", User: "root", AuthType: "password", Credential: "secret-pass",
	})

	svc.TestConnection(1)

	if receivedCred != "secret-pass" {
		t.Fatalf("Expected decrypted credential 'secret-pass', got %q", receivedCred)
	}
}
