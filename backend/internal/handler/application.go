package handler

import (
	"errors"
	"net/http"

	"serverdock/internal/dto"
	"serverdock/internal/pkg"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type ApplicationHandler struct {
	appService *service.ApplicationService
}

func NewApplicationHandler(appService *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appService: appService}
}

// Public endpoints

func (h *ApplicationHandler) PublicListServers(c *gin.Context) {
	servers, err := h.appService.ListPublicServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, pkg.SuccessResponse(servers))
}

func (h *ApplicationHandler) PublicListImages(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	images, err := h.appService.ListPublicImages(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, pkg.SuccessResponse(images))
}

func (h *ApplicationHandler) PublicSubmit(c *gin.Context) {
	var req dto.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	resp, err := h.appService.Submit(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, pkg.SuccessResponse(resp))
}

// Admin endpoints

func (h *ApplicationHandler) List(c *gin.Context) {
	status := c.Query("status")
	apps, err := h.appService.List(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, pkg.SuccessResponse(apps))
}

func (h *ApplicationHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	resp, err := h.appService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, pkg.ErrorResponse(404, err.Error()))
		return
	}
	c.JSON(http.StatusOK, pkg.SuccessResponse(resp))
}

func (h *ApplicationHandler) Action(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.ApplicationActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	var resp *dto.ApplicationResponse
	var err error
	switch req.Action {
	case "approve":
		resp, err = h.appService.Approve(id, req.AdminNotes)
	case "reject":
		resp, err = h.appService.Reject(id, req.AdminNotes)
	}

	if err != nil {
		switch {
		case errors.Is(err, service.ErrApplicationNotFound):
			c.JSON(http.StatusNotFound, pkg.ErrorResponse(404, err.Error()))
		case errors.Is(err, service.ErrApplicationNotPending):
			c.JSON(http.StatusConflict, pkg.ErrorResponse(409, err.Error()))
		case errors.Is(err, service.ErrContainerProvisioning):
			c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		}
		return
	}
	c.JSON(http.StatusOK, pkg.SuccessResponse(resp))
}
