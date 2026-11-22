package services

import (
	"context"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/domain/repositories"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
)

// TokenCleanup periodically deletes expired refresh tokens and clears stale
// password-reset fields.
type TokenCleanup struct {
	Users         repositories.UserRepository
	RefreshTokens repositories.RefreshTokenRepository
	// Interval <= 0 disables the job.
	Interval time.Duration
	// Retention is how long expired or revoked refresh tokens are kept.
	Retention time.Duration
}

// Run blocks, cleaning once immediately and then every Interval, until ctx
// is cancelled.
func (t TokenCleanup) Run(ctx context.Context) {
	if t.Interval <= 0 {
		return
	}
	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()
	for {
		t.runOnce(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (t TokenCleanup) runOnce(ctx context.Context) {
	cutoff := time.Now().Add(-t.Retention)
	if n, err := t.RefreshTokens.DeleteExpiredBefore(ctx, cutoff); err != nil {
		logger.Warnf("refresh token cleanup failed: %v", err)
	} else if n > 0 {
		logger.Infof("deleted %d expired refresh token rows", n)
	}
	if n, err := t.Users.ClearExpiredPasswordResets(ctx, time.Now()); err != nil {
		logger.Warnf("password reset cleanup failed: %v", err)
	} else if n > 0 {
		logger.Infof("cleared %d expired password reset tokens", n)
	}
}
