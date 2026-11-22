package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/domain/models"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"gorm.io/gorm"
)

// RefreshTokenRepository defines data access for the refresh token rotation chain.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByTokenHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	// Rotate atomically revokes the active token oldID and inserts newToken in one
	// transaction. Returns rotated=false when the row was already revoked (concurrent
	// refresh or reuse) without an error.
	Rotate(ctx context.Context, oldID uint, newToken *models.RefreshToken) (rotated bool, err error)
	// MarkRotated revokes the token identified by id. newTokenHash is recorded
	// as ReplacedByHash when non-empty (rotation); left empty for a terminal
	// revoke (logout). Only affects rows that are still active (revoked_at IS NULL).
	MarkRotated(ctx context.Context, id uint, newTokenHash string) error
	// RevokeFamily revokes every still-active token sharing familyID.
	// Used for reuse (theft) detection: a rotated-out token being replayed
	// means the whole chain must be killed.
	RevokeFamily(ctx context.Context, familyID string) error
	// RevokeAllForUser revokes every still-active token for userID (logout-all-devices).
	RevokeAllForUser(ctx context.Context, userID uint) error
	// DeleteExpiredBefore removes refresh tokens expired or long-revoked before cutoff.
	DeleteExpiredBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

type refreshTokenRepo struct {
	db *gorm.DB
}

// NewRefreshTokenRepository returns a RefreshTokenRepository backed by db.
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(ctx context.Context, token *models.RefreshToken) error {
	if err := primary(ctx, r.db).Create(token).Error; err != nil {
		logger.Errorf("failed to create refresh token: %v", err)
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepo) GetByTokenHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := primary(ctx, r.db).Where("token_hash = ?", hash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Errorf("failed to get refresh token by hash: %v", err)
		return nil, fmt.Errorf("failed to get refresh token by hash: %w", err)
	}
	return &token, nil
}

func (r *refreshTokenRepo) Rotate(ctx context.Context, oldID uint, newToken *models.RefreshToken) (bool, error) {
	rotated := false
	err := primary(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&models.RefreshToken{}).
			Where("id = ? AND revoked_at IS NULL", oldID).
			Updates(map[string]interface{}{
				"revoked_at":       now,
				"replaced_by_hash": newToken.TokenHash,
			})
		if result.Error != nil {
			logger.Errorf("failed to revoke refresh token for rotation: %v", result.Error)
			return fmt.Errorf("failed to revoke refresh token for rotation: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return nil
		}
		if err := tx.Create(newToken).Error; err != nil {
			logger.Errorf("failed to create rotated refresh token: %v", err)
			return fmt.Errorf("failed to create rotated refresh token: %w", err)
		}
		rotated = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return rotated, nil
}

func (r *refreshTokenRepo) MarkRotated(ctx context.Context, id uint, newTokenHash string) error {
	updates := map[string]interface{}{
		"revoked_at":       time.Now(),
		"replaced_by_hash": newTokenHash,
	}
	result := primary(ctx, r.db).Model(&models.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Updates(updates)
	if result.Error != nil {
		logger.Errorf("failed to mark refresh token rotated: %v", result.Error)
		return fmt.Errorf("failed to mark refresh token rotated: %w", result.Error)
	}
	return nil
}

func (r *refreshTokenRepo) RevokeFamily(ctx context.Context, familyID string) error {
	err := primary(ctx, r.db).Model(&models.RefreshToken{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Update("revoked_at", time.Now()).Error
	if err != nil {
		logger.Errorf("failed to revoke refresh token family: %v", err)
		return fmt.Errorf("failed to revoke refresh token family: %w", err)
	}
	return nil
}

func (r *refreshTokenRepo) RevokeAllForUser(ctx context.Context, userID uint) error {
	err := primary(ctx, r.db).Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now()).Error
	if err != nil {
		logger.Errorf("failed to revoke all refresh tokens for user: %v", err)
		return fmt.Errorf("failed to revoke all refresh tokens for user: %w", err)
	}
	return nil
}

func (r *refreshTokenRepo) DeleteExpiredBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	result := primary(ctx, r.db).
		Where("expires_at < ? OR (revoked_at IS NOT NULL AND revoked_at < ?)", cutoff, cutoff).
		Delete(&models.RefreshToken{})
	if result.Error != nil {
		logger.Errorf("failed to delete expired refresh tokens: %v", result.Error)
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", result.Error)
	}
	return result.RowsAffected, nil
}
