package handler

import (
	"net/http"

	"navisha-prayer/internal/service"

	"github.com/gin-gonic/gin"
)

// LocationHandler handles location configuration HTTP requests.
type LocationHandler struct {
	service service.LocationService
}

func NewLocationHandler(service service.LocationService) *LocationHandler {
	return &LocationHandler{service: service}
}

// GetLocation handles GET /api/v1/config/location
func (h *LocationHandler) GetLocation(c *gin.Context) {
	resp, err := h.service.GetLocation(c.Request.Context())
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

// SetLocationGPS handles POST /api/v1/config/location/gps
func (h *LocationHandler) SetLocationGPS(c *gin.Context) {
	var req struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "invalid request body: latitude and longitude are required",
		})
		return
	}

	if req.Latitude < -90 || req.Latitude > 90 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "invalid latitude: must be between -90 and 90",
		})
		return
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "invalid longitude: must be between -180 and 180",
		})
		return
	}

	if err := h.service.SetLocationGPS(c.Request.Context(), req.Latitude, req.Longitude); err != nil {
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
		"message": "location updated successfully",
	})
}

// SetLocationCity handles POST /api/v1/config/location/city
func (h *LocationHandler) SetLocationCity(c *gin.Context) {
	var req struct {
		City    string `json:"city" binding:"required"`
		Country string `json:"country"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "invalid request body: city is required",
		})
		return
	}

	if err := h.service.SetLocationCity(c.Request.Context(), req.City, req.Country); err != nil {
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
		"message": "location updated successfully",
	})
}

// SearchCities handles GET /api/v1/config/location/search
func (h *LocationHandler) SearchCities(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "q query parameter is required",
		})
		return
	}

	resp, err := h.service.SearchCities(c.Request.Context(), query)
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

