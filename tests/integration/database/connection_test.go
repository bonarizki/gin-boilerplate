package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/domain/models"
	"github.com/bonarizki/gin-boilerplate/internal/domain/repositories"
	"github.com/bonarizki/gin-boilerplate/tests/integration/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests need TEST_DB_MASTER_DSN, e.g.
// postgres://user:password@localhost:5432/gin_test?sslmode=disable

func TestDatabaseConnection(t *testing.T) {
	db := testdb.Open(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())
	assert.Equal(t, 5, sqlDB.Stats().MaxOpenConnections)
}

func TestUserRepository(t *testing.T) {
	ctx := context.Background()
	repo := repositories.NewUserRepository(testdb.Open(t))

	missing, err := repo.GetUserByEmail(ctx, "nobody@example.com")
	require.NoError(t, err)
	assert.Nil(t, missing)

	user := &models.User{Name: "Ada", Email: "ada@example.com", Password: "hash"}
	require.NoError(t, repo.CreateUser(ctx, user))
	require.NotZero(t, user.ID)

	byID, err := repo.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "ada@example.com", byID.Email)

	expired := time.Now().Add(-time.Minute)
	user.PasswordResetToken = "reset-hash"
	user.PasswordResetExpiry = &expired
	require.NoError(t, repo.UpdateUser(ctx, user))

	byToken, err := repo.GetUserByPasswordResetToken(ctx, "reset-hash")
	require.NoError(t, err)
	require.NotNil(t, byToken)
	assert.Equal(t, user.ID, byToken.ID)

	n, err := repo.ClearExpiredPasswordResets(ctx, time.Now())
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
	cleared, err := repo.GetUserByPasswordResetToken(ctx, "reset-hash")
	require.NoError(t, err)
	assert.Nil(t, cleared)
}

func TestRefreshTokenRepository(t *testing.T) {
	ctx := context.Background()
	db := testdb.Open(t)
	users := repositories.NewUserRepository(db)
	tokens := repositories.NewRefreshTokenRepository(db)

	user := &models.User{Name: "Ada", Email: "ada@example.com", Password: "hash"}
	require.NoError(t, users.CreateUser(ctx, user))

	first := &models.RefreshToken{UserID: user.ID, TokenHash: "h1", FamilyID: "fam", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, tokens.Create(ctx, first))

	second := &models.RefreshToken{UserID: user.ID, TokenHash: "h2", FamilyID: "fam", ExpiresAt: time.Now().Add(time.Hour)}
	rotated, err := tokens.Rotate(ctx, first.ID, second)
	require.NoError(t, err)
	assert.True(t, rotated)

	third := &models.RefreshToken{UserID: user.ID, TokenHash: "h3", FamilyID: "fam", ExpiresAt: time.Now().Add(time.Hour)}
	rotated, err = tokens.Rotate(ctx, first.ID, third)
	require.NoError(t, err)
	assert.False(t, rotated, "an already rotated token must not rotate again")

	old, err := tokens.GetByTokenHash(ctx, "h1")
	require.NoError(t, err)
	require.NotNil(t, old.RevokedAt)
	assert.Equal(t, "h2", old.ReplacedByHash)

	require.NoError(t, tokens.RevokeFamily(ctx, "fam"))
	current, err := tokens.GetByTokenHash(ctx, "h2")
	require.NoError(t, err)
	assert.NotNil(t, current.RevokedAt)

	other := &models.RefreshToken{UserID: user.ID, TokenHash: "h4", FamilyID: "fam2", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, tokens.Create(ctx, other))
	require.NoError(t, tokens.RevokeAllForUser(ctx, user.ID))
	got, err := tokens.GetByTokenHash(ctx, "h4")
	require.NoError(t, err)
	assert.NotNil(t, got.RevokedAt)

	n, err := tokens.DeleteExpiredBefore(ctx, time.Now().Add(time.Minute))
	require.NoError(t, err)
	assert.Equal(t, int64(3), n, "all three revoked rows are older than the cutoff")
}

func TestRepositoryHonoursCancelledContext(t *testing.T) {
	repo := repositories.NewUserRepository(testdb.Open(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetUserByEmail(ctx, "ada@example.com")
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled), "got %v", err)
}
