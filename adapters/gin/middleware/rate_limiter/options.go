package rate_limiter

import (
	"github.com/redis/go-redis/v9"
)

// RateLimiterOptions holds all configuration for the middleware
type RateLimiterOptions struct {
	RedisClient       *redis.Client  // Required
	EnableIPLimit     bool           // Enable IP-based rate limiting
	EnableRoleLimit   bool           // Enable role-based rate limiting
	WindowSeconds     int64          // Sliding window in seconds
	DefaultIPLimit    int            // Default requests per window per IP
	DefaultRoleLimits map[string]int // Role => limit
}

// Functional option type
type RateLimiterOption func(*RateLimiterOptions)

// WithRedisClient sets the Redis client
func WithRedisClient(client *redis.Client) RateLimiterOption {
	return func(opts *RateLimiterOptions) {
		opts.RedisClient = client
	}
}

// EnableIPRateLimit enables or disables IP-based rate limiting
func EnableIPRateLimit(enable bool) RateLimiterOption {
	return func(opts *RateLimiterOptions) {
		opts.EnableIPLimit = enable
	}
}

// EnableRoleRateLimit enables or disables role-based rate limiting
func EnableRoleRateLimit(enable bool) RateLimiterOption {
	return func(opts *RateLimiterOptions) {
		opts.EnableRoleLimit = enable
	}
}

// WithWindow sets the sliding window in seconds
func WithWindow(seconds int64) RateLimiterOption {
	return func(opts *RateLimiterOptions) {
		opts.WindowSeconds = seconds
	}
}

// WithDefaultIPLimit sets the default IP limit
func WithDefaultIPLimit(limit int) RateLimiterOption {
	return func(opts *RateLimiterOptions) {
		opts.DefaultIPLimit = limit
	}
}

// WithRoleLimits sets the role-based limits
func WithRoleLimits(roleLimits map[string]int) RateLimiterOption {
	return func(opts *RateLimiterOptions) {
		opts.DefaultRoleLimits = roleLimits
	}
}
