package service

import (
	"errors"
	"time"

	"serverdock/internal/model"
	"serverdock/internal/pkg"
	"serverdock/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	adminRepo *repository.AdminRepo
	secretKey string
}

func NewAuthService(adminRepo *repository.AdminRepo, secretKey string) *AuthService {
	return &AuthService{adminRepo: adminRepo, secretKey: secretKey}
}

func (s *AuthService) Login(username, password string) (string, error) {
	admin, err := s.adminRepo.FindByUsername(username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !pkg.CheckPassword(password, admin.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	return s.generateToken(admin.ID, admin.Username)
}

func (s *AuthService) GetAdminByID(id uint) (*model.Admin, error) {
	return s.adminRepo.FindByID(id)
}

func (s *AuthService) ChangePassword(adminID uint, oldPassword, newPassword string) error {
	admin, err := s.adminRepo.FindByID(adminID)
	if err != nil {
		return errors.New("admin not found")
	}

	if !pkg.CheckPassword(oldPassword, admin.PasswordHash) {
		return errors.New("old password is incorrect")
	}

	hash, err := pkg.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.adminRepo.UpdatePassword(adminID, hash)
}

func (s *AuthService) generateToken(adminID uint, username string) (string, error) {
	claims := jwt.MapClaims{
		"admin_id": adminID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *AuthService) ValidateToken(tokenString string) (uint, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return 0, "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, "", errors.New("invalid token")
	}

	adminID := uint(claims["admin_id"].(float64))
	username := claims["username"].(string)
	return adminID, username, nil
}

// EnsureDefaultAdmin creates default admin if none exists.
func (s *AuthService) EnsureDefaultAdmin(username, password string) error {
	count, err := s.adminRepo.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := pkg.HashPassword(password)
	if err != nil {
		return err
	}

	return s.adminRepo.Create(&model.Admin{
		Username:     username,
		PasswordHash: hash,
	})
}
