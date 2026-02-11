package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-boilerplate/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBPool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Close()
}

type Database struct {
	Master DBPool
	Slave  DBPool
}

func (d *Database) Close() {
	if d.Master != nil {
		d.Master.Close()
	}
	if d.Slave != nil {
		d.Slave.Close()
	}
}

func Connect(cfg config.DatabaseConfig) *Database {
	master := connectPool(cfg.Master, "Master")
	slave := connectPool(cfg.Slave, "Slave")

	return &Database{
		Master: master,
		Slave:  slave,
	}
}

func connectPool(cfg config.DBConnectionConfig, name string) *pgxpool.Pool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Failed to parse %s database config: %v", name, err)
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
				log.Printf("Successfully connected to %s database", name)
				break
			}
			db.Close()
		}
		log.Printf("Failed to connect to %s database: %v. Retrying in 2 seconds...", name, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to %s database after retries: %v", name, err)
	}

	return db
}
