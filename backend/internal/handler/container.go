package handler

import (
	"net/http"
	"strconv"

	"serverdock/internal/dto"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type ContainerHandler struct{ service *service.ContainerService }

func NewContainerHandler(containerService *service.ContainerService) *ContainerHandler {
	return &ContainerHandler{service: containerService}
}

func (h *ContainerHandler) List(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	containers, err := h.service.ListContainers(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

func (h *ContainerHandler) Create(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var request dto.CreateContainerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	result, err := h.service.CreateContainer(id, request.Name, request.Image, request.ExtraArgs, 20000, 30000, 5, "/data")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ContainerHandler) Action(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var request dto.ContainerActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if err := h.service.ContainerAction(id, c.Param("name"), request.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ContainerHandler) Logs(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	tail, _ := strconv.Atoi(c.DefaultQuery("tail", "100"))
	logs, err := h.service.GetContainerLogs(id, c.Param("name"), tail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

type VolumeHandler struct{ service *service.ContainerService }

func NewVolumeHandler(containerService *service.ContainerService) *VolumeHandler {
	return &VolumeHandler{service: containerService}
}

func (h *VolumeHandler) List(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	volumes, err := h.service.ListVolumes(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, volumes)
}

func (h *VolumeHandler) Create(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var request dto.CreateVolumeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if err := h.service.CreateVolumeSingle(id, request.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *VolumeHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.RemoveVolume(id, c.Param("name")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
