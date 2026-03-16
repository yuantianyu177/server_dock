package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"serverdock/internal/model"
	"serverdock/internal/pkg"
	"serverdock/internal/repository"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthTest(t *testing.T) (*gin.Engine, *service.AuthService) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Admin{})

	repo := repository.NewAdminRepo(db)
	authSvc := service.NewAuthService(repo, "test-secret-key-for-jwt-signing")
	authSvc.EnsureDefaultAdmin("admin", "admin123")

	authHandler := NewAuthHandler(authSvc)

	r := gin.New()
	api := r.Group("/api")
	auth := api.Group("/auth")
	auth.POST("/login", authHandler.Login)

	protected := auth.Group("")
	protected.Use(AuthMiddleware(authSvc))
	protected.GET("/me", authHandler.Me)
	protected.POST("/change-password", authHandler.ChangePassword)

	return r, authSvc
}

func TestLogin_Success(t *testing.T) {
	r, _ := setupAuthTest(t)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp pkg.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("Expected code 0, got %d", resp.Code)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	r, _ := setupAuthTest(t)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("Expected 401, got %d", w.Code)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	r, _ := setupAuthTest(t)

	body, _ := json.Marshal(map[string]string{"username": "admin"})
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("Expected 400, got %d", w.Code)
	}
}

func getAuthToken(t *testing.T, r *gin.Engine) string {
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	return data["token"].(string)
}

func TestMe_Success(t *testing.T) {
	r, _ := setupAuthTest(t)
	token := getAuthToken(t, r)

	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMe_NoToken(t *testing.T) {
	r, _ := setupAuthTest(t)

	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("Expected 401, got %d", w.Code)
	}
}

func TestMe_InvalidToken(t *testing.T) {
	r, _ := setupAuthTest(t)

	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("Expected 401, got %d", w.Code)
	}
}

func TestChangePassword_Success(t *testing.T) {
	r, _ := setupAuthTest(t)
	token := getAuthToken(t, r)

	body, _ := json.Marshal(map[string]string{"old_password": "admin123", "new_password": "newpass123"})
	req, _ := http.NewRequest("POST", "/api/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChangePassword_WrongOld(t *testing.T) {
	r, _ := setupAuthTest(t)
	token := getAuthToken(t, r)

	body, _ := json.Marshal(map[string]string{"old_password": "wrong", "new_password": "newpass123"})
	req, _ := http.NewRequest("POST", "/api/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("Expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChangePassword_NoAuth(t *testing.T) {
	r, _ := setupAuthTest(t)

	body, _ := json.Marshal(map[string]string{"old_password": "admin123", "new_password": "newpass"})
	req, _ := http.NewRequest("POST", "/api/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("Expected 401, got %d", w.Code)
	}
}
