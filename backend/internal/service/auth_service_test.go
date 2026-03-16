package service

import (
	"serverdock/internal/model"
	"serverdock/internal/pkg"
	"serverdock/internal/repository"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Admin{})
	return db
}

func setupAuthService(t *testing.T) (*AuthService, *gorm.DB) {
	db := setupAuthTestDB(t)
	repo := repository.NewAdminRepo(db)
	svc := NewAuthService(repo, "test-secret-key-for-jwt-signing")
	return svc, db
}

func TestAuthService_EnsureDefaultAdmin(t *testing.T) {
	svc, db := setupAuthService(t)

	if err := svc.EnsureDefaultAdmin("admin", "admin123"); err != nil {
		t.Fatalf("EnsureDefaultAdmin failed: %v", err)
	}

	var count int64
	db.Model(&model.Admin{}).Count(&count)
	if count != 1 {
		t.Fatalf("Expected 1 admin, got %d", count)
	}

	// Calling again should not create a second admin
	svc.EnsureDefaultAdmin("admin", "admin123")
	db.Model(&model.Admin{}).Count(&count)
	if count != 1 {
		t.Fatalf("Expected still 1 admin, got %d", count)
	}
}

func TestAuthService_LoginSuccess(t *testing.T) {
	svc, _ := setupAuthService(t)
	svc.EnsureDefaultAdmin("admin", "admin123")

	token, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if token == "" {
		t.Fatal("Expected non-empty token")
	}
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	svc, _ := setupAuthService(t)
	svc.EnsureDefaultAdmin("admin", "admin123")

	_, err := svc.Login("admin", "wrongpassword")
	if err == nil {
		t.Fatal("Expected error for wrong password")
	}
}

func TestAuthService_LoginNonexistentUser(t *testing.T) {
	svc, _ := setupAuthService(t)

	_, err := svc.Login("nobody", "password")
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	svc, _ := setupAuthService(t)
	svc.EnsureDefaultAdmin("admin", "admin123")

	token, _ := svc.Login("admin", "admin123")

	adminID, username, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if adminID == 0 {
		t.Fatal("Expected non-zero admin ID")
	}
	if username != "admin" {
		t.Fatalf("Expected username 'admin', got %s", username)
	}
}

func TestAuthService_ValidateTokenInvalid(t *testing.T) {
	svc, _ := setupAuthService(t)

	_, _, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}
}

func TestAuthService_ValidateTokenWrongKey(t *testing.T) {
	svc1, _ := setupAuthService(t)
	svc1.EnsureDefaultAdmin("admin", "admin123")
	token, _ := svc1.Login("admin", "admin123")

	// Different secret key
	db := setupAuthTestDB(t)
	repo := repository.NewAdminRepo(db)
	svc2 := NewAuthService(repo, "different-secret-key-here!!!!!")

	_, _, err := svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("Expected error for token signed with different key")
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	svc, _ := setupAuthService(t)
	svc.EnsureDefaultAdmin("admin", "admin123")

	// Get admin ID
	token, _ := svc.Login("admin", "admin123")
	adminID, _, _ := svc.ValidateToken(token)

	if err := svc.ChangePassword(adminID, "admin123", "newpass123"); err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Old password should no longer work
	_, err := svc.Login("admin", "admin123")
	if err == nil {
		t.Fatal("Expected login with old password to fail")
	}

	// New password should work
	_, err = svc.Login("admin", "newpass123")
	if err != nil {
		t.Fatalf("Login with new password failed: %v", err)
	}
}

func TestAuthService_ChangePasswordWrongOld(t *testing.T) {
	svc, _ := setupAuthService(t)
	svc.EnsureDefaultAdmin("admin", "admin123")

	hash, _ := pkg.HashPassword("admin123")
	_ = hash

	token, _ := svc.Login("admin", "admin123")
	adminID, _, _ := svc.ValidateToken(token)

	err := svc.ChangePassword(adminID, "wrongold", "newpass")
	if err == nil {
		t.Fatal("Expected error for wrong old password")
	}
}
