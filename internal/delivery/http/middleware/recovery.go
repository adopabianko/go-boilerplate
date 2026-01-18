package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go-boilerplate/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				stack := string(debug.Stack())
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", stack),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Internal Server Error: %v", err),
				})
			}
		}()
		c.Next()
	}
}
