package middlewares_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bonarizki/gin-boilerplate/internal/app/middlewares"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bodyLimitRouter(t *testing.T, limit int) *gin.Engine {
	t.Helper()
	viper.Set("MAX_BODY_BYTES", limit)
	t.Cleanup(func() { viper.Set("MAX_BODY_BYTES", 0) })

	router := setupTestRouter()
	router.Use(middlewares.BodyLimitMiddleware())
	router.POST("/echo", func(c *gin.Context) {
		var body map[string]any
		if err := c.ShouldBindJSON(&body); err != nil {
			utils.HandleErrors(c, http.StatusBadRequest, err, "Invalid request")
			return
		}
		c.JSON(http.StatusOK, body)
	})
	return router
}

func TestBodyLimit_DeclaredLengthOverLimitIs413(t *testing.T) {
	router := bodyLimitRouter(t, 16)
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"name":"this is too long"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Contains(t, w.Body.String(), "request body exceeds 16 bytes")
}

func TestBodyLimit_ChunkedBodyOverLimitIs413(t *testing.T) {
	router := bodyLimitRouter(t, 16)
	// io.MultiReader hides the length, so httptest sends ContentLength -1.
	body := io.MultiReader(strings.NewReader(`{"name":"this is too long"}`))
	req := httptest.NewRequest(http.MethodPost, "/echo", body)
	req.ContentLength = -1
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestBodyLimit_SmallBodyPasses(t *testing.T) {
	router := bodyLimitRouter(t, 1024)
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"a":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecovery_ReturnsJSON500WithoutPanicValue(t *testing.T) {
	var logs bytes.Buffer
	logger.SetOutput(&logs)
	t.Cleanup(func() { logger.SetOutput(os.Stdout) })

	router := setupTestRouter()
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.RecoveryMiddleware())
	router.GET("/boom", func(c *gin.Context) {
		panic("secret internal detail")
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.NotContains(t, w.Body.String(), "secret internal detail")

	assert.Contains(t, logs.String(), "panic recovered: secret internal detail")
	assert.Contains(t, logs.String(), w.Header().Get("X-Request-ID"))
}
