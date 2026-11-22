package middlewares

import (
	"net/http"

	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/gin-gonic/gin"
)

// MetricsAuthMiddleware protects /metrics when METRICS_TOKEN is set.
func MetricsAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := config.GetString("METRICS_TOKEN")
		if token == "" {
			c.Next()
			return
		}
		if c.GetHeader("Authorization") == "Bearer "+token || c.Query("token") == token {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}