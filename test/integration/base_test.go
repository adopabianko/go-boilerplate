package integration

import (
	"context"
	"fmt"
	"go-boilerplate/internal/config"
	"go-boilerplate/internal/infrastructure/database"
	"go-boilerplate/internal/infrastructure/redis"
	"go-boilerplate/pkg/logger"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

var (
	cfg *config.Config
	db  *database.Database
	rdb *redis.Client
)

func TestMain(m *testing.M) {
	// Load .env from root
	godotenv.Load("../../.env")

	// Load config
	cfg = config.LoadConfig()

	// Initialize Logger
	logger.InitLogger(cfg)

	// Connect to Database
	db = database.Connect(cfg.Database)
	// For integration tests, we use Master for both to avoid replication lag
	db.Slave = db.Master

	// Connect to Redis
	rdb = redis.Connect(cfg.Redis)
	
	// Override JWT cert paths to be relative to test/integration
	cfg.JWT.PrivateKeyPath = "../../" + cfg.JWT.PrivateKeyPath
	cfg.JWT.PublicKeyPath = "../../" + cfg.JWT.PublicKeyPath

	// Run tests
	code := m.Run()

	// Cleanup
	db.Close()
	rdb.Close()

	os.Exit(code)
}

func truncateTables() {
	log.Println("--- TRUNCATE TABLES CALLED ---")
	tables := []string{"users"}
	ctx := context.Background()
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		_, err := db.Master.Exec(ctx, query)
		if err != nil {
			log.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}

	// Give some time for replication to sync
	time.Sleep(1 * time.Second)

	// Flush Redis
	err := rdb.FlushDB(ctx).Err()
	if err != nil {
		log.Fatalf("Failed to flush redis: %v", err)
	}
}
