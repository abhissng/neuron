package rate_limiter

/*
import (
	"net/http"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/gin-contrib/gin-limiter"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

// Redis client
var redisClient *redis.Client

func InitRedisClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     viper.GetString(constant.REDIS_ADDR),
		Password: viper.GetString(constant.REDIS_PASSWORD),
		DB:       constant.REDIS_DB,
	})
}

// RateLimiterMiddleware applies rate limiting to each request based on dynamic rules
func RateLimiterMiddlewareUsingRedis(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Fetch dynamic rate limits from configuration or environment
		rateLimit := getRateLimit(c, path)
		burstLimit := rateLimit * 2 // Arbitrary burst value (can be customized)

		// Use Redis-backed rate limiting
		limiterInstance := limiter.NewRedisLimiter(redisClient, rateLimit, burstLimit)

		// Apply rate limiting to the IP address
		limiterInstance.LimitByIP(c)

		// Proceed if the request passes rate limiting
		c.Next()
	}
}

// getRateLimit returns the dynamic rate limit based on the client's request (IP, path, etc.)
func getRateLimit(c *gin.Context, path string) int {
	// For now, use viper to fetch rate limit dynamically from a configuration
	// You can further customize this logic based on your use case
	rateLimit := viper.GetInt(constant.RATE_LIMIT_DEFAULT)

	// Custom dynamic logic can be added here (e.g., based on path, user role, etc.)
	if c.FullPath() == path { //""/api/special""
		rateLimit = viper.GetInt(constant.RATE_LIMIT_SPECIAL)
	}

	return rateLimit
}

// NewRedisLimiter creates a new Redis-backed rate limiter
func NewRedisLimiter(client *redis.Client, rateLimit, burstLimit int) *limiter.Limiter {
	// Create a new limiter with dynamic rate limit and burst capacity
	limiterInstance := limiter.New(limiter.NewStore(redisClient))
	limiterInstance.SetLimit(rate.Limit(rateLimit), burstLimit)

	return limiterInstance
}

// LimitByIP limits the requests based on the client IP
func (l *limiter.Limiter) LimitByIP(c *gin.Context) {
	ip := c.ClientIP()

	// Apply the rate limit based on the IP address
	limiterKey := "rate_limit:" + ip
	limiterResult := l.AllowIP(context.Background(), limiterKey)

	// If the limit is exceeded, reject the request
	if !limiterResult.Allow {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Rate limit exceeded. Try again later.",
		})
		c.Abort()
		return
	}
}

*/
