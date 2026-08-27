package handler

import (
	"errors"
	"net/http"

	"serverdock/internal/dto"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type ApplicationHandler struct{ service *service.ApplicationService }

func NewApplicationHandler(applicationService *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{service: applicationService}
}

func (h *ApplicationHandler) PublicListServers(c *gin.Context) {
	servers, err := h.service.ListPublicServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, servers)
}

func (h *ApplicationHandler) PublicListImages(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	images, err := h.service.ListPublicImages(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}

func (h *ApplicationHandler) PublicSubmit(c *gin.Context) {
	var request dto.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	response, err := h.service.Submit(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *ApplicationHandler) List(c *gin.Context) {
	applications, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, applications)
}

func (h *ApplicationHandler) Action(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var request dto.ApplicationActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	var response *dto.ApplicationResponse
	var err error
	switch request.Action {
	case "approve":
		response, err = h.service.Approve(id)
	case "reject":
		response, err = h.service.Reject(id)
	case "ignore":
		response, err = h.service.Ignore(id)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application action"})
		return
	}
	if err == nil {
		c.JSON(http.StatusOK, response)
		return
	}

	status := http.StatusInternalServerError
	if errors.Is(err, service.ErrApplicationNotFound) {
		status = http.StatusNotFound
	} else if errors.Is(err, service.ErrApplicationNotPending) {
		status = http.StatusConflict
	}
	c.JSON(status, gin.H{"error": err.Error()})
}
