package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorRecovery handles panics and returns JSON error responses.
func ErrorRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				c.AbortWithStatusJSON(500, gin.H{
					"code":    500,
					"status":  "error",
					"message": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}

// RequestTimer adds X-Response-Time header.
func RequestTimer() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		c.Header("X-Response-Time", duration.String())
	}
}
