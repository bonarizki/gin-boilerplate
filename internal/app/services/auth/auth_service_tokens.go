package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/app/dto"
	"github.com/bonarizki/gin-boilerplate/internal/domain/models"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
)

// RefreshToken generates new access and refresh tokens using a valid refresh token.
//
// The presented token is rotated: it is revoked and replaced by a newly
// issued token in the same family. If a token that was already rotated (or
// revoked) is presented again, this is treated as reuse/theft and the entire
// token family is revoked, forcing re-login on every device in that session.
//
// Returns ErrInvalidRefreshToken if the refresh token is invalid, revoked, or expired.
func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (resp *dto.RefreshTokenResponse, err error) {
	tokenRecord, err := s.refreshTokenRepo.GetByTokenHash(ctx, hashToken(req.RefreshToken))
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to get refresh token: %v", err)
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	if tokenRecord == nil {
		logger.FromContext(ctx).Warnf("refresh token attempt with unknown token")
		return nil, ErrInvalidRefreshToken
	}

	if tokenRecord.RevokedAt != nil {
		logger.FromContext(ctx).Warnf("refresh token reuse detected for user %d, revoking family %s", tokenRecord.UserID, tokenRecord.FamilyID)
		if revokeErr := s.refreshTokenRepo.RevokeFamily(ctx, tokenRecord.FamilyID); revokeErr != nil {
			logger.FromContext(ctx).Errorf("failed to revoke token family after reuse detection: %v", revokeErr)
		}
		return nil, ErrInvalidRefreshToken
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		logger.FromContext(ctx).Warnf("refresh token attempt with expired token for user %d", tokenRecord.UserID)
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.GetUserByID(ctx, tokenRecord.UserID)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to get user for refresh token: %v", err)
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Generate new access token
	accessToken, err := s.generateToken(user)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to generate new access token: %v", err)
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	rawRefresh, err := s.generateRefreshToken()
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to generate rotated refresh token: %v", err)
		return nil, fmt.Errorf("failed to generate rotated refresh token: %w", err)
	}
	newRecord := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(rawRefresh),
		FamilyID:  tokenRecord.FamilyID,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTTL),
	}
	rotated, err := s.refreshTokenRepo.Rotate(ctx, tokenRecord.ID, newRecord)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to rotate refresh token: %v", err)
		return nil, fmt.Errorf("failed to rotate refresh token: %w", err)
	}
	if !rotated {
		logger.FromContext(ctx).Warnf("concurrent refresh or reuse for user %d, revoking family %s", tokenRecord.UserID, tokenRecord.FamilyID)
		if revokeErr := s.refreshTokenRepo.RevokeFamily(ctx, tokenRecord.FamilyID); revokeErr != nil {
			logger.FromContext(ctx).Errorf("failed to revoke token family after failed rotation: %v", revokeErr)
		}
		return nil, ErrInvalidRefreshToken
	}
	newRefreshToken := rawRefresh

	logger.FromContext(ctx).Infof("refresh token successful for user: %s", utils.MaskEmail(user.Email))

	// Build response
	response := &dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
	}

	return response, nil
}

// Logout revokes the given refresh token (this device/session only).
//
// Always returns nil for an unknown or already-revoked token so callers
// cannot use this endpoint to probe whether a token exists.
func (s *AuthService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	tokenRecord, err := s.refreshTokenRepo.GetByTokenHash(ctx, hashToken(req.RefreshToken))
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to get refresh token for logout: %v", err)
		return fmt.Errorf("failed to process logout: %w", err)
	}
	if tokenRecord == nil || tokenRecord.RevokedAt != nil {
		return nil
	}

	if err := s.refreshTokenRepo.MarkRotated(ctx, tokenRecord.ID, ""); err != nil {
		logger.FromContext(ctx).Errorf("failed to revoke refresh token on logout: %v", err)
		return fmt.Errorf("failed to process logout: %w", err)
	}

	logger.FromContext(ctx).Infof("logout successful for user: %d", tokenRecord.UserID)
	return nil
}

// LogoutAll revokes every active refresh token for userID (all devices/sessions).
func (s *AuthService) LogoutAll(ctx context.Context, userID uint) error {
	if err := s.refreshTokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		logger.FromContext(ctx).Errorf("failed to revoke all refresh tokens: %v", err)
		return fmt.Errorf("failed to process logout-all: %w", err)
	}

	logger.FromContext(ctx).Infof("logout-all successful for user: %d", userID)
	return nil
}

// ForgotPassword generates a password reset token and sends it via the configured mailer.
// The raw token is never returned to callers.
// Returns ErrUserNotFound if email doesn't exist.
func (s *AuthService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) (err error) {
	if s.mailer == nil {
		err = ErrMailerNotConfigured
		return err
	}

	req.Email = normalizeEmail(req.Email)

	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to get user by email: %v", err)
		return fmt.Errorf("failed to process request: %w", err)
	}

	if user == nil {
		logger.FromContext(ctx).Warnf("password reset attempt for non-existent email: %s", utils.MaskEmail(req.Email))
		return ErrUserNotFound
	}

	resetToken, err := s.generatePasswordResetToken()
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to generate reset token: %v", err)
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	expiry := time.Now().Add(15 * time.Minute)
	user.PasswordResetToken = hashToken(resetToken)
	user.PasswordResetExpiry = &expiry

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		logger.FromContext(ctx).Errorf("failed to save reset token: %v", err)
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	logger.FromContext(ctx).Infof("password reset token generated for user: %s", utils.MaskEmail(user.Email))

	if err := s.mailer.SendPasswordResetEmail(user.Email, resetToken); err != nil {
		logger.FromContext(ctx).Errorf("failed to send password reset email: %v", err)
		return fmt.Errorf("failed to send reset email: %w", err)
	}
	return nil
}

// ResetPassword resets user password using a valid reset token.
//
// Returns ErrInvalidResetToken if token is invalid.
// Returns ErrResetTokenExpired if token has expired.
func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) (err error) {
	user, err := s.userRepo.GetUserByPasswordResetToken(ctx, hashToken(req.Token))
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to get user by reset token: %v", err)
		return fmt.Errorf("failed to validate reset token: %w", err)
	}

	if user == nil {
		logger.FromContext(ctx).Warnf("password reset attempt with invalid token")
		return ErrInvalidResetToken
	}

	// Check if token has expired
	if user.PasswordResetExpiry == nil || time.Now().After(*user.PasswordResetExpiry) {
		logger.FromContext(ctx).Warnf("password reset attempt with expired token for user: %s", utils.MaskEmail(user.Email))
		return ErrResetTokenExpired
	}

	// Hash new password
	hashedPassword, err := s.hashPassword(req.NewPassword)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to hash new password: %v", err)
		return fmt.Errorf("failed to process password: %w", err)
	}

	// Update password and clear reset token
	user.Password = hashedPassword
	user.PasswordResetToken = ""
	user.PasswordResetExpiry = nil

	// Save to database
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		logger.FromContext(ctx).Errorf("failed to update password: %v", err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := s.refreshTokenRepo.RevokeAllForUser(ctx, user.ID); err != nil {
		logger.FromContext(ctx).Errorf("failed to revoke refresh tokens after password reset: %v", err)
		return fmt.Errorf("failed to revoke sessions after password reset: %w", err)
	}

	logger.FromContext(ctx).Infof("password reset successful for user: %s", utils.MaskEmail(user.Email))

	return nil
}

// generatePasswordResetToken creates a cryptographically secure random reset token.
//
// Returns a 64-character hexadecimal string.
func (s *AuthService) generatePasswordResetToken() (string, error) {
	// Generate 32 random bytes (will be 64 chars in hex)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
