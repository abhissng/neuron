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
)

// RateLimiterMiddleware applies rate limiting to each request based on dynamic rules
func RateLimiterMiddleware(rateLimit time.Duration) gin.HandlerFunc {
	if rateLimit == 0 || viper.GetDuration(constant.RateLimitDurationInSecondKey) == 0 { // Limit each client to one request every 10 seconds
		rateLimit = time.Second * 10
	}
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
