package middlewares

import (
	"net/http"

	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/gin-gonic/gin"
)

// BodyLimitMiddleware caps the request body size using http.MaxBytesReader.
func BodyLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		max := config.GetInt("MAX_BODY_BYTES")
		if max <= 0 {
			max = 1 << 20
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(max))
		}
		c.Next()
	}
}