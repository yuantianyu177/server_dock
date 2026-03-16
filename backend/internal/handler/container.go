package handler

import (
	"net/http"
	"strconv"

	"serverdock/internal/dto"
	"serverdock/internal/pkg"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type ContainerHandler struct {
	containerService *service.ContainerService
}

func NewContainerHandler(containerService *service.ContainerService) *ContainerHandler {
	return &ContainerHandler{containerService: containerService}
}

func (h *ContainerHandler) List(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	containers, err := h.containerService.ListContainers(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(containers))
}

func (h *ContainerHandler) Create(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.CreateContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	// Use default config values; these will be replaced by dynamic config in Phase 6
	result, err := h.containerService.CreateContainer(id, req.Name, req.Image, req.ExtraArgs, 20000, 30000, 5, "/data")
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(result))
}

func (h *ContainerHandler) Action(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	name := c.Param("name")
	var req dto.ContainerActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	if err := h.containerService.ContainerAction(id, name, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(nil))
}

func (h *ContainerHandler) Logs(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	name := c.Param("name")
	tail, _ := strconv.Atoi(c.DefaultQuery("tail", "100"))

	logs, err := h.containerService.GetContainerLogs(id, name, tail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(gin.H{"logs": logs}))
}

func (h *ContainerHandler) Exec(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.ExecCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	output, err := h.containerService.ExecCommand(id, req.Command)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(gin.H{"output": output}))
}

type VolumeHandler struct {
	containerService *service.ContainerService
}

func NewVolumeHandler(containerService *service.ContainerService) *VolumeHandler {
	return &VolumeHandler{containerService: containerService}
}

func (h *VolumeHandler) List(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	volumes, err := h.containerService.ListVolumes(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(volumes))
}

func (h *VolumeHandler) Create(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	if err := h.containerService.CreateVolumeSingle(id, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(nil))
}

func (h *VolumeHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	name := c.Param("name")
	if err := h.containerService.RemoveVolume(id, name); err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(nil))
}
