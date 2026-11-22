package controllers

import (
	"errors"
	"net/http"

	"github.com/bonarizki/gin-boilerplate/internal/app/dto"
	"github.com/bonarizki/gin-boilerplate/internal/app/services/auth"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/bonarizki/gin-boilerplate/pkg/types"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/gin-gonic/gin"
)

// authErrToAPIError maps known auth domain errors to types.APIError for HTTP response.
// Returns nil if err is not a known auth error.
func authErrToAPIError(err error) *types.APIError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, auth.ErrEmailAlreadyExists):
		return &types.APIError{Code: http.StatusConflict, Message: "Email already exists", Details: nil}
	case errors.Is(err, auth.ErrInvalidCredentials):
		return &types.APIError{Code: http.StatusUnauthorized, Message: "Invalid email or password", Details: nil}
	case errors.Is(err, auth.ErrInvalidRefreshToken):
		return &types.APIError{Code: http.StatusUnauthorized, Message: "Invalid or expired refresh token", Details: nil}
	case errors.Is(err, auth.ErrUserNotFound):
		return &types.APIError{Code: http.StatusNotFound, Message: "User not found", Details: nil}
	case errors.Is(err, auth.ErrInvalidResetToken):
		return &types.APIError{Code: http.StatusBadRequest, Message: "Invalid reset token", Details: nil}
	case errors.Is(err, auth.ErrResetTokenExpired):
		return &types.APIError{Code: http.StatusBadRequest, Message: "Reset token has expired", Details: nil}
	default:
		return nil
	}
}

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	service auth.AuthServicer
}

// NewAuthController creates a new AuthController instance
func NewAuthController(service auth.AuthServicer) *AuthController {
	return &AuthController{
		service: service,
	}
}

// Register handles user registration endpoint.
//
// POST /auth/register
// Request body: RegisterRequest (JSON)
// Response: AuthResponse with user info and JWT token
func (ctrl *AuthController) Register(c *gin.Context) {
	var err error
	ctx := c.Request.Context()

	var req dto.RegisterRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("invalid registration request: %v", err)
		utils.BadRequest(c, err, "Invalid request data")
		return
	}

	response, err := ctrl.service.Register(ctx, &req)
	if err != nil {
		if apiErr := authErrToAPIError(err); apiErr != nil {
			utils.RespondWithAPIError(c, apiErr)
			return
		}
		logger.Errorf("registration failed: %v", err)
		utils.InternalServerError(c, err, "Failed to register user")
		return
	}

	utils.Created(c, response, "User registered successfully")
}

// Login handles user authentication endpoint.
//
// POST /auth/login
// Request body: LoginRequest (JSON)
// Response: AuthResponse with user info and JWT token
func (ctrl *AuthController) Login(c *gin.Context) {
	var err error
	ctx := c.Request.Context()

	var req dto.LoginRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("invalid login request: %v", err)
		utils.BadRequest(c, err, "Invalid request data")
		return
	}

	response, err := ctrl.service.Login(ctx, &req)
	if err != nil {
		if apiErr := authErrToAPIError(err); apiErr != nil {
			utils.RespondWithAPIError(c, apiErr)
			return
		}
		logger.Errorf("login failed: %v", err)
		utils.InternalServerError(c, err, "Failed to authenticate user")
		return
	}

	utils.Ok(c, response, "Login successful")
}

// RefreshToken handles token refresh endpoint.
//
// POST /auth/refresh
// Request body: RefreshTokenRequest (JSON)
// Response: RefreshTokenResponse with new access and refresh tokens
func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	var err error
	ctx := c.Request.Context()

	var req dto.RefreshTokenRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("invalid refresh token request: %v", err)
		utils.BadRequest(c, err, "Invalid request data")
		return
	}

	response, err := ctrl.service.RefreshToken(ctx, &req)
	if err != nil {
		if apiErr := authErrToAPIError(err); apiErr != nil {
			utils.RespondWithAPIError(c, apiErr)
			return
		}
		logger.Errorf("token refresh failed: %v", err)
		utils.InternalServerError(c, err, "Failed to refresh token")
		return
	}

	utils.Ok(c, response, "Token refreshed successfully")
}

// ForgotPassword handles forgot password endpoint.
//
// POST /auth/forgot-password
// Request body: ForgotPasswordRequest (JSON)
// Response: Success message (token sent via email in production)
func (ctrl *AuthController) ForgotPassword(c *gin.Context) {
	var err error
	ctx := c.Request.Context()

	var req dto.ForgotPasswordRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("invalid forgot password request: %v", err)
		utils.BadRequest(c, err, "Invalid request data")
		return
	}

	if err := ctrl.service.ForgotPassword(ctx, &req); err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			utils.Ok(c, map[string]interface{}{"message": "If the email exists, a password reset link has been sent"}, "Password reset initiated")
			return
		}
		if apiErr := authErrToAPIError(err); apiErr != nil {
			utils.RespondWithAPIError(c, apiErr)
			return
		}
		logger.Errorf("forgot password failed: %v", err)
		utils.InternalServerError(c, err, "Failed to process request")
		return
	}

	utils.Ok(c, map[string]interface{}{"message": "If the email exists, a password reset link has been sent"}, "Password reset initiated")
}

// ResetPassword handles password reset endpoint.
//
// POST /auth/reset-password
// Request body: ResetPasswordRequest (JSON)
// Response: Success message
func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var err error
	ctx := c.Request.Context()

	var req dto.ResetPasswordRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("invalid reset password request: %v", err)
		utils.BadRequest(c, err, "Invalid request data")
		return
	}

	err = ctrl.service.ResetPassword(ctx, &req)
	if err != nil {
		if apiErr := authErrToAPIError(err); apiErr != nil {
			utils.RespondWithAPIError(c, apiErr)
			return
		}
		logger.Errorf("password reset failed: %v", err)
		utils.InternalServerError(c, err, "Failed to reset password")
		return
	}

	utils.Ok(c, nil, "Password reset successfully")
}

// Logout handles single-session logout by revoking the given refresh token.
//
// POST /auth/logout
// Request body: LogoutRequest (JSON)
// Response: success regardless of whether the token existed (avoids leaking token validity)
func (ctrl *AuthController) Logout(c *gin.Context) {
	var err error
	ctx := c.Request.Context()

	var req dto.LogoutRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("invalid logout request: %v", err)
		utils.BadRequest(c, err, "Invalid request data")
		return
	}

	if err = ctrl.service.Logout(ctx, &req); err != nil {
		logger.Errorf("logout failed: %v", err)
		utils.InternalServerError(c, err, "Failed to logout")
		return
	}

	utils.Ok(c, nil, "Logout successful")
}

// LogoutAll revokes every refresh token for the authenticated user (all devices/sessions).
//
// POST /api/v1/logout-all (requires JWT)
// Response: success message
func (ctrl *AuthController) LogoutAll(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetUint("user_id")
	if err := ctrl.service.LogoutAll(ctx, userID); err != nil {
		logger.Errorf("logout-all failed: %v", err)
		utils.InternalServerError(c, err, "Failed to logout from all devices")
		return
	}

	utils.Ok(c, nil, "Logged out from all devices successfully")
}

// Profile returns the current authenticated user's profile.
//
// GET /api/v1/profile (requires JWT)
// Response: UserResponse (id, name, email) in standard response format
func (ctrl *AuthController) Profile(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetUint("user_id")
	profile, err := ctrl.service.GetProfile(ctx, userID)
	if err != nil {
		if apiErr := authErrToAPIError(err); apiErr != nil {
			utils.RespondWithAPIError(c, apiErr)
			return
		}
		logger.Errorf("get profile failed: %v", err)
		utils.InternalServerError(c, err, "Failed to retrieve profile")
		return
	}

	utils.Ok(c, profile, "Profile retrieved successfully")
}
