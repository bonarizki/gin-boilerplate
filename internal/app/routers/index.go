package routers

import (
	"github.com/bonarizki/gin-boilerplate/internal/app/middlewares"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes adds all routing list here and delegates to Register* per feature.
// API versioning: /health and /metrics at root; /api/v1/* for versioned API.
func RegisterRoutes(route *gin.Engine, deps Deps) {
	route.NoRoute(func(ctx *gin.Context) {
		utils.NotFound(ctx, nil, "Route not found")
	})

	RegisterHealthRoutes(route, deps.Health)

	apiV1 := route.Group("/api/v1")
	apiV1.Use(middlewares.RateLimitMiddleware())

	authController := RegisterAuthRoutes(apiV1, deps.Auth)
	RegisterExampleRoutes(apiV1, deps.Example)

	// Protected routes (require authentication)
	protectedRoutes := apiV1.Group("")
	protectedRoutes.Use(middlewares.AuthMiddleware(deps.Auth))
	{
		protectedRoutes.GET("/profile", authController.Profile)
		protectedRoutes.POST("/logout-all", authController.LogoutAll)
	}
}
