package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/domain/models"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// UserRepository defines data access for user entity (used by auth and others).
type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	GetUserByPasswordResetToken(ctx context.Context, token string) (*models.User, error)
	ClearExpiredPasswordResets(ctx context.Context, cutoff time.Time) (int64, error)
}

// userRepo is the default implementation of UserRepository.
type userRepo struct {
	db *gorm.DB
}

// NewUserRepository returns a UserRepository backed by db.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// primary scopes a query to ctx and pins it to the primary, since auth reads
// must see writes that just happened.
func primary(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx).Clauses(dbresolver.Write)
}

func (r *userRepo) CreateUser(ctx context.Context, user *models.User) error {
	if err := primary(ctx, r.db).Create(user).Error; err != nil {
		logger.Errorf("failed to create user: %v", err)
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := primary(ctx, r.db).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Errorf("failed to get user by email: %v", err)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, user *models.User) error {
	if err := primary(ctx, r.db).Save(user).Error; err != nil {
		logger.Errorf("failed to update user: %v", err)
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := primary(ctx, r.db).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Errorf("failed to get user by id: %v", err)
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetUserByPasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := primary(ctx, r.db).Where("password_reset_token = ?", token).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Errorf("failed to get user by password reset token: %v", err)
		return nil, fmt.Errorf("failed to get user by password reset token: %w", err)
	}
	return &user, nil
}

func (r *userRepo) ClearExpiredPasswordResets(ctx context.Context, cutoff time.Time) (int64, error) {
	result := primary(ctx, r.db).Model(&models.User{}).
		Where("password_reset_expiry IS NOT NULL AND password_reset_expiry < ?", cutoff).
		Updates(map[string]interface{}{
			"password_reset_token":  nil,
			"password_reset_expiry": nil,
		})
	if result.Error != nil {
		logger.Errorf("failed to clear expired password reset tokens: %v", result.Error)
		return 0, fmt.Errorf("failed to clear expired password reset tokens: %w", result.Error)
	}
	return result.RowsAffected, nil
}
