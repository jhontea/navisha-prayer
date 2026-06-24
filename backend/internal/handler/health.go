package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// GetHealth handles GET /api/v1/health
func (h *HealthHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"status":  "ok",
		"message": "Navisha Prayer API is running",
		"data": gin.H{
			"service": "navisha-prayer-backend",
			"version": "1.0.0",
		},
	})
}
