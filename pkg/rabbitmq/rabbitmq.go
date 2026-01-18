package rabbitmq

import (
	"log"

	"go-boilerplate/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(cfg config.RabbitMQConfig) *amqp.Connection {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	return conn
}
