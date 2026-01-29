package middleware

import (
	"go-boilerplate/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	c := cors.DefaultConfig()
	if len(cfg.AllowedOrigins) > 0 && cfg.AllowedOrigins[0] != "*" {
		c.AllowOrigins = cfg.AllowedOrigins
	} else {
		c.AllowAllOrigins = true
	}

	c.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	c.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After"}

	return cors.New(c)
}
