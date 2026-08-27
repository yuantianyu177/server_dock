package handler

import (
	"net/http"

	"serverdock/internal/dto"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: username and password required"})
		return
	}

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{Token: token})
}

func (h *AuthHandler) Me(c *gin.Context) {
	adminID := c.GetUint("admin_id")
	username, _ := c.Get("admin_username")

	c.JSON(http.StatusOK, dto.AdminInfo{
		ID:       adminID,
		Username: username.(string),
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	adminID := c.GetUint("admin_id")
	if err := h.authService.ChangePassword(adminID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "old password is incorrect" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change password"})
		return
	}

	c.Status(http.StatusNoContent)
}
