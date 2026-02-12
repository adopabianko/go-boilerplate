package container

import (
	"go-boilerplate/internal/config"
	"go-boilerplate/internal/delivery/http/handler"
	grpcgateway "go-boilerplate/internal/gateway/grpc"
	httpgateway "go-boilerplate/internal/gateway/http"
	"go-boilerplate/internal/infrastructure/database"
	"go-boilerplate/internal/infrastructure/redis"
	"go-boilerplate/internal/repository"
	"go-boilerplate/internal/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Container holds all initialised handlers, ready to be consumed by the router.
type Container struct {
	UserHandler    *handler.UserHandler
	HealthHandler  *handler.HealthHandler
	ProductHandler *handler.ProductHandler
	PaymentHandler *handler.PaymentHandler
}

// NewContainer wires repositories → usecases → handlers and returns a ready-to-use Container.
func NewContainer(
	cfg *config.Config,
	db *database.Database,
	rdb *redis.Client,
	mqConn *amqp.Connection,
	productGateway httpgateway.ProductGateway,
	paymentGateway grpcgateway.PaymentGateway,
) *Container {
	// Repositories
	userRepo := repository.NewUserRepository(db)

	// Usecases
	userUsecase := usecase.NewUserUsecase(userRepo, cfg, rdb)
	productUsecase := usecase.NewProductUsecase(productGateway)
	paymentUsecase := usecase.NewPaymentUsecase(paymentGateway)

	// Handlers
	userHandler := handler.NewUserHandler(userUsecase)
	healthHandler := handler.NewHealthHandler(db, rdb, mqConn)
	productHandler := handler.NewProductHandler(productUsecase)
	paymentHandler := handler.NewPaymentHandler(paymentUsecase)

	return &Container{
		UserHandler:    userHandler,
		HealthHandler:  healthHandler,
		ProductHandler: productHandler,
		PaymentHandler: paymentHandler,
	}
}
