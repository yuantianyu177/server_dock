package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"serverdock/internal/model"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testEncryptKey = "0123456789abcdef0123456789abcdef"

type mockSSH struct{ testError error }

func (m *mockSSH) TestConnection(hostname string, port int, user, authType, credential string) error {
	return m.testError
}
func (m *mockSSH) ExecuteCommand(hostname string, port int, user, authType, credential string, command string) (string, error) {
	return "", nil
}

func setupServerHandlerTest(t *testing.T) (*gin.Engine, string) {
	return setupServerHandlerTestWithConnectionError(t, nil)
}

func setupServerHandlerTestWithConnectionError(t *testing.T, testError error) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Admin{}, &model.Server{})

	authSvc := service.NewAuthService(db, "test-secret-key-for-jwt-signing")
	authSvc.EnsureDefaultAdmin("admin", "admin123")

	ssh := &mockSSH{testError: testError}
	serverSvc := service.NewServerService(db, ssh.TestConnection, ssh.ExecuteCommand, testEncryptKey)
	serverHandler := NewServerHandler(serverSvc)
	authHandler := NewAuthHandler(authSvc)

	r := gin.New()
	api := r.Group("/api")
	api.POST("/auth/login", authHandler.Login)

	servers := api.Group("/servers")
	servers.Use(AuthMiddleware(authSvc))
	{
		servers.GET("", serverHandler.List)
		servers.POST("", serverHandler.Create)
		servers.GET("/:id", serverHandler.Get)
		servers.PUT("/:id", serverHandler.Update)
		servers.DELETE("/:id", serverHandler.Delete)
		servers.POST("/:id/test", serverHandler.TestConnection)
		servers.POST("/test-direct", serverHandler.TestConnectionDirect)
	}

	// Get token
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	token := resp["token"].(string)

	return r, token
}

func TestServerHandler_TestConnectionOffline(t *testing.T) {
	r, token := setupServerHandlerTestWithConnectionError(t, errors.New("connection refused"))
	body, _ := json.Marshal(map[string]interface{}{
		"hostname": "1.1.1.1", "port": 22, "user": "root",
		"auth_type": "password", "credential": "pass123",
	})
	req, _ := http.NewRequest("POST", "/api/servers/test-direct", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["status"] != "offline" {
		t.Fatalf("expected offline status, got %q", response["status"])
	}
}

func TestServerHandler_CRUD(t *testing.T) {
	r, token := setupServerHandlerTest(t)

	// Create
	body, _ := json.Marshal(map[string]interface{}{
		"host": "Test Server", "hostname": "1.1.1.1", "port": 22,
		"user": "root", "auth_type": "password", "credential": "pass123",
	})
	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Create: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// List
	req, _ = http.NewRequest("GET", "/api/servers", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("List: expected 200, got %d", w.Code)
	}

	var listResp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if len(listResp) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(listResp))
	}

	// Get
	req, _ = http.NewRequest("GET", "/api/servers/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Get: expected 200, got %d", w.Code)
	}

	// Update
	body, _ = json.Marshal(map[string]interface{}{"host": "Updated Server"})
	req, _ = http.NewRequest("PUT", "/api/servers/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Update: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Test connection
	req, _ = http.NewRequest("POST", "/api/servers/1/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Test: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var testResponse map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &testResponse); err != nil {
		t.Fatalf("Test: decode response: %v", err)
	}
	if testResponse["status"] != "online" {
		t.Fatalf("Test: expected online status, got %q", testResponse["status"])
	}

	// Delete
	req, _ = http.NewRequest("DELETE", "/api/servers/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Delete: expected 204, got %d", w.Code)
	}
}

func TestServerHandler_GetNotFound(t *testing.T) {
	r, token := setupServerHandlerTest(t)

	req, _ := http.NewRequest("GET", "/api/servers/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestServerHandler_NoAuth(t *testing.T) {
	r, _ := setupServerHandlerTest(t)

	req, _ := http.NewRequest("GET", "/api/servers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("Expected 401, got %d", w.Code)
	}
}

func TestServerHandler_CreateInvalidRequest(t *testing.T) {
	r, token := setupServerHandlerTest(t)

	body, _ := json.Marshal(map[string]string{"host": "test"}) // missing required fields
	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("Expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
