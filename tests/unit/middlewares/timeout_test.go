package middlewares_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/app/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestTimeoutMiddleware_SlowHandlerHitsDeadline(t *testing.T) {
	router := setupTestRouter()
	router.Use(middlewares.RequestTimeoutMiddleware(20 * time.Millisecond))

	var ctxErr error
	router.GET("/slow", func(c *gin.Context) {
		select {
		case <-c.Request.Context().Done():
			ctxErr = c.Request.Context().Err()
		case <-time.After(time.Second):
		}
		c.Status(http.StatusOK)
	})

	start := time.Now()
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/slow", nil))

	assert.True(t, errors.Is(ctxErr, context.DeadlineExceeded), "got %v", ctxErr)
	assert.Less(t, time.Since(start), 500*time.Millisecond)
}

func TestRequestTimeoutMiddleware_ZeroDisables(t *testing.T) {
	router := setupTestRouter()
	router.Use(middlewares.RequestTimeoutMiddleware(0))

	var hasDeadline bool
	router.GET("/", func(c *gin.Context) {
		_, hasDeadline = c.Request.Context().Deadline()
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.False(t, hasDeadline)
}
