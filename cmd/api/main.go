package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "go-boilerplate/pkg/pb/payment"

	"go-boilerplate/internal/config"
	httpDelivery "go-boilerplate/internal/delivery/http"
	"go-boilerplate/internal/delivery/http/handler"
	grpcgateway "go-boilerplate/internal/gateway/grpc"
	httpgateway "go-boilerplate/internal/gateway/http"
	"go-boilerplate/internal/repository"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/database"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/rabbitmq"
	"go-boilerplate/pkg/redis"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// @title           Go Boilerplate API
// @version         1.0
// @description     This is a sample server for Go Boilerplate.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load Config
	cfg := config.LoadConfig()

	// Initialize Logger
	logger.InitLogger(cfg)
	defer logger.Log.Sync()

	// Initialize Database
	db := database.Connect(cfg.Database)

	// Initialize Redis
	rdb := redis.Connect(cfg.Redis)

	// Initialize RabbitMQ
	mqConn := rabbitmq.Connect(cfg.RabbitMQ)
	defer mqConn.Close()

	// Initialize Minio (Optional, just ensuring it connects)
	// minioClient := minio.Connect(cfg.Minio)

	// Initialize Repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize Gateways
	productGateway := httpgateway.NewProductGateway(cfg.External.ProductAPIURL)

	// Payment gRPC Connection
	// NOTE: In a real app, use a real address. connecting to localhost:50051 (dummy)
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}
	defer conn.Close()
	paymentClient := pb.NewPaymentServiceClient(conn)
	paymentGateway := grpcgateway.NewPaymentGateway(paymentClient)

	// Initialize Usecases
	userUsecase := usecase.NewUserUsecase(userRepo, cfg, rdb)
	productUsecase := usecase.NewProductUsecase(productGateway)
	paymentUsecase := usecase.NewPaymentUsecase(paymentGateway)

	// Initialize Handlers
	userHandler := handler.NewUserHandler(userUsecase)
	healthHandler := handler.NewHealthHandler(db, rdb, mqConn)
	productHandler := handler.NewProductHandler(productUsecase)
	paymentHandler := handler.NewPaymentHandler(paymentUsecase)

	// Initialize Router
	router := httpDelivery.NewRouter(cfg, userHandler, healthHandler, productHandler, paymentHandler)

	// Start Server
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// Close Database Connection
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get SQL DB for closing", zap.Error(err))
	} else {
		if err := sqlDB.Close(); err != nil {
			logger.Error("Failed to close database connection", zap.Error(err))
		} else {
			log.Println("Database connection closed")
		}
	}

	log.Println("Server exiting")
}
