package middlewares_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bonarizki/gin-boilerplate/internal/app/middlewares"
	"github.com/bonarizki/gin-boilerplate/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_AuthorizationHeader(t *testing.T) {
	svc := &mocks.MockAuthServicer{
		ValidateTokenFunc: func(token string) (uint, error) {
			if token == "good" {
				return 42, nil
			}
			return 0, errors.New("bad token")
		},
	}
	router := setupTestRouter()
	router.Use(middlewares.AuthMiddleware(svc))
	router.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetUint("user_id")})
	})

	cases := []struct {
		header string
		want   int
	}{
		{"Bearer good", http.StatusOK},
		{"bearer good", http.StatusOK},
		{"BEARER  good ", http.StatusOK},
		{"", http.StatusUnauthorized},
		{"Bearer", http.StatusUnauthorized},
		{"Bearer ", http.StatusUnauthorized},
		{"Basic good", http.StatusUnauthorized},
		{"Bearer bad", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.header, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.want, w.Code)
			if tc.want == http.StatusOK {
				assert.JSONEq(t, `{"user_id":42}`, w.Body.String())
			}
		})
	}
}
