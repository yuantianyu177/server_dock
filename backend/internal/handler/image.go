package handler

import (
	"net/http"
	"strconv"

	"serverdock/internal/dto"
	"serverdock/internal/pkg"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageService *service.ImageService
}

func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{imageService: imageService}
}

func (h *ImageHandler) Create(c *gin.Context) {
	var req dto.CreateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	resp, err := h.imageService.Create(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(resp))
}

func (h *ImageHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	resp, err := h.imageService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, pkg.ErrorResponse(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(resp))
}

func (h *ImageHandler) List(c *gin.Context) {
	var serverID *uint
	if sid := c.Query("server_id"); sid != "" {
		id, err := strconv.ParseUint(sid, 10, 32)
		if err == nil {
			uid := uint(id)
			serverID = &uid
		}
	}

	images, err := h.imageService.List(serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(images))
}

func (h *ImageHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request"))
		return
	}

	resp, err := h.imageService.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, pkg.ErrorResponse(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(resp))
}

func (h *ImageHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.imageService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, pkg.ErrorResponse(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(nil))
}

// ListRemoteImages lists Docker images on a server.
func (h *ImageHandler) ListRemote(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	images, err := h.imageService.ListRemoteImages(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(images))
}

// PullRemoteImage pulls an image on a server.
func (h *ImageHandler) PullRemote(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.PullImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: image is required"))
		return
	}

	output, err := h.imageService.PullRemoteImage(id, req.Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(gin.H{"output": output}))
}

// RemoveRemoteImage removes a Docker image from a server.
func (h *ImageHandler) RemoveRemote(c *gin.Context) {
	serverID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	imageID := c.Param("image_id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "image_id is required"))
		return
	}

	if err := h.imageService.RemoveRemoteImage(serverID, imageID); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(nil))
}
