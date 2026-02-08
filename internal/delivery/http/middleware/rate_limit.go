package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-boilerplate/internal/config"
	"go-boilerplate/internal/infrastructure/redis"
	"go-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(rdb *redis.Client, cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		// Increment the counter
		count, err := rdb.Incr(c.Request.Context(), key).Result()
		if err != nil {
			// Fail open if Redis is down, or log and fail closed depending on requirements
			// For now, let's log and proceed to avoid blocking users due to infra issues
			// In a strict environment, might want to return 500
			c.Next()
			return
		}

		// If this is the first request, set expiration
		if count == 1 {
			rdb.Expire(c.Request.Context(), key, time.Duration(cfg.Window)*time.Second)
		}

		// Check limit
		if count > int64(cfg.Limit) {
			// Get TTL to set Retry-After header
			ttl, _ := rdb.TTL(c.Request.Context(), key).Result()

			c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(ttl.Seconds()), 10))

			response.Error(c, fmt.Errorf("too many requests"))
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Limit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(cfg.Limit)-count, 10))

		c.Next()
	}
}
