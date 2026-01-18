package http

import (
	"net/http"

	"go-boilerplate/docs"
	"go-boilerplate/internal/config"
	"go-boilerplate/internal/delivery/http/handler"
	"go-boilerplate/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(
	cfg *config.Config,
	userHandler *handler.UserHandler,
	healthHandler *handler.HealthHandler,
	productHandler *handler.ProductHandler,
	paymentHandler *handler.PaymentHandler,
) *gin.Engine {
	// Gin Mode
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// Middlewares
	r.Use(gin.Logger())
	r.Use(middleware.RecoveryMiddleware())

	// Swagger
	docs.SwaggerInfo.BasePath = "/" // Fix fetch error
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	// Health Check
	r.GET("/health", healthHandler.Check)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		user := api.Group("/users")
		user.Use(middleware.AuthMiddleware(cfg.JWT))
		{
			user.GET("", userHandler.ListUsers) // GET /api/v1/users
			user.GET("/me", func(c *gin.Context) {
				// Example protected route
				userID, _ := c.Get("userID")
				c.JSON(http.StatusOK, gin.H{"user_id": userID})
			})
		}

		product := api.Group("/products")
		// Optional: Add AuthMiddleware if needed
		{
			product.GET("", productHandler.ListProducts)
		}

		// Payment (gRPC)
		payment := api.Group("/payments")
		{
			payment.GET("/:id", paymentHandler.CheckStatus)
		}
	}

	return r
}
