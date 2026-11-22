package routers

import (
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/app/middlewares"
	"github.com/bonarizki/gin-boilerplate/internal/app/services"
	"github.com/bonarizki/gin-boilerplate/internal/app/services/auth"
	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/gin-gonic/gin"
)

// Deps are the services the HTTP layer needs; they are built once in main.
type Deps struct {
	Auth    auth.AuthServicer
	Example *services.ExampleService
	Health  *services.HealthService
}

// SetupRoute builds the gin engine with global middleware and all routes.
func SetupRoute(deps Deps) *gin.Engine {
	cfg := config.Get()
	debug := cfg != nil && cfg.Server.Debug
	var trustedProxies []string
	var requestTimeout time.Duration
	if cfg != nil {
		trustedProxies = cfg.Server.TrustedProxies
		requestTimeout = time.Duration(cfg.Server.RequestTimeoutSeconds) * time.Second
	}
	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.SetTrustedProxies(trustedProxies)
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.RecoveryMiddleware())
	router.Use(middlewares.RequestTimeoutMiddleware(requestTimeout))
	router.Use(middlewares.SecurityHeadersMiddleware())
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.BodyLimitMiddleware())
	router.Use(middlewares.RequestLogMiddleware())
	router.Use(middlewares.MetricsMiddleware())

	RegisterRoutes(router, deps)

	// Swagger UI and OpenAPI spec (only when Debug is true)
	RegisterSwaggerRoutes(router, true)

	return router
}
