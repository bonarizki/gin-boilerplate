package middlewares

import (
	"context"
	"regexp"

	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var clientRequestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func resolveRequestID(header string) string {
	if header != "" && clientRequestIDPattern.MatchString(header) {
		return header
	}
	return uuid.New().String()
}

// RequestIDMiddleware adds a unique request ID to each request.
//
// The request ID is:
// - Read from header X-Request-ID when it matches ^[A-Za-z0-9._-]{1,64}$; otherwise a new UUID is generated
// - Set in gin context as "request_id"
// - Injected into c.Request.Context() for use in services
// - Added to response header as "X-Request-ID"
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := resolveRequestID(c.GetHeader("X-Request-ID"))

		c.Set("request_id", requestID)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), logger.RequestIDContextKey, requestID))
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}