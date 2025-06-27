package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/controllers"
	"github.com/Bbanks14/dashboard-server/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	dbConnectTimeout = 10 * time.Second
	serverTimeout    = 5 * time.Second
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("WARNING: Error loading .env file - using environment variables")
	}

	// Access environment variables
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/dashboard_db"
		log.Println("WARNING: DATABASE_DSN not set, using default:", dsn)
	}

	// Parse database configuration from environment
	dbConfig, err := parseDBConfig()
	if err != nil {
		log.Fatalf("Invalid database configuration: %v", err)
	}

	// Create context for database connection
	ctx, cancel := context.WithTimeout(context.Background(), dbConnectTimeout)
	defer cancel()

	// Connect to PostgreSQL database
	if err := db.ConnectDB(ctx, dsn, dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.CloseDB(5 * time.Second); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Initialize controllers
	dashboardController := controllers.NewDashboardController()
	clientController := controllers.NewClientController(clientService)
	productController := controllers.NewProductController()
	customerController := controllers.NewCustomerController()
	transactionController := controllers.NewTransactionController()
	geographyController := controllers.NewGeographyController()

	// Initialize router
	router := gin.Default()

	// Define API routes
	api := router.Group("/api")
	{
		// Dashboard routes
		dashboard := api.Group("/dashboard")
		{
			dashboard.GET("/users", dashboardController.GetUser)
			dashboard.GET("/stats", dashboardController.GetDashboardStats)
			dashboard.GET("/summary", dashboardController.GetDashboardSummary)
		}

		// Product routes
		products := api.Group("/products")
		{
			products.GET("", productController.GetProducts)
		}

		// Customer routes
		customers := api.Group("/customers")
		{
			customers.GET("", customerController.GetCustomers)
		}

		// Transaction routes
		transactions := api.Group("/transactions")
		{
			transactions.GET("", transactionController.GetTransactions)
		}

		// Geography routes
		geography := api.Group("/geography")
		{
			geography.GET("", geographyController.GetGeography)
		}
		
		// Client routes
		client := api.Group("/clients")
		{
			client.GET("/users", clientController.GetClients)
			client.GET("/products", clientController.GetProducts)
			client.GET("/transactions", clientController.GetCustomers)
			client.GET("/geography", clientController.GetTransactions)
	}

	// Get server port
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server for graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server shutting down...")

	// Create shutdown context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), serverTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

	func setupRoutes() *gin.Engine {
	// Initialize database
	db := database.NewDatabase() // Your existing database initialization
	
	// Initialize repository layer
	clientRepository := repositories.NewClientRepository(db)
	
	// Initialize service layer with repository dependency
	clientService := services.NewClientService(clientRepository)
	
	// Initialize controller with service dependency
	clientController := controllers.NewClientController(clientService)
	
	// Setup Gin router
	router := gin.Default()
	
	// Define API routes
	api := router.Group("/api")
	{
		api.GET("/clients", clientController.GetClients)
		api.GET("/products", clientController.GetProducts)
		api.GET("/customers", clientController.GetCustomers)
		api.GET("/transactions", clientController.GetTransactions)
		api.GET("/geography", clientController.GetGeography)
	}
	
	return router
}

// Alternative: Dependency injection container approach
type Dependencies struct {
	DB                *database.Database
	ClientRepository  repositories.ClientRepositoryInterface
	ClientService     services.ClientServiceInterface
	ClientController  *controllers.ClientController
}

func NewDependencies() *Dependencies {
	// Initialize in order of dependencies
	db := database.NewDatabase()
	clientRepo := repositories.NewClientRepository(db)
	clientService := services.NewClientService(clientRepo)
	clientController := controllers.NewClientController(clientService)
	
	return &Dependencies{
		DB:                db,
		ClientRepository:  clientRepo,
		ClientService:     clientService,
		ClientController:  clientController,
	}
}

func setupRoutesWithContainer() *gin.Engine {
	deps := NewDependencies()
	
	router := gin.Default()
	
	api := router.Group("/api")
	{
		api.GET("/clients", deps.ClientController.GetClients)
		api.GET("/products", deps.ClientController.GetProducts)
		api.GET("/customers", deps.ClientController.GetCustomers)
		api.GET("/transactions", deps.ClientController.GetTransactions)
		api.GET("/geography", deps.ClientController.GetGeography)
	}
	
	return router
}

// parseDBConfig reads database configuration from environment variables
func parseDBConfig() (db.Config, error) {
	config := db.Config{
		MaxConns:          10, // Default values
		MinConns:          2,
		MaxConnLifetime:   30 * time.Minute,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		ConnectTimeout:    10 * time.Second,
		MaxRetries:        3,
	}

	if val := os.Getenv("DB_MAX_CONNS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			config.MaxConns = n
		}
	}

	if val := os.Getenv("DB_MIN_CONNS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			config.MinConns = n
		}
	}

	if val := os.Getenv("DB_MAX_CONN_LIFETIME"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.MaxConnLifetime = d
		}
	}

	if val := os.Getenv("DB_MAX_CONN_IDLE_TIME"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.MaxConnIdleTime = d
		}
	}

	if val := os.Getenv("DB_HEALTH_CHECK_PERIOD"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.HealthCheckPeriod = d
		}
	}

	if val := os.Getenv("DB_CONNECT_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.ConnectTimeout = d
		}
	}

	if val := os.Getenv("DB_MAX_RETRIES"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			config.MaxRetries = n
		}
	}

	// Validate configuration
	if config.MaxConns <= 0 {
		return config, fmt.Errorf("DB_MAX_CONNS must be > 0")
	}
	if config.MinConns < 0 {
		return config, fmt.Errorf("DB_MIN_CONNS cannot be negative")
	}
	if config.MaxRetries < 1 {
		return config, fmt.Errorf("DB_MAX_RETRIES must be >= 1")
	}

	return config, nil
}
