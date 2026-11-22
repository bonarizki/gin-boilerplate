package middlewares

import (
	"strings"

	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware sets baseline HTTP security headers on every response.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth") ||
			strings.HasPrefix(c.Request.URL.Path, "/api/v1/profile") {
			c.Header("Cache-Control", "no-store")
		}
		if config.GetBool("ENABLE_HSTS") {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}