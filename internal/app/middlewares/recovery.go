package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware turns a handler panic into a standard JSON 500 response
// and logs the panic value and stack trace with the request ID. The panic
// value is never sent to the client.
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		logger.WithRequestID(c.GetString("request_id")).
			WithField("stack", string(debug.Stack())).
			Errorf("panic recovered: %v", recovered)
		utils.HandleErrors(c, http.StatusInternalServerError, nil, "Internal server error")
		c.Abort()
	})
}
