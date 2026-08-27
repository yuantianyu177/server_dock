package handler

import (
	"net/http"

	"serverdock/internal/dto"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	resp, err := h.serverService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ServerHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	resp, err := h.serverService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ServerHandler) List(c *gin.Context) {
	servers, err := h.serverService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, servers)
}

func (h *ServerHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	resp, err := h.serverService.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ServerHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.serverService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ServerHandler) TestConnection(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.serverService.TestConnection(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "offline", "error": "connection failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "online"})
}

func (h *ServerHandler) TestConnectionDirect(c *gin.Context) {
	var req dto.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if err := h.serverService.TestConnectionDirect(req.Hostname, req.Port, req.User, req.AuthType, req.Credential); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "offline", "error": "connection failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "online"})
}
