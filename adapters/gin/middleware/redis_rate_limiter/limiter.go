package rate_limiter

import (
	"fmt"
	"net/http"
	"time"

	"context"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiterMiddleware returns a Gin middleware configured via RateLimiterOptions
func (opts *RateLimiterOptions) RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		now := time.Now().Unix()
		window := opts.WindowSeconds
		if window <= 0 {
			window = 60
		}

		exceeded := false
		reason := ""

		// ===== IP Rate Limit =====
		if opts.EnableIPLimit {
			ipKey := fmt.Sprintf("rate_limit:ip:%s:%s", c.ClientIP(), c.FullPath())
			ipLimit := opts.DefaultIPLimit
			if ipLimit <= 0 {
				ipLimit = 20
			}

			opts.RedisClient.ZRemRangeByScore(ctx, ipKey, "0", fmt.Sprint(now-window))
			opts.RedisClient.ZAdd(ctx, ipKey, redis.Z{
				Score:  float64(now),
				Member: now,
			})
			count, _ := opts.RedisClient.ZCard(ctx, ipKey).Result()
			opts.RedisClient.Expire(ctx, ipKey, time.Duration(window)*time.Second)

			if count > int64(ipLimit) {
				exceeded = true
				reason = "IP rate limit exceeded"
			}
		}

		// ===== Role Rate Limit =====
		if opts.EnableRoleLimit && !exceeded {
			role := getUserRole(c)
			roleLimit := getRoleLimit(opts.DefaultRoleLimits, role)
			roleKey := fmt.Sprintf("rate_limit:role:%s:%s:%s", c.ClientIP(), role, c.FullPath())

			opts.RedisClient.ZRemRangeByScore(ctx, roleKey, "0", fmt.Sprint(now-window))
			opts.RedisClient.ZAdd(ctx, roleKey, redis.Z{
				Score:  float64(now),
				Member: now,
			})
			count, _ := opts.RedisClient.ZCard(ctx, roleKey).Result()
			opts.RedisClient.Expire(ctx, roleKey, time.Duration(window)*time.Second)

			if count > int64(roleLimit) {
				exceeded = true
				reason = fmt.Sprintf("Role rate limit exceeded for role %s", role)
			}
		}

		if exceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": reason})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ===== Helper to get role from header or JWT =====
func getUserRole(c *gin.Context) string {
	role := c.GetHeader(constant.XUserRole)
	if role == "" {
		role = "guest"
	}
	return role
}

// ===== Get role limit from map =====
func getRoleLimit(roleLimits map[string]int, role string) int {
	if roleLimits == nil {
		return 10
	}
	if val, ok := roleLimits[role]; ok {
		return val
	}
	return 10
}
