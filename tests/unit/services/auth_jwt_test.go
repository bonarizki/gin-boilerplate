package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/app/dto"
	"github.com/bonarizki/gin-boilerplate/internal/app/services/auth"
	"github.com/bonarizki/gin-boilerplate/tests/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func signTestToken(t *testing.T, method jwt.SigningMethod, claims jwt.MapClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(method, claims).SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return s
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"user_id": 7,
		"iss":     testAuthConfig.Issuer,
		"aud":     testAuthConfig.Audience,
		"exp":     time.Now().Add(time.Minute).Unix(),
	}
}

func TestValidateToken_RoundTrip(t *testing.T) {
	svc := newAuthServiceWithMockRepo()
	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Name: "Ada", Email: "ada@example.com", Password: "SecurePass123!",
	})
	require.NoError(t, err)
	resp, err := svc.Login(context.Background(), &dto.LoginRequest{Email: "ada@example.com", Password: "SecurePass123!"})
	require.NoError(t, err)

	userID, err := svc.ValidateToken(resp.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, resp.User.ID, userID)
}

func TestValidateToken_RejectsWeakOrForeignTokens(t *testing.T) {
	svc := newAuthServiceWithMockRepo()

	ok := signTestToken(t, jwt.SigningMethodHS256, validClaims())
	id, err := svc.ValidateToken(ok)
	require.NoError(t, err)
	assert.Equal(t, uint(7), id)

	cases := map[string]func(jwt.MapClaims) (jwt.SigningMethod, jwt.MapClaims){
		"HS512 instead of HS256": func(c jwt.MapClaims) (jwt.SigningMethod, jwt.MapClaims) { return jwt.SigningMethodHS512, c },
		"missing exp": func(c jwt.MapClaims) (jwt.SigningMethod, jwt.MapClaims) {
			delete(c, "exp")
			return jwt.SigningMethodHS256, c
		},
		"expired beyond leeway": func(c jwt.MapClaims) (jwt.SigningMethod, jwt.MapClaims) {
			c["exp"] = time.Now().Add(-time.Minute).Unix()
			return jwt.SigningMethodHS256, c
		},
		"wrong issuer": func(c jwt.MapClaims) (jwt.SigningMethod, jwt.MapClaims) {
			c["iss"] = "someone-else"
			return jwt.SigningMethodHS256, c
		},
		"wrong audience": func(c jwt.MapClaims) (jwt.SigningMethod, jwt.MapClaims) {
			c["aud"] = "other-api"
			return jwt.SigningMethodHS256, c
		},
		"missing user_id": func(c jwt.MapClaims) (jwt.SigningMethod, jwt.MapClaims) {
			delete(c, "user_id")
			return jwt.SigningMethodHS256, c
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			method, claims := mutate(validClaims())
			_, err := svc.ValidateToken(signTestToken(t, method, claims))
			assert.Error(t, err)
		})
	}
}

func TestValidateToken_NoSecretConfigured(t *testing.T) {
	svc := auth.NewAuthService(mocks.NewMockUserRepository(), mocks.NewMockRefreshTokenRepository(), nil, auth.Config{})
	_, err := svc.ValidateToken(signTestToken(t, jwt.SigningMethodHS256, validClaims()))
	assert.Error(t, err)
}
