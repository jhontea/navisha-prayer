package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"navisha-prayer/internal/service"

	"github.com/gin-gonic/gin"
)

// PrayerHandler handles prayer times HTTP requests.
type PrayerHandler struct {
	service service.PrayerService
}

func NewPrayerHandler(service service.PrayerService) *PrayerHandler {
	return &PrayerHandler{service: service}
}

// GetTodayPrayerTimes handles GET /api/v1/prayer-times/today
// Optional query params: latitude, longitude, method
func (h *PrayerHandler) GetTodayPrayerTimes(c *gin.Context) {
	var lat, lon float64
	var method int

	// Parse optional query params for location override
	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")
	methodStr := c.Query("method")

	if latStr != "" {
		v, parseErr := strconv.ParseFloat(latStr, 64)
		if parseErr == nil {
			lat = v
		}
	}
	if lonStr != "" {
		v, parseErr := strconv.ParseFloat(lonStr, 64)
		if parseErr == nil {
			lon = v
		}
	}
	if methodStr != "" {
		v, parseErr := strconv.Atoi(methodStr)
		if parseErr == nil {
			method = v
		}
	}

	var resp *service.PrayerTimesResponse
	var err error

	// If location params provided, use them directly
	if lat != 0 || lon != 0 {
		resp, err = h.service.GetPrayerTimesWithLocation(c.Request.Context(), lat, lon, method)
	} else {
		resp, err = h.service.GetTodayPrayerTimes(c.Request.Context())
	}

	if err != nil {
		var staleErr *service.StaleCacheError
		var noCacheErr *service.NoCacheError
		switch {
		case errors.As(err, &staleErr) && staleErr.Response != nil:
			c.Header("X-Cache-Stale", "true")
			c.JSON(http.StatusOK, gin.H{
				"code":   200,
				"status": "ok",
				"data":   staleErr.Response,
			})
		case errors.As(err, &noCacheErr):
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"status":  "error",
				"message": "Service temporarily unavailable. Please try again later.",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"status":  "error",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"status": "ok",
		"data":   resp,
	})
}

// GetPrayerTimesByDate handles GET /api/v1/prayer-times/date/:date
func (h *PrayerHandler) GetPrayerTimesByDate(c *gin.Context) {
	date := c.Param("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "date parameter is required (DD-MM-YYYY)",
		})
		return
	}

	resp, err := h.service.GetPrayerTimesByDate(c.Request.Context(), date)
	if err != nil {
		var staleErr *service.StaleCacheError
		var noCacheErr *service.NoCacheError
		switch {
		case errors.As(err, &staleErr) && staleErr.Response != nil:
			c.Header("X-Cache-Stale", "true")
			c.JSON(http.StatusOK, gin.H{
				"code":   200,
				"status": "ok",
				"data":   staleErr.Response,
			})
		case errors.As(err, &noCacheErr):
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"status":  "error",
				"message": "Service temporarily unavailable. Please try again later.",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"status":  "error",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"status": "ok",
		"data":   resp,
	})
}

// GetMonthlyPrayerTimes handles GET /api/v1/prayer-times/monthly
func (h *PrayerHandler) GetMonthlyPrayerTimes(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")

	if year == "" || month == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "year and month query parameters are required",
		})
		return
	}

	var y, m int
	if _, err := fmt.Sscanf(year, "%d", &y); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "invalid year",
		})
		return
	}
	if _, err := fmt.Sscanf(month, "%d", &m); err != nil || m < 1 || m > 12 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"status":  "error",
			"message": "invalid month (1-12)",
		})
		return
	}

	resp, err := h.service.GetMonthlyPrayerTimes(c.Request.Context(), y, m)
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

// GetNextPrayer handles GET /api/v1/prayer-times/next
func (h *PrayerHandler) GetNextPrayer(c *gin.Context) {
	nextPrayer, err := h.service.GetNextPrayer(c.Request.Context())
	if err != nil {
		var noCacheErr *service.NoCacheError
		switch {
		case errors.As(err, &noCacheErr):
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"status":  "error",
				"message": "Service temporarily unavailable. Please try again later.",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"status":  "error",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"status": "ok",
		"data":   nextPrayer,
	})
}
