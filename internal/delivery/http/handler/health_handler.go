package handler

import (
	"net/http"
	"time"

	"go-boilerplate/internal/infrastructure/database"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db       *database.Database
	rdb      *redis.Client
	rabbitmq *amqp.Connection
}

func NewHealthHandler(db *database.Database, rdb *redis.Client, rabbitmq *amqp.Connection) *HealthHandler {
	return &HealthHandler{
		db:       db,
		rdb:      rdb,
		rabbitmq: rabbitmq,
	}
}

type HealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Services  map[string]ServiceStatus `json:"services"`
}

type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Check godoc
// @Summary      Health Check
// @Description  Get the health status of the application and its dependencies
// @Tags         health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      503  {object}  HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	response := HealthResponse{
		Status:    "up",
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceStatus),
	}
	statusCode := http.StatusOK
	ctx := c.Request.Context()

	// Check postgres master
	if err := h.db.Master.Ping(ctx); err != nil {
		response.Services["postgres_master"] = ServiceStatus{Status: "down", Message: "Connection failed"}
		response.Status = "down"
		statusCode = http.StatusServiceUnavailable
	} else {
		response.Services["postgres_master"] = ServiceStatus{Status: "up"}
	}

	// Check postgres slave
	if err := h.db.Slave.Ping(ctx); err != nil {
		response.Services["postgres_slave"] = ServiceStatus{Status: "down", Message: "Connection failed"}
		response.Status = "down"
		statusCode = http.StatusServiceUnavailable
	} else {
		response.Services["postgres_slave"] = ServiceStatus{Status: "up"}
	}

	// Check Redis
	if _, err := h.rdb.Ping(c.Request.Context()).Result(); err != nil {
		response.Services["redis"] = ServiceStatus{Status: "down", Message: err.Error()}
		response.Status = "down"
		statusCode = http.StatusServiceUnavailable
	} else {
		response.Services["redis"] = ServiceStatus{Status: "up"}
	}

	// Check RabbitMQ
	if h.rabbitmq.IsClosed() {
		response.Services["rabbitmq"] = ServiceStatus{Status: "down", Message: "Connection closed"}
		response.Status = "down"
		statusCode = http.StatusServiceUnavailable
	} else {
		response.Services["rabbitmq"] = ServiceStatus{Status: "up"}
	}

	c.JSON(statusCode, response)
}
