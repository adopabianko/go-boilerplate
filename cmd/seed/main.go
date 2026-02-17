package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"go-boilerplate/internal/config"
	"go-boilerplate/internal/infrastructure/database"
	"go-boilerplate/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Initialize Logger
	logger.InitLogger(&logger.LogstashConfig{
		Host: cfg.Logstash.Host,
		Port: cfg.Logstash.Port,
	})

	// Connect to Database
	db := database.Connect(cfg.Database)
	defer db.Close()

	ctx := context.Background()

	// Path to seeder directory
	seederDir := "migrations/seeder"

	files, err := os.ReadDir(seederDir)
	if err != nil {
		log.Fatalf("Failed to read seeder directory: %v", err)
	}

	logger.Info("Running SQL seeders...")

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		logger.Info("Executing seeder file", zap.String("file", file.Name()))

		content, err := os.ReadFile(filepath.Join(seederDir, file.Name()))
		if err != nil {
			logger.Error("Failed to read seeder file", zap.String("file", file.Name()), zap.Error(err))
			continue
		}

		query := string(content)
		_, err = db.Master.Exec(ctx, query)
		if err != nil {
			logger.Error("Failed to execute seeder", zap.String("file", file.Name()), zap.Error(err))
		} else {
			logger.Info("Successfully executed seeder", zap.String("file", file.Name()))
		}
	}

	logger.Info("Seeding completed!")
	os.Exit(0)
}
