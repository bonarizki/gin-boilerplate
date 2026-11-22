package middlewares

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestTimeoutMiddleware puts a deadline on the request context. Handlers
// and the database queries they run through the repositories see
// context.DeadlineExceeded once d has passed. d <= 0 disables it.
func RequestTimeoutMiddleware(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d <= 0 {
			c.Next()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
