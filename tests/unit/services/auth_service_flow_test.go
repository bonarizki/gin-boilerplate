package services_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/app/dto"
	"github.com/bonarizki/gin-boilerplate/internal/app/services/auth"
	"github.com/bonarizki/gin-boilerplate/internal/domain/models"
	"github.com/bonarizki/gin-boilerplate/tests/mocks"
	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var hex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// TestRegisterRequestValidation runs the real binding validator the
// controllers use on dto.RegisterRequest.
func TestRegisterRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.RegisterRequest
		wantErr bool
	}{
		{"valid", dto.RegisterRequest{Name: "John Doe", Email: "john@example.com", Password: "SecurePass123!"}, false},
		{"empty name", dto.RegisterRequest{Email: "john@example.com", Password: "SecurePass123!"}, true},
		{"short name", dto.RegisterRequest{Name: "Jo", Email: "john@example.com", Password: "SecurePass123!"}, true},
		{"invalid email", dto.RegisterRequest{Name: "John Doe", Email: "invalid-email", Password: "SecurePass123!"}, true},
		{"short password", dto.RegisterRequest{Name: "John Doe", Email: "john@example.com", Password: "short"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := binding.Validator.ValidateStruct(&tt.req)
			assert.Equal(t, tt.wantErr, err != nil, "ValidateStruct() = %v", err)
		})
	}
}

func TestRegister(t *testing.T) {

	t.Run("stores bcrypt hash and normalized email", func(t *testing.T) {
		repo := mocks.NewMockUserRepository()
		svc := newAuthServiceWithRepo(repo)

		resp, err := svc.Register(context.Background(), &dto.RegisterRequest{
			Name: "Ada", Email: "  Ada@Example.COM ", Password: "SecurePass123!",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.Regexp(t, hex64, resp.RefreshToken)

		stored, _ := repo.GetUserByEmail(context.Background(), "ada@example.com")
		require.NotNil(t, stored, "user should be stored under the normalized email")
		assert.NotEqual(t, "SecurePass123!", stored.Password)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("SecurePass123!")))
	})

	t.Run("duplicate email", func(t *testing.T) {
		repo := mocks.NewMockUserRepository()
		repo.AddUserByEmail("ada@example.com", &models.User{ID: 1, Name: "Ada"})
		svc := newAuthServiceWithRepo(repo)

		_, err := svc.Register(context.Background(), &dto.RegisterRequest{
			Name: "Ada", Email: "ADA@example.com", Password: "SecurePass123!",
		})
		assert.ErrorIs(t, err, auth.ErrEmailAlreadyExists)
	})
}

func TestLogin(t *testing.T) {

	hash, err := bcrypt.GenerateFromPassword([]byte("SecurePass123!"), bcrypt.MinCost)
	require.NoError(t, err)
	repo := mocks.NewMockUserRepository()
	repo.AddUserByEmail("ada@example.com", &models.User{ID: 7, Name: "Ada", Password: string(hash)})
	tokens := mocks.NewMockRefreshTokenRepository()
	svc := auth.NewAuthService(repo, tokens, auth.NewLogMailer(), testAuthConfig)

	t.Run("unknown email", func(t *testing.T) {
		_, err := svc.Login(context.Background(), &dto.LoginRequest{Email: "nobody@example.com", Password: "SecurePass123!"})
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login(context.Background(), &dto.LoginRequest{Email: "ada@example.com", Password: "wrong-password"})
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})

	t.Run("success issues a stored refresh token", func(t *testing.T) {
		resp, err := svc.Login(context.Background(), &dto.LoginRequest{Email: "Ada@Example.com", Password: "SecurePass123!"})
		require.NoError(t, err)
		assert.Equal(t, uint(7), resp.User.ID)

		userID, err := svc.ValidateToken(resp.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, uint(7), userID)

		rec, err := tokens.GetByTokenHash(context.Background(), testHashToken(resp.RefreshToken))
		require.NoError(t, err)
		require.NotNil(t, rec, "refresh token must be persisted by hash")
		assert.Equal(t, uint(7), rec.UserID)
	})
}

// TestRefreshTokensAreUniqueHex checks tokens come from crypto/rand: 64
// lowercase hex characters and no repeats across many sessions.
func TestRefreshTokensAreUniqueHex(t *testing.T) {

	hash, err := bcrypt.GenerateFromPassword([]byte("SecurePass123!"), bcrypt.MinCost)
	require.NoError(t, err)
	repo := mocks.NewMockUserRepository()
	repo.AddUserByEmail("ada@example.com", &models.User{ID: 1, Password: string(hash)})
	svc := newAuthServiceWithRepo(repo)

	seen := make(map[string]bool)
	for i := 0; i < 200; i++ {
		resp, err := svc.Login(context.Background(), &dto.LoginRequest{Email: "ada@example.com", Password: "SecurePass123!"})
		require.NoError(t, err)
		require.Regexp(t, hex64, resp.RefreshToken)
		require.False(t, seen[resp.RefreshToken], "duplicate refresh token after %d logins", i)
		seen[resp.RefreshToken] = true
	}
}

// TestForgotPasswordSetsFifteenMinuteExpiry checks the stored reset token
// is a hash (not the emailed value) expiring 15 minutes from now.
func TestForgotPasswordSetsFifteenMinuteExpiry(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	repo.AddUserByEmail("ada@example.com", &models.User{ID: 1, Name: "Ada"})
	mailer := &captureMailer{}
	svc := auth.NewAuthService(repo, mocks.NewMockRefreshTokenRepository(), mailer, testAuthConfig)

	before := time.Now()
	require.NoError(t, svc.ForgotPassword(context.Background(), &dto.ForgotPasswordRequest{Email: "ada@example.com"}))

	user, _ := repo.GetUserByEmail(context.Background(), "ada@example.com")
	require.NotNil(t, user.PasswordResetExpiry)
	assert.WithinDuration(t, before.Add(15*time.Minute), *user.PasswordResetExpiry, 5*time.Second)
	assert.Regexp(t, hex64, mailer.token)
	assert.Equal(t, testHashToken(mailer.token), user.PasswordResetToken, "only the hash may be stored")
}

func TestGetProfile(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	repo.AddUserByEmail("ada@example.com", &models.User{ID: 3, Name: "Ada"})
	svc := newAuthServiceWithRepo(repo)

	got, err := svc.GetProfile(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, "ada@example.com", got.Email)

	_, err = svc.GetProfile(context.Background(), 99)
	assert.True(t, errors.Is(err, auth.ErrUserNotFound), "err = %v", err)
}

type captureMailer struct {
	token string
}

func (m *captureMailer) SendPasswordResetEmail(_ string, resetToken string) error {
	m.token = resetToken
	return nil
}
