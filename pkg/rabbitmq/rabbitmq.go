package rabbitmq

import (
	"log"
	"time"

	"go-boilerplate/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(cfg config.RabbitMQConfig) *amqp.Connection {
	var conn *amqp.Connection
	var err error

	for i := 0; i < 30; i++ {
		conn, err = amqp.Dial(cfg.URL)
		if err == nil {
			return conn
		}
		log.Printf("Failed to connect to RabbitMQ: %v. Retrying in 2 seconds...", err)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("Failed to connect to RabbitMQ after retries: %v", err)
	return nil
}
