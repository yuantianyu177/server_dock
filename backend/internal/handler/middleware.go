package handler

import (
	"net/http"
	"strconv"
	"strings"

	"serverdock/internal/pkg"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, pkg.ErrorResponse(401, "authorization header required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, pkg.ErrorResponse(401, "invalid authorization format"))
			c.Abort()
			return
		}

		adminID, username, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, pkg.ErrorResponse(401, "invalid or expired token"))
			c.Abort()
			return
		}

		c.Set("admin_id", adminID)
		c.Set("admin_username", username)
		c.Next()
	}
}

// parseUintParam parses a uint path parameter. Returns 0 and writes a 400
// response if parsing fails. Callers should return immediately when ok is false.
func parseUintParam(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse(400, "invalid "+name))
		return 0, false
	}
	return uint(id), true
}
