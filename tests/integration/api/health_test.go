package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bonarizki/gin-boilerplate/internal/app/routers"
	"github.com/bonarizki/gin-boilerplate/internal/app/services"
	"github.com/bonarizki/gin-boilerplate/tests/integration/testdb"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func healthRouter(checker services.DatabaseChecker) *gin.Engine {
	gin.SetMode(gin.TestMode)
	svc := services.NewHealthService()
	svc.AddChecker("database", checker)
	router := gin.New()
	routers.RegisterHealthRoutes(router, svc)
	return router
}

func getHealth(router *gin.Engine) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	return w
}

func TestHealth_DatabaseDownIs503(t *testing.T) {
	w := getHealth(healthRouter(services.DatabaseChecker{}))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"unhealthy"`)
}

func TestHealth_DatabaseUpIs200(t *testing.T) {
	sqlDB, err := testdb.Open(t).DB()
	require.NoError(t, err)

	w := getHealth(healthRouter(services.DatabaseChecker{DB: sqlDB}))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"healthy"`)
}
