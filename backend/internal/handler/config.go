package handler

import (
	"net/http"

	"serverdock/internal/dto"
	"serverdock/internal/pkg"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService *service.ConfigService
	emailService  service.EmailService
}

func NewConfigHandler(configService *service.ConfigService, emailService service.EmailService) *ConfigHandler {
	return &ConfigHandler{configService: configService, emailService: emailService}
}

func (h *ConfigHandler) List(c *gin.Context) {
	items, err := h.configService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, pkg.SuccessResponse(items))
}

func (h *ConfigHandler) GetAll(c *gin.Context) {
	m, err := h.configService.GetAllAsMap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, pkg.SuccessResponse(m))
}

func (h *ConfigHandler) Update(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "key is required"))
		return
	}

	var req dto.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	if err := h.configService.Set(key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(nil))
}

func (h *ConfigHandler) TestEmail(c *gin.Context) {
	adminEmail := h.configService.Get("admin_email")
	if adminEmail == "" {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "admin_email not configured"))
		return
	}

	err := h.emailService.Send(adminEmail, "ServerDock Test Email", "<h2>Test Email</h2><p>SMTP configuration is working correctly.</p>")
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, "failed to send test email: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(gin.H{"message": "test email sent to " + adminEmail}))
}
