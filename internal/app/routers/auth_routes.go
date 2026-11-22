package routers

import (
	"github.com/bonarizki/gin-boilerplate/internal/app/controllers"
	"github.com/bonarizki/gin-boilerplate/internal/app/middlewares"
	"github.com/bonarizki/gin-boilerplate/internal/app/services/auth"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers authentication routes under the given group (e.g. /api/v1).
// Full path will be groupPrefix/auth/... Stricter auth rate limits apply in addition to
// the global limit on the parent group (see index.go).
func RegisterAuthRoutes(group *gin.RouterGroup, authService auth.AuthServicer) *controllers.AuthController {
	authController := controllers.NewAuthController(authService)
	authRoutes := group.Group("/auth")
	authRoutes.Use(middlewares.AuthRateLimitMiddleware())
	{
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/refresh", authController.RefreshToken)
		authRoutes.POST("/logout", authController.Logout)
		authRoutes.POST("/forgot-password", authController.ForgotPassword)
		authRoutes.POST("/reset-password", authController.ResetPassword)
	}
	return authController
}
