package service

import (
	"errors"
	"time"

	"serverdock/internal/model"
	"serverdock/internal/pkg"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService struct {
	db        *gorm.DB
	secretKey string
}

func NewAuthService(db *gorm.DB, secretKey string) *AuthService {
	return &AuthService{db: db, secretKey: secretKey}
}

func (s *AuthService) Login(username, password string) (string, error) {
	var admin model.Admin
	if err := s.db.Where("username = ?", username).First(&admin).Error; err != nil || !pkg.CheckPassword(password, admin.PasswordHash) {
		return "", errors.New("invalid credentials")
	}
	return s.generateToken(admin.ID, admin.Username)
}

func (s *AuthService) ChangePassword(adminID uint, oldPassword, newPassword string) error {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return errors.New("admin not found")
	}
	if !pkg.CheckPassword(oldPassword, admin.PasswordHash) {
		return errors.New("old password is incorrect")
	}
	hash, err := pkg.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.db.Model(&model.Admin{}).Where("id = ?", adminID).Update("password_hash", hash).Error
}

func (s *AuthService) generateToken(adminID uint, username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": adminID, "username": username, "exp": time.Now().Add(24 * time.Hour).Unix(),
	})
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
	adminID, idOK := claims["admin_id"].(float64)
	username, nameOK := claims["username"].(string)
	if !idOK || !nameOK {
		return 0, "", errors.New("invalid token claims")
	}
	return uint(adminID), username, nil
}

func (s *AuthService) EnsureDefaultAdmin(username, password string) error {
	var count int64
	if err := s.db.Model(&model.Admin{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	hash, err := pkg.HashPassword(password)
	if err != nil {
		return err
	}
	return s.db.Create(&model.Admin{Username: username, PasswordHash: hash}).Error
}
