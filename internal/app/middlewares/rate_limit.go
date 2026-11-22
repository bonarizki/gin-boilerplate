package middlewares

import (
	"fmt"
	"sync"
	"time"

	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	rateLimiterMaxKeys  = 10_000
	rateLimiterIdleTTL  = 15 * time.Minute
	rateLimiterCleanup  = 5 * time.Minute
)

// IPRateLimiter manages rate limiters per IP address
type IPRateLimiter struct {
	ips      map[string]*rate.Limiter
	lastSeen map[string]time.Time
	mu       sync.Mutex
	r        rate.Limit
	b        int
	maxKeys  int
	cleanup  sync.Once
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:      make(map[string]*rate.Limiter),
		lastSeen: make(map[string]time.Time),
		r:        r,
		b:        b,
		maxKeys:  rateLimiterMaxKeys,
	}
}

func (i *IPRateLimiter) ensureCleanup() {
	i.cleanup.Do(func() {
		go func() {
			ticker := time.NewTicker(rateLimiterCleanup)
			defer ticker.Stop()
			for range ticker.C {
				i.CleanupOldLimiters()
				logger.Debugf("rate limiter cleanup completed")
			}
		}()
	})
}

// GetLimiter returns the rate limiter for the given IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.ensureCleanup()
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	i.lastSeen[ip] = now

	limiter, exists := i.ips[ip]
	if !exists {
		if len(i.ips) >= i.maxKeys {
			i.evictStaleLocked(now)
		}
		if len(i.ips) >= i.maxKeys {
			i.evictOldestLocked()
		}
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

func (i *IPRateLimiter) evictStaleLocked(now time.Time) {
	for key, seen := range i.lastSeen {
		if now.Sub(seen) > rateLimiterIdleTTL {
			delete(i.ips, key)
			delete(i.lastSeen, key)
		}
	}
}

func (i *IPRateLimiter) evictOldestLocked() {
	var oldestKey string
	var oldest time.Time
	first := true
	for key, seen := range i.lastSeen {
		if first || seen.Before(oldest) {
			oldestKey = key
			oldest = seen
			first = false
		}
	}
	if oldestKey != "" {
		delete(i.ips, oldestKey)
		delete(i.lastSeen, oldestKey)
	}
}

// CleanupOldLimiters removes limiters that haven't been used recently
func (i *IPRateLimiter) CleanupOldLimiters() {
	i.mu.Lock()
	defer i.mu.Unlock()
	now := time.Now()
	for key, seen := range i.lastSeen {
		if now.Sub(seen) > rateLimiterIdleTTL {
			delete(i.ips, key)
			delete(i.lastSeen, key)
			continue
		}
		if lim, ok := i.ips[key]; ok && lim.Tokens() == float64(i.b) && now.Sub(seen) > rateLimiterCleanup {
			delete(i.ips, key)
			delete(i.lastSeen, key)
		}
	}
}

var (
	globalLimiter *IPRateLimiter
	once          sync.Once
)

// initRateLimiter initializes the global rate limiter from config (config.Get().Server).
func initRateLimiter() {
	once.Do(func() {
		rps, burst := 100, 200
		if c := config.Get(); c != nil {
			if c.Server.RateLimitRPS > 0 {
				rps = c.Server.RateLimitRPS
			}
			if c.Server.RateLimitBurst > 0 {
				burst = c.Server.RateLimitBurst
			}
		}
		globalLimiter = NewIPRateLimiter(rate.Limit(rps), burst)
	})
}

// rateLimitKey returns the key for rate limiting: user ID if RateLimitUseUser is true
// and user_id is set in context (e.g. by AuthMiddleware), otherwise client IP.
func rateLimitKey(c *gin.Context) string {
	useUser := false
	if cfg := config.Get(); cfg != nil {
		useUser = cfg.Server.RateLimitUseUser
	}
	if useUser {
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(uint); ok {
				return fmt.Sprintf("user:%d", id)
			}
		}
	}
	return c.ClientIP()
}

// RateLimitMiddleware creates a rate limiting middleware.
func RateLimitMiddleware() gin.HandlerFunc {
	initRateLimiter()

	return func(c *gin.Context) {
		key := rateLimitKey(c)
		limiter := globalLimiter.GetLimiter(key)

		if !limiter.Allow() {
			logger.Warnf("rate limit exceeded for key: %s", key)
			utils.TooManyRequests(c, nil, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

var (
	authLimiterOnce sync.Once
	authLimiter     *IPRateLimiter
)

func authRateLimiter() *IPRateLimiter {
	authLimiterOnce.Do(func() {
		rpm, burst := 5, 5
		if c := config.Get(); c != nil {
			if c.Server.AuthRateLimitRPM > 0 {
				rpm = c.Server.AuthRateLimitRPM
			}
			if c.Server.AuthRateLimitBurst > 0 {
				burst = c.Server.AuthRateLimitBurst
			}
		}
		limit := rate.Limit(float64(rpm) / 60.0)
		authLimiter = NewIPRateLimiter(limit, burst)
	})
	return authLimiter
}

// AuthRateLimitMiddleware applies a stricter per-IP limit to authentication routes.
func AuthRateLimitMiddleware() gin.HandlerFunc {
	limiter := authRateLimiter()
	return rateLimitMiddlewareWithLimiter(limiter)
}

func rateLimitMiddlewareWithLimiter(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		ipLimiter := limiter.GetLimiter(key)

		if !ipLimiter.Allow() {
			logger.Warnf("auth rate limit exceeded for key: %s", key)
			utils.TooManyRequests(c, nil, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddlewareWithConfig creates a rate limiting middleware with custom config
func RateLimitMiddlewareWithConfig(requestsPerSecond int, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(requestsPerSecond), burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		ipLimiter := limiter.GetLimiter(ip)

		if !ipLimiter.Allow() {
			logger.Warnf("rate limit exceeded for IP: %s (limit: %d/s, burst: %d)", ip, requestsPerSecond, burst)
			utils.TooManyRequests(c, nil, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}