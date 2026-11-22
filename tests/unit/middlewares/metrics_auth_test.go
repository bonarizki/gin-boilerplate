package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bonarizki/gin-boilerplate/internal/app/middlewares"
	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMetricsAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("open when token unset", func(t *testing.T) {
		viper.Set("METRICS_TOKEN", "")
		defer viper.Set("METRICS_TOKEN", "")

		router := gin.New()
		router.GET("/metrics", middlewares.MetricsAuthMiddleware(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("401 without token when configured", func(t *testing.T) {
		viper.Set("METRICS_TOKEN", "secret-metrics")
		defer viper.Set("METRICS_TOKEN", "")

		router := gin.New()
		router.GET("/metrics", middlewares.MetricsAuthMiddleware(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("allows bearer token", func(t *testing.T) {
		viper.Set("METRICS_TOKEN", "secret-metrics")
		defer viper.Set("METRICS_TOKEN", "")
		_ = config.Get()

		router := gin.New()
		router.GET("/metrics", middlewares.MetricsAuthMiddleware(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
		req.Header.Set("Authorization", "Bearer secret-metrics")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}