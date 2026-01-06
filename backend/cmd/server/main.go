package main

import (
	"fmt"
	"log"
	"medical-qa-assistant/api"
	"medical-qa-assistant/internal/config"
	"medical-qa-assistant/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load .env file for local development; fall back to existing environment.
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file loaded: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Connect to database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&models.User{}, &models.Document{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Setup routes
	router := api.SetupRoutes(db, cfg)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
