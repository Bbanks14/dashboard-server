// File: cmd/server/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bbanks14/dashboard-server/internal/controllers"
	"github.com/Bbanks14/dashboard-server/internal/middleware"
	"github.com/Bbanks14/dashboard-server/internal/routes"
	"github.com/Bbanks14/dashboard-server/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize logger
	logger := logger.NewLogger()

	// Load application configuration
	cfg, err := config.LoadConfig("./configs/app.yaml")
	if err != nil {
		logger.Fatal("Failed to load configuration", err)
	}

	// Initialize database connection
	db, err := data.NewDatabase(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	// Setup router with all routes
	router := routes.SetupRouter(db, logger, cfg)

	// Start server
	go func() {
		logger.Info("Starting server on port " + cfg.Server.Port)
		if err := router.Run(":" + cfg.Server.Port); err != nil {
			logger.Fatal("Failed to start server", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Server shutting down...")
}
