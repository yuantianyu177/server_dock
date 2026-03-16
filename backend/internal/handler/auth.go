package handler

import (
	"net/http"

	"serverdock/internal/dto"
	"serverdock/internal/pkg"
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
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: username and password required"))
		return
	}

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, pkg.ErrorResponse(401, "invalid credentials"))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(dto.LoginResponse{Token: token}))
}

func (h *AuthHandler) Me(c *gin.Context) {
	adminID := c.GetUint("admin_id")
	username, _ := c.Get("admin_username")

	c.JSON(http.StatusOK, pkg.SuccessResponse(dto.AdminInfo{
		ID:       adminID,
		Username: username.(string),
	}))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request"))
		return
	}

	adminID := c.GetUint("admin_id")
	if err := h.authService.ChangePassword(adminID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "old password is incorrect" {
			c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, "failed to change password"))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(nil))
}
