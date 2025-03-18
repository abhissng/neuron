package rate_limiter

import (
	"sync"
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	clients      = make(map[string]time.Time)
	clientsMutex sync.Mutex
	rateLimit    = viper.GetDuration(constant.RateLimitDurationInSecondKey) * time.Second // Limit each client to one request every 10 seconds
)

// RateLimiterMiddleware applies rate limiting to each request based on dynamic rules
func RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		clientsMutex.Lock()
		lastRequest, exists := clients[clientIP]
		if exists && time.Since(lastRequest) < rateLimit {
			clientsMutex.Unlock()
			c.AbortWithStatusJSON(429, gin.H{"error": "Too many requests"})
			return
		}
		clients[clientIP] = time.Now()
		clientsMutex.Unlock()
		c.Next()
	}
}
