package middlewares

import (
	"fmt"
	"net/http"

	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/gin-gonic/gin"
)

// BodyLimitMiddleware caps the request body size. Requests whose declared
// Content-Length is over the limit are rejected with 413 up front; chunked
// bodies are capped with http.MaxBytesReader, and utils.HandleErrors maps
// the resulting *http.MaxBytesError to 413 when a handler fails to bind.
func BodyLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		max := int64(config.GetInt("MAX_BODY_BYTES"))
		if max <= 0 {
			max = 1 << 20
		}
		if c.Request.ContentLength > max {
			utils.HandleErrors(c, http.StatusRequestEntityTooLarge,
				fmt.Errorf("request body exceeds %d bytes", max), "")
			c.Abort()
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		}
		c.Next()
	}
}
