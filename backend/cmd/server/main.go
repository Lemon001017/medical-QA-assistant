package main

import (
	"fmt"
	"medical-qa-assistant/api"
	"medical-qa-assistant/internal/config"
	"medical-qa-assistant/internal/logger"
	"medical-qa-assistant/internal/models"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Initialize structured logger first.
	logger.Init()
	defer logger.Sync()

	// Load .env file for local development; fall back to existing environment.
	if err := godotenv.Load(); err != nil {
		logger.L.Warn("no .env file loaded", zap.Error(err))
	}

	// Load configuration
	cfg := config.Load()
	logger.L.Info("configuration loaded",
		zap.String("db_host", cfg.DBHost),
		zap.String("db_name", cfg.DBName),
		zap.String("llm_provider", cfg.LLMProvider),
		zap.String("port", cfg.Port),
	)

	// Connect to database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.L.Fatal("failed to connect to database", zap.Error(err))
	}

	// Auto migrate (document chunks and vectors are stored in Chroma, not MySQL)
	if err := db.AutoMigrate(&models.User{}, &models.Document{}); err != nil {
		logger.L.Fatal("failed to migrate database", zap.Error(err))
	}

	// Setup routes
	router := api.SetupRoutes(db, cfg)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.L.Info("server starting",
		zap.String("addr", addr),
		zap.String("llm_provider", cfg.LLMProvider),
		zap.String("chroma_base_url", cfg.ChromaBaseURL),
		zap.String("chroma_collection", cfg.ChromaCollection),
	)
	if err := router.Run(addr); err != nil {
		logger.L.Fatal("failed to start server", zap.Error(err))
	}

}
