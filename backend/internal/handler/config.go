package handler

import (
	"net/http"

	"serverdock/internal/dto"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	service   *service.ConfigService
	sendEmail func(to, subject, body string) error
}

func NewConfigHandler(configService *service.ConfigService, sendEmail func(string, string, string) error) *ConfigHandler {
	return &ConfigHandler{service: configService, sendEmail: sendEmail}
}

func (h *ConfigHandler) GetAll(c *gin.Context) {
	config, err := h.service.GetAllAsMap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *ConfigHandler) Update(c *gin.Context) {
	var request dto.UpdateConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if err := h.service.Set(c.Param("key"), request.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ConfigHandler) TestEmail(c *gin.Context) {
	to := h.service.Get("admin_email")
	if to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin_email not configured"})
		return
	}
	if err := h.sendEmail(to, "ServerDock Test Email", "<h2>Test Email</h2><p>SMTP configuration is working correctly.</p>"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test email: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "test email sent to " + to})
}
