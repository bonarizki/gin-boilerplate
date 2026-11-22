package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bonarizki/gin-boilerplate/internal/app/routers"
	"github.com/bonarizki/gin-boilerplate/internal/app/services"
	"github.com/bonarizki/gin-boilerplate/internal/app/services/auth"
	"github.com/bonarizki/gin-boilerplate/internal/domain/repositories"
	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/bonarizki/gin-boilerplate/tests/integration/testdb"
	"github.com/stretchr/testify/require"
)

type apiClient struct {
	t      *testing.T
	router http.Handler
}

func (c apiClient) do(method, path, bearer string, body any) (int, map[string]any) {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(c.t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	w := httptest.NewRecorder()
	c.router.ServeHTTP(w, req)

	var resp map[string]any
	require.NoError(c.t, json.Unmarshal(w.Body.Bytes(), &resp), w.Body.String())
	data, _ := resp["data"].(map[string]any)
	return w.Code, data
}

// TestAuthFlow drives the real router, services and Postgres through
// register, login, profile, refresh rotation, reuse detection and logout-all.
func TestAuthFlow(t *testing.T) {
	db := testdb.Open(t)

	cfg := &config.Configuration{}
	cfg.Server.AuthRateLimitRPM = 600
	cfg.Server.AuthRateLimitBurst = 100
	cfg.Server.RateLimitRPS = 100
	cfg.Server.RateLimitBurst = 100
	config.SetForTest(cfg)

	authService := auth.NewAuthService(
		repositories.NewUserRepository(db),
		repositories.NewRefreshTokenRepository(db),
		nil,
		auth.Config{JWTSecret: "integration-jwt-secret-at-least-32-chars", Issuer: "it", Audience: "it-api"},
	)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	health := services.NewHealthService()
	health.AddChecker("database", services.DatabaseChecker{DB: sqlDB})
	c := apiClient{t: t, router: routers.SetupRoute(routers.Deps{
		Auth:    authService,
		Example: services.NewExampleService(repositories.NewExampleRepository(db)),
		Health:  health,
	})}

	creds := map[string]string{"name": "Ada Lovelace", "email": "Ada@Example.com", "password": "SecurePass123!"}
	code, _ := c.do(http.MethodPost, "/api/v1/auth/register", "", creds)
	require.Equal(t, http.StatusCreated, code)

	code, login := c.do(http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email": "ada@example.com", "password": "SecurePass123!",
	})
	require.Equal(t, http.StatusOK, code)
	access, _ := login["access_token"].(string)
	refresh, _ := login["refresh_token"].(string)
	require.NotEmpty(t, access)
	require.NotEmpty(t, refresh)

	code, profile := c.do(http.MethodGet, "/api/v1/profile", access, nil)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "ada@example.com", profile["email"])

	code, rotated := c.do(http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{"refresh_token": refresh})
	require.Equal(t, http.StatusOK, code)
	newRefresh, _ := rotated["refresh_token"].(string)
	require.NotEmpty(t, newRefresh)
	require.NotEqual(t, refresh, newRefresh)

	code, _ = c.do(http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{"refresh_token": refresh})
	require.Equal(t, http.StatusUnauthorized, code, "replaying a rotated token must fail")

	code, _ = c.do(http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{"refresh_token": newRefresh})
	require.Equal(t, http.StatusUnauthorized, code, "reuse detection must revoke the whole family")

	code, second := c.do(http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email": "ada@example.com", "password": "SecurePass123!",
	})
	require.Equal(t, http.StatusOK, code)
	secondRefresh, _ := second["refresh_token"].(string)

	code, _ = c.do(http.MethodPost, "/api/v1/logout-all", access, nil)
	require.Equal(t, http.StatusOK, code)

	code, _ = c.do(http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{"refresh_token": secondRefresh})
	require.Equal(t, http.StatusUnauthorized, code, "logout-all must revoke every session")

	code, _ = c.do(http.MethodGet, "/api/v1/profile", "not-a-jwt", nil)
	require.Equal(t, http.StatusUnauthorized, code)
}
