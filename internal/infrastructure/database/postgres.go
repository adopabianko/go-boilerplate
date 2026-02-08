package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-boilerplate/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(cfg config.DatabaseConfig) *pgxpool.Pool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime

	// Add query tracer for logging SQL queries to ELK
	poolConfig.ConnConfig.Tracer = &QueryTracer{}

	var db *pgxpool.Pool
	ctx := context.Background()

	// Retry connection loop
	for i := 0; i < 30; i++ {
		db, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			if err = db.Ping(ctx); err == nil {
				break
			}
			db.Close()
		}
		log.Printf("Failed to connect to database: %v. Retrying in 2 seconds...", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}

	return db
}
