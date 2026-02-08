package middleware

import (
	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmgin/v2"
)

// APMMiddleware returns a Gin middleware that traces all HTTP requests using Elastic APM
func APMMiddleware(engine *gin.Engine) gin.HandlerFunc {
	return apmgin.Middleware(engine)
}
