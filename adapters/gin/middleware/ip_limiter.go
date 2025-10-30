package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// clientLimiter holds the limiter and the last seen time for a client
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter manages rate limiters for all clients (IPs)
type IPRateLimiter struct {
	clients map[string]*clientLimiter
	mu      *sync.Mutex
	rate    rate.Limit    // The rate of token generation (e.g., 10 requests per second)
	burst   int           // The maximum burst size (e.g., 100 requests)
	ttl     time.Duration // Time-to-live for inactive client entries
	stop    chan struct{}
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
	// Check for inactive clients every (ttl / 2)
	// Or a fixed interval, like every minute.
	ticker := time.NewTicker(l.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-l.stop:
			return // Stop the goroutine
		case <-ticker.C:
			l.mu.Lock()
			for ip, client := range l.clients {
				if time.Since(client.lastSeen) > l.ttl {
					delete(l.clients, ip)
				}
			}
			l.mu.Unlock()
		}
	}
}

// StopCleanup stops the cleanup goroutine.
func (l *IPRateLimiter) StopCleanup() {
	close(l.stop)
}

// Middleware returns the Gin middleware handler.
func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the limiter for the specific client IP
		limiter := l.getLimiter(c.ClientIP())

		// Check if the request is allowed
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		c.Next()
	}
}
