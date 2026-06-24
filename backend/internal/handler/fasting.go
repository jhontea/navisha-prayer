package handler

import (
	"net/http"

	"navisha-prayer/internal/service"

	"github.com/gin-gonic/gin"
)

// FastingHandler handles fasting schedule HTTP requests.
type FastingHandler struct {
	service service.FastingService
}

func NewFastingHandler(service service.FastingService) *FastingHandler {
	return &FastingHandler{service: service}
}

// GetFastingStatus handles GET /api/v1/fasting/status
func (h *FastingHandler) GetFastingStatus(c *gin.Context) {
	resp, err := h.service.GetCurrentFasting(c.Request.Context())
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

// UpdateFastingType handles POST /api/v1/fasting/type
func (h *FastingHandler) UpdateFastingType(c *gin.Context) {
	var req struct {
		Type string `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "invalid request body: type is required",
		})
		return
	}

	if err := h.service.UpdateFastingType(c.Request.Context(), req.Type); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"status": "ok",
		"message": "fasting type updated successfully",
	})
}

// GetFastingSchedule handles GET /api/v1/fasting/schedule
func (h *FastingHandler) GetFastingSchedule(c *gin.Context) {
	// Reuse status endpoint which includes schedule times
	resp, err := h.service.GetCurrentFasting(c.Request.Context())
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
