package config

import (
	"fmt"

	"github.com/bonarizki/gin-boilerplate/pkg/logger"
)

type ServerConfiguration struct {
	Host                   string
	Port                   string
	JWTSecret              string
	JWTIssuer              string // written into and required on access tokens
	JWTAudience            string // written into and required on access tokens
	Debug                  bool
	Timezone               string // IANA name, e.g. UTC, Asia/Jakarta
	RequestTimeoutSeconds  int    // http.Server read/write timeout
	ShutdownTimeoutSeconds int    // graceful shutdown max wait
	RefreshTokenTTLDays    int    // refresh token lifetime before it must be re-issued via login
	AccessTokenTTLMinutes  int    // JWT access token lifetime; kept short since it cannot be revoked
	// Rate limit (used by middlewares)
	RateLimitRPS     int
	RateLimitBurst   int
	RateLimitUseUser bool
	// AuthRateLimitRPM / AuthRateLimitBurst apply stricter per-IP limits on /auth routes only.
	AuthRateLimitRPM   int
	AuthRateLimitBurst int
	// CORSAllowedOrigins is the allowlist of exact frontend origins permitted
	// to make cross-origin requests. Never contains "*".
	CORSAllowedOrigins []string
	// TrustedProxies is the list of IPs/CIDRs trusted to set X-Forwarded-For,
	// passed to gin.Engine.SetTrustedProxies(). Unrelated to CORS.
	TrustedProxies []string
}

// ServerConfig returns the server address (host:port) from loaded config.
// Call after SetupConfig(). Returns default "0.0.0.0:8000" if config not loaded.
func ServerConfig() string {
	if c := Get(); c != nil {
		addr := fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
		logger.Infof("server listening at %s", addr)
		return addr
	}
	logger.Infof("server listening at 0.0.0.0:8000 (config not loaded)")
	return "0.0.0.0:8000"
}
