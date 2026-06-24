package handler

import (
	"net/http"

	"navisha-prayer/internal/service"

	"github.com/gin-gonic/gin"
)

// EidHandler handles Eid countdown HTTP requests.
type EidHandler struct {
	service service.EidService
}

func NewEidHandler(service service.EidService) *EidHandler {
	return &EidHandler{service: service}
}

// GetEidCountdown handles GET /api/v1/eid/countdown
func (h *EidHandler) GetEidCountdown(c *gin.Context) {
	resp, err := h.service.GetEidAlFitrCountdown(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"status": "ok",
		"data":   resp,
	})
}
