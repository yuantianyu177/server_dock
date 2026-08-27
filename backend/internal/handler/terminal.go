package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type TerminalHandler struct {
	terminalService *service.TerminalService
	authService     *service.AuthService
}

func NewTerminalHandler(terminalService *service.TerminalService, authService *service.AuthService) *TerminalHandler {
	return &TerminalHandler{terminalService: terminalService, authService: authService}
}

type resizeMessage struct {
	Type string `json:"type"`
	Rows int    `json:"rows"`
	Cols int    `json:"cols"`
}

// validateToken checks the JWT from query param; returns false and writes response if invalid.
func (h *TerminalHandler) validateToken(c *gin.Context) bool {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return false
	}
	_, _, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return false
	}
	return true
}

// ServerTerminal handles WebSocket connection for server terminal.
func (h *TerminalHandler) ServerTerminal(c *gin.Context) {
	if !h.validateToken(c) {
		return
	}

	serverID, ok := parseUintParam(c, "server_id")
	if !ok {
		return
	}

	h.handleTerminal(c, serverID, "")
}

// ContainerTerminal handles WebSocket connection for container terminal.
func (h *TerminalHandler) ContainerTerminal(c *gin.Context) {
	if !h.validateToken(c) {
		return
	}

	serverID, ok := parseUintParam(c, "server_id")
	if !ok {
		return
	}

	containerName := c.Param("container_name")
	if !service.ValidateContainerName(containerName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container name"})
		return
	}

	command := "docker exec -it " + containerName + " /bin/bash"
	h.handleTerminal(c, serverID, command)
}

func (h *TerminalHandler) handleTerminal(c *gin.Context, serverID uint, command string) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	session, err := h.terminalService.CreateSession(serverID, command)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}
	defer session.Close()

	// Mutex to protect concurrent WebSocket writes
	var wsMu sync.Mutex

	// Read from SSH -> send to WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := session.Read(buf)
			if err != nil {
				wsMu.Lock()
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				wsMu.Unlock()
				return
			}
			wsMu.Lock()
			writeErr := conn.WriteMessage(websocket.TextMessage, buf[:n])
			wsMu.Unlock()
			if writeErr != nil {
				return
			}
		}
	}()

	// Read from WebSocket -> send to SSH
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// Check if it's a resize message
		var resize resizeMessage
		if json.Unmarshal(message, &resize) == nil && resize.Type == "resize" {
			session.Resize(resize.Rows, resize.Cols)
			continue
		}

		session.Write(message)
	}
}
