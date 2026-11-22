package services

import (
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/domain/repositories"
	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
)

// StartTokenCleanup runs periodic deletion of expired refresh tokens and
// stale password-reset fields when TOKEN_CLEANUP_INTERVAL_MINUTES > 0.
func StartTokenCleanup(userRepo repositories.UserRepository, refreshRepo repositories.RefreshTokenRepository) {
	intervalMin := config.GetInt("TOKEN_CLEANUP_INTERVAL_MINUTES")
	if intervalMin <= 0 {
		return
	}
	retentionDays := config.GetInt("TOKEN_RETENTION_DAYS")
	if retentionDays <= 0 {
		retentionDays = 30
	}

	go func() {
		ticker := time.NewTicker(time.Duration(intervalMin) * time.Minute)
		defer ticker.Stop()
		run := func() {
			cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
			if n, err := refreshRepo.DeleteExpiredBefore(cutoff); err != nil {
				logger.Warnf("refresh token cleanup failed: %v", err)
			} else if n > 0 {
				logger.Infof("deleted %d expired refresh token rows", n)
			}
			if n, err := userRepo.ClearExpiredPasswordResets(time.Now()); err != nil {
				logger.Warnf("password reset cleanup failed: %v", err)
			} else if n > 0 {
				logger.Infof("cleared %d expired password reset tokens", n)
			}
		}
		run()
		for range ticker.C {
			run()
		}
	}()
}