package handler

import (
	"net/http"
	"strconv"

	"serverdock/internal/dto"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct{ service *service.ImageService }

func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{service: imageService}
}

func (h *ImageHandler) Create(c *gin.Context) {
	var request dto.CreateImageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	response, err := h.service.Create(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *ImageHandler) List(c *gin.Context) {
	var serverID *uint
	if value := c.Query("server_id"); value != "" {
		if id, err := strconv.ParseUint(value, 10, 32); err == nil {
			parsed := uint(id)
			serverID = &parsed
		}
	}
	images, err := h.service.List(serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}

func (h *ImageHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ImageHandler) ListRemote(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	images, err := h.service.ListRemoteImages(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}

func (h *ImageHandler) PullRemote(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var request dto.PullImageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: image is required"})
		return
	}
	output, err := h.service.PullRemoteImage(id, request.Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"output": output})
}

func (h *ImageHandler) RemoveRemote(c *gin.Context) {
	serverID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	imageID := c.Param("image_id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image_id is required"})
		return
	}
	if err := h.service.RemoveRemoteImage(serverID, imageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
