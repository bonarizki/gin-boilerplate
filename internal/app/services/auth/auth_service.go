package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/app/dto"
	"github.com/bonarizki/gin-boilerplate/internal/domain/models"
	"github.com/bonarizki/gin-boilerplate/internal/domain/repositories"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Common errors for auth service
var (
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrMailerNotConfigured = errors.New("password reset mailer not configured")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidResetToken   = errors.New("invalid or expired reset token")
	ErrResetTokenExpired   = errors.New("reset token has expired")
)

// Config holds the token settings AuthService needs. Zero TTLs fall back to
// 15 minutes (access) and 7 days (refresh).
type Config struct {
	JWTSecret string
	Issuer    string
	Audience  string
	// AccessTTL is kept short because access tokens are stateless and cannot
	// be revoked; logout only revokes refresh tokens.
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// AuthService handles authentication-related business logic
type AuthService struct {
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	mailer           EmailSender // nil disables ForgotPassword (ErrMailerNotConfigured)
	cfg              Config
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(userRepo repositories.UserRepository, refreshTokenRepo repositories.RefreshTokenRepository, mailer EmailSender, cfg Config) *AuthService {
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 15 * time.Minute
	}
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = 7 * 24 * time.Hour
	}
	return &AuthService{userRepo: userRepo, refreshTokenRepo: refreshTokenRepo, mailer: mailer, cfg: cfg}
}

// Register creates a new user account with validation and password hashing.
//
// Returns ErrEmailAlreadyExists if email is already registered.
// Password is hashed using bcrypt before storage.
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (resp *dto.AuthResponse, err error) {
	req.Email = normalizeEmail(req.Email)

	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	if existingUser != nil {
		logger.FromContext(ctx).Warnf("registration attempt with existing email: %s", utils.MaskEmail(req.Email))
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to hash password: %v", err)
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	// Create user model
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Save to database
	if err = s.userRepo.CreateUser(ctx, user); err != nil {
		logger.FromContext(ctx).Errorf("failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.FromContext(ctx).Infof("user registered successfully: %s", utils.MaskEmail(user.Email))

	response, err := s.issueAuthResponse(ctx, user)
	return response, err
}

// Login authenticates a user with email and password.
//
// Returns ErrInvalidCredentials if email or password is incorrect.
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (resp *dto.AuthResponse, err error) {
	req.Email = normalizeEmail(req.Email)

	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to get user: %v", err)
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if user == nil {
		logger.FromContext(ctx).Warnf("login attempt with non-existent email: %s", utils.MaskEmail(req.Email))
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err = s.verifyPassword(user.Password, req.Password); err != nil {
		logger.FromContext(ctx).Warnf("login attempt with invalid password: %s", utils.MaskEmail(req.Email))
		return nil, ErrInvalidCredentials
	}

	logger.FromContext(ctx).Infof("user logged in successfully: %s", utils.MaskEmail(user.Email))

	response, err := s.issueAuthResponse(ctx, user)
	return response, err
}

// issueAuthResponse generates an access token and a new refresh token rotation
// chain (new session/family) for user, and builds the resulting AuthResponse.
// Shared by Register and Login since both start a new session identically.
func (s *AuthService) issueAuthResponse(ctx context.Context, user *models.User) (*dto.AuthResponse, error) {
	accessToken, err := s.generateToken(user)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to generate access token: %v", err)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	familyID, err := s.newFamilyID()
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to generate token family id: %v", err)
		return nil, fmt.Errorf("failed to generate token family id: %w", err)
	}
	refreshToken, err := s.issueRefreshToken(ctx, user.ID, familyID)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to issue refresh token: %v", err)
		return nil, fmt.Errorf("failed to issue refresh token: %w", err)
	}

	return &dto.AuthResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

// accessClaims is the JWT payload issued by generateToken.
type accessClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ValidateToken validates a JWT access token and returns the user ID.
//
// Only HS256 is accepted, and exp, iss and aud (when configured) are required.
func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
	if s.cfg.JWTSecret == "" {
		return 0, errors.New("JWT secret not configured")
	}
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30 * time.Second),
	}
	if s.cfg.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(s.cfg.Issuer))
	}
	if s.cfg.Audience != "" {
		opts = append(opts, jwt.WithAudience(s.cfg.Audience))
	}

	var claims accessClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(*jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	}, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}
	if claims.UserID == 0 {
		return 0, errors.New("invalid user_id in token")
	}
	return claims.UserID, nil
}

// GetProfile returns the current user's profile data.
//
// Returns ErrUserNotFound if userID doesn't exist (e.g. deleted after token was issued).
func (s *AuthService) GetProfile(ctx context.Context, userID uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logger.FromContext(ctx).Errorf("failed to get user for profile: %v", err)
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &dto.UserResponse{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}

// normalizeEmail lowercases and trims an email address for consistent lookups and storage.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Private helper methods

// hashPassword hashes a plain text password using bcrypt.
func (s *AuthService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// verifyPassword compares a hashed password with a plain text password.
func (s *AuthService) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// generateToken creates a signed HS256 access token for user that expires
// after cfg.AccessTTL.
func (s *AuthService) generateToken(user *models.User) (string, error) {
	if s.cfg.JWTSecret == "" {
		return "", errors.New("JWT secret not configured")
	}
	now := time.Now()
	claims := accessClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			Issuer:    s.cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	if s.cfg.Audience != "" {
		claims.Audience = jwt.ClaimStrings{s.cfg.Audience}
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

// generateRefreshToken creates a cryptographically secure random refresh token.
//
// Returns a 64-character hexadecimal string.
func (s *AuthService) generateRefreshToken() (string, error) {
	// Generate 32 random bytes (will be 64 chars in hex)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken returns the SHA-256 hex digest of a raw token for at-rest storage.
//
// Plain SHA-256 (not bcrypt) is intentional: refresh tokens are 256-bit
// cryptographically random values, not low-entropy user secrets, so a fast
// hash is sufficient to make a stolen database dump useless without needing
// a deliberately slow KDF.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// newFamilyID generates a new rotation-chain identifier, one per login session.
func (s *AuthService) newFamilyID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate family id: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// issueRefreshToken generates a new raw refresh token, persists its hash (with
// expiry) under familyID, and returns the raw token to send to the client.
// The raw value is never stored; only hashToken(raw) is written to the database.
func (s *AuthService) issueRefreshToken(ctx context.Context, userID uint, familyID string) (string, error) {
	raw, err := s.generateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	record := &models.RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(raw),
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTTL),
	}
	if err := s.refreshTokenRepo.Create(ctx, record); err != nil {
		return "", fmt.Errorf("failed to save refresh token: %w", err)
	}
	return raw, nil
}
