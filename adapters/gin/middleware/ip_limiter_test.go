package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestNewIPRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 100, 5*time.Minute)
	defer limiter.StopCleanup()

	assert.NotNil(t, limiter)
	assert.NotNil(t, limiter.clients)
	assert.NotNil(t, limiter.mu)
	assert.Equal(t, rate.Limit(10), limiter.rate)
	assert.Equal(t, 100, limiter.burst)
	assert.Equal(t, 5*time.Minute, limiter.ttl)
	assert.NotNil(t, limiter.stop)
}

func TestGetLimiter_NewIP(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 100, 5*time.Minute)
	defer limiter.StopCleanup()

	ip := "192.168.1.1"
	rl := limiter.getLimiter(ip)

	assert.NotNil(t, rl)
	assert.Len(t, limiter.clients, 1)
	assert.Contains(t, limiter.clients, ip)
}

func TestGetLimiter_ExistingIP(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 100, 5*time.Minute)
	defer limiter.StopCleanup()

	ip := "192.168.1.1"
	rl1 := limiter.getLimiter(ip)
	rl2 := limiter.getLimiter(ip)

	assert.Equal(t, rl1, rl2, "Should return the same limiter for the same IP")
	assert.Len(t, limiter.clients, 1, "Should not create duplicate entries")
}

func TestGetLimiter_UpdatesLastSeen(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 100, 5*time.Minute)
	defer limiter.StopCleanup()

	ip := "192.168.1.1"
	limiter.getLimiter(ip)

	firstSeen := limiter.clients[ip].lastSeen
	time.Sleep(100 * time.Millisecond)

	limiter.getLimiter(ip)
	secondSeen := limiter.clients[ip].lastSeen

	assert.True(t, secondSeen.After(firstSeen), "lastSeen should be updated")
}

func TestRateLimiter_AllowBehavior(t *testing.T) {
	// Create a very restrictive limiter: 1 request per second, burst of 1
	limiter := NewIPRateLimiter(rate.Limit(1), 1, 5*time.Minute)
	defer limiter.StopCleanup()

	ip := "192.168.1.1"
	rl := limiter.getLimiter(ip)

	// First request should be allowed (uses burst token)
	assert.True(t, rl.Allow(), "First request should be allowed")

	// Second immediate request should be denied (no tokens available)
	assert.False(t, rl.Allow(), "Second immediate request should be denied")

	// Wait for token replenishment (1 second + buffer)
	time.Sleep(1100 * time.Millisecond)

	// Request should now be allowed
	assert.True(t, rl.Allow(), "Request after waiting should be allowed")
}

func TestConcurrentAccess_SameIP(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(100), 200, 5*time.Minute)
	defer limiter.StopCleanup()

	ip := "192.168.1.1"
	concurrency := 50
	var wg sync.WaitGroup

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			rl := limiter.getLimiter(ip)
			assert.NotNil(t, rl)
		}()
	}

	wg.Wait()

	// Should only have one entry for the IP
	assert.Len(t, limiter.clients, 1)
}

func TestConcurrentAccess_DifferentIPs(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 100, 5*time.Minute)
	defer limiter.StopCleanup()

	concurrency := 100
	var wg sync.WaitGroup

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			ip := "192.168.1." + string(rune(idx))
			rl := limiter.getLimiter(ip)
			assert.NotNil(t, rl)
		}(i)
	}

	wg.Wait()

	// Should have entries for all different IPs
	assert.Equal(t, concurrency, len(limiter.clients))
}

func TestCleanupClients_RemovesInactiveEntries(t *testing.T) {
	// Use a very short TTL for testing
	ttl := 500 * time.Millisecond
	limiter := NewIPRateLimiter(rate.Limit(10), 100, ttl)
	defer limiter.StopCleanup()

	// Add some clients
	limiter.getLimiter("192.168.1.1")
	limiter.getLimiter("192.168.1.2")
	limiter.getLimiter("192.168.1.3")

	assert.Len(t, limiter.clients, 3, "Should have 3 clients initially")

	// Wait for TTL to expire and cleanup to run
	// Cleanup runs every ttl/2, so wait ttl + buffer
	time.Sleep(ttl + 300*time.Millisecond)

	limiter.mu.Lock()
	clientCount := len(limiter.clients)
	limiter.mu.Unlock()

	assert.Equal(t, 0, clientCount, "All inactive clients should be cleaned up")
}

func TestCleanupClients_KeepsActiveEntries(t *testing.T) {
	// Use a short TTL for testing
	ttl := 1 * time.Second
	limiter := NewIPRateLimiter(rate.Limit(10), 100, ttl)
	defer limiter.StopCleanup()

	ip := "192.168.1.1"
	limiter.getLimiter(ip)

	// Keep the client active by accessing it
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		for i := 0; i < 4; i++ {
			<-ticker.C
			limiter.getLimiter(ip)
		}
		done <- true
	}()

	<-done

	limiter.mu.Lock()
	clientCount := len(limiter.clients)
	limiter.mu.Unlock()

	assert.Equal(t, 1, clientCount, "Active client should not be cleaned up")
}

func TestStopCleanup(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 100, 5*time.Minute)

	// Add a client
	limiter.getLimiter("192.168.1.1")

	// Stop cleanup
	limiter.StopCleanup()

	// Give some time for the goroutine to exit
	time.Sleep(100 * time.Millisecond)

	// Channel should be closed (attempting to close again would panic)
	// We can verify by checking if we can read from the closed channel
	select {
	case <-limiter.stop:
		// Successfully read from closed channel
	default:
		t.Error("Stop channel should be closed")
	}
}

func TestMiddleware_AllowsNormalRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewIPRateLimiter(rate.Limit(10), 100, 5*time.Minute)
	defer limiter.StopCleanup()

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make a normal request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_Returns429WhenRateLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a very restrictive limiter: 1 request per second, burst of 2
	limiter := NewIPRateLimiter(rate.Limit(1), 2, 5*time.Minute)
	defer limiter.StopCleanup()

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	ip := "192.168.1.1:1234"

	// First two requests should succeed (burst allows 2)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// Third request should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = ip
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests")
}

func TestMiddleware_DifferentIPsHaveSeparateLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a very restrictive limiter: 1 request per second, burst of 1
	limiter := NewIPRateLimiter(rate.Limit(1), 1, 5*time.Minute)
	defer limiter.StopCleanup()

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First IP - first request should succeed
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Different IP - should also succeed (separate limiter)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func BenchmarkMiddleware_SingleIP(b *testing.B) {
	gin.SetMode(gin.TestMode)

	limiter := NewIPRateLimiter(rate.Limit(1000000), 1000000, 5*time.Minute)
	defer limiter.StopCleanup()

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_MultipleIPs(b *testing.B) {
	gin.SetMode(gin.TestMode)

	limiter := NewIPRateLimiter(rate.Limit(1000000), 1000000, 5*time.Minute)
	defer limiter.StopCleanup()

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		// Simulate different IPs
		req.RemoteAddr = "192.168.1." + string(rune(i%255)) + ":1234"
		router.ServeHTTP(w, req)
	}
}
