package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

/*
Package middleware provides IP-based rate limiting for Gin applications.

# IPRateLimiter Usage

The IPRateLimiter provides per-client (IP address) rate limiting using the token bucket algorithm
from golang.org/x/time/rate. It automatically cleans up inactive client entries based on a TTL.

# Basic Usage

	package main

	import (
		"time"
		"github.com/gin-gonic/gin"
		"golang.org/x/time/rate"
		"your-module/adapters/gin/middleware"
	)

	func main() {
		// Create a rate limiter:
		// - 10 requests per second per IP
		// - Burst of 100 requests
		// - Clean up inactive IPs after 5 minutes
		limiter := middleware.NewIPRateLimiter(
			rate.Limit(10),    // 10 requests/second
			100,               // burst size
			5*time.Minute,     // TTL for inactive clients
		)

		// Important: Stop the cleanup goroutine on shutdown
		defer limiter.StopCleanup()

		router := gin.Default()

		// Apply rate limiting globally
		router.Use(limiter.Middleware())

		// Or apply to specific routes
		api := router.Group("/api")
		api.Use(limiter.Middleware())
		{
			api.GET("/resource", handleResource)
		}

		router.Run(":8080")
	}

# Recommended Default Values

For typical API servers:
  - Rate: 10-100 requests per second (rate.Limit(10) to rate.Limit(100))
  - Burst: 2x to 10x the rate (e.g., 100 for rate of 10/sec)
  - TTL: 5-10 minutes (5*time.Minute to 10*time.Minute)

For stricter rate limiting (e.g., login endpoints):
  - Rate: 1-5 requests per second (rate.Limit(1) to rate.Limit(5))
  - Burst: 5-10 requests
  - TTL: 15 minutes

# Lifecycle Management

Always call StopCleanup() when shutting down to prevent goroutine leaks:

	limiter := middleware.NewIPRateLimiter(rate.Limit(10), 100, 5*time.Minute)
	defer limiter.StopCleanup()

	// Or with graceful shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	limiter.StopCleanup()  // Stop cleanup before shutting down server

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

# How It Works

1. Each client IP gets its own rate.Limiter instance
2. The limiter uses a token bucket algorithm: tokens are generated at the specified rate,
   up to the burst limit
3. Each request consumes one token; if no tokens are available, the request is rejected with 429
4. A background goroutine periodically removes limiters for IPs that haven't made requests
   within the TTL period
5. The lastSeen timestamp is updated on every request to track activity

# Response on Rate Limit

When rate limit is exceeded, clients receive:
  - HTTP Status: 429 Too Many Requests
  - Response Body: {"error": "Too many requests"}
*/

// clientLimiter holds the limiter and the last seen time for a client
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter manages rate limiters for all clients (IPs)
type IPRateLimiter struct {
	clients  map[string]*clientLimiter
	mu       *sync.Mutex
	rate     rate.Limit    // The rate of token generation (e.g., 10 requests per second)
	burst    int           // The maximum burst size (e.g., 100 requests)
	ttl      time.Duration // Time-to-live for inactive client entries
	stop     chan struct{}
	stopOnce sync.Once
}

// NewIPRateLimiter creates a new rate limiter manager.
// r: The number of events allowed per second.
// b: The burst size (how many requests can be made in a short burst).
// ttl: How long to keep an IP's limiter in memory after its last request.
func NewIPRateLimiter(r rate.Limit, b int, ttl time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients: make(map[string]*clientLimiter),
		mu:      &sync.Mutex{},
		rate:    r,
		burst:   b,
		ttl:     ttl,
		stop:    make(chan struct{}),
	}

	// Start a background goroutine to clean up old entries
	go limiter.cleanupClients()

	return limiter
}

// getLimiter retrieves or creates a limiter for a given IP address.
func (l *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	client, exists := l.clients[ip]
	if !exists {
		// Create a new limiter for this IP
		client = &clientLimiter{
			limiter: rate.NewLimiter(l.rate, l.burst),
		}
		l.clients[ip] = client
	}

	// Update the last seen time
	client.lastSeen = time.Now()
	return client.limiter
}

// cleanupClients periodically removes limiters for inactive IPs.
func (l *IPRateLimiter) cleanupClients() {
	// Use a minimum interval to prevent excessive cleanup frequency
	interval := l.ttl / 2
	if interval < time.Minute {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-l.stop:
			return
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						helpers.Println(constant.ERROR, "exception: occurred in cleanupClients", "stack:", string(debug.Stack()))
						// Log the panic but continue cleanup loop
					}
				}()
				// Collect IPs to delete without holding lock for entire iteration
				var toDelete []string
				now := time.Now()
				l.mu.Lock()
				for ip, client := range l.clients {
					if now.Sub(client.lastSeen) > l.ttl {
						toDelete = append(toDelete, ip)
					}
				}
				for _, ip := range toDelete {
					delete(l.clients, ip)
				}
				l.mu.Unlock()
			}()
		}
	}
}

// StopCleanup stops the cleanup goroutine.
func (l *IPRateLimiter) StopCleanup() {
	l.stopOnce.Do(func() {
		close(l.stop)
	})
}

// Middleware returns the Gin middleware handler.
func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the real client IP from RemoteAddr (not spoofable)
		// If behind a trusted proxy, configure Gin's TrustedProxies instead
		ip := c.Request.RemoteAddr
		// Strip port if present
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		limiter := l.getLimiter(ip)

		// Check if the request is allowed
		if !limiter.Allow() {
			// Calculate retry-after based on rate limit
			retryAfter := int(time.Second / time.Duration(l.rate))
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", l.burst))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		c.Next()
	}
}
