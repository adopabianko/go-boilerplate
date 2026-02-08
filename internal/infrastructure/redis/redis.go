package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-boilerplate/internal/config"

	"github.com/redis/go-redis/v9"
)

type Client = redis.Client

func Connect(cfg config.RedisConfig) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	var err error
	for i := 0; i < 30; i++ {
		if err = rdb.Ping(context.Background()).Err(); err == nil {
			return rdb
		}
		log.Printf("Failed to connect to redis: %v. Retrying in 2 seconds...", err)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("Failed to connect to redis after retries: %v", err)
	return nil
}
