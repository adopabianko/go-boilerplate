package redis

import (
	"context"
	"fmt"
	"log"

	"go-boilerplate/internal/config"

	"github.com/redis/go-redis/v9"
)

func Connect(cfg config.RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	return rdb
}
