package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a CORS middleware configuration.
//
// In addition to the explicitly configured origins, any localhost / 127.0.0.1
// origin is allowed. This avoids "Failed to fetch" errors during development
// when the Next.js dev server starts on an alternate port (e.g. 3001 when 3000
// is already in use).
func CORS(allowOrigins string) gin.HandlerFunc {
	allowed := make(map[string]bool)
	for _, o := range strings.Split(allowOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			allowed[o] = true
		}
	}

	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	config.MaxAge = 86400
	config.AllowOriginFunc = func(origin string) bool {
		if allowed[origin] {
			return true
		}
		// Allow any localhost / 127.0.0.1 origin regardless of port (dev convenience).
		return strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:")
	}

	return cors.New(config)
}
