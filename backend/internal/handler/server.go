package handler

import (
	"net/http"

	"serverdock/internal/dto"
	"serverdock/internal/pkg"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	serverService *service.ServerService
}

func NewServerHandler(serverService *service.ServerService) *ServerHandler {
	return &ServerHandler{serverService: serverService}
}

func (h *ServerHandler) Create(c *gin.Context) {
	var req dto.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	resp, err := h.serverService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(resp))
}

func (h *ServerHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	resp, err := h.serverService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, pkg.ErrorResponse(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(resp))
}

func (h *ServerHandler) List(c *gin.Context) {
	servers, err := h.serverService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(servers))
}

func (h *ServerHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	resp, err := h.serverService.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, pkg.ErrorResponse(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(resp))
}

func (h *ServerHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.serverService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, pkg.ErrorResponse(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(nil))
}

func (h *ServerHandler) TestConnection(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.serverService.TestConnection(id); err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, "connection failed: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(gin.H{"status": "connected"}))
}

func (h *ServerHandler) TestConnectionDirect(c *gin.Context) {
	var req dto.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	if err := h.serverService.TestConnectionDirect(req.Hostname, req.Port, req.User, req.AuthType, req.Credential); err != nil {
		c.JSON(http.StatusInternalServerError, pkg.ErrorResponse(500, "connection failed: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, pkg.SuccessResponse(gin.H{"status": "connected"}))
}
