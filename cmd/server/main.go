package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Bbanks14/dashboard-server/internal/controllers"
	"github.com/Bbanks14/dashboard-server/internal/repositories"
	"github.com/Bbanks14/dashboard-server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	config := loadConfig()

	// Test database connection
	testConnection(config)

	// Initialize database pool
	db := initializeDatabase(config)
	defer db.Close()

	// Initialize Dashboard layers
	dashboardRepo, err := repositories.NewDashboardRepository(db)
	if err != nil {
		log.Fatalf("❌ Failed to create dashboard repository: %v", err)
	}
	dashboardService := services.NewDashboardService(dashboardRepo)
	dashboardController := controllers.NewDashboardController(dashboardService)

	// Setup Gin router
	router := gin.Default()

	// Add global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Register dashboard routes
	dashboardController.RegisterRoutes(router)

	// Add a root health check endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Dashboard Server is running",
			"status":  "healthy",
			"time":    time.Now().UTC(),
			"features": []string{
				"dashboard-overview",
				"revenue-growth",
			},
		})
	})

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("🚀 Server starting on port %s", port)

	// Dashboard endpoints
	log.Printf("📊 Dashboard Endpoints:")
	log.Printf("   - GET  http://localhost:%s/dashboard/test", port)
	log.Printf("   - GET  http://localhost:%s/dashboard/admin/health", port)
	log.Printf("   - GET  http://localhost:%s/dashboard/data/metrics?startDate=2024-01-01&endDate=2024-01-31", port)
	log.Printf("   - GET  http://localhost:%s/dashboard/data/complete?startDate=2024-01-01&endDate=2024-01-31", port)

	// Revenue Growth endpoints
	log.Printf("💰 Revenue Growth Endpoints:")
	log.Printf("   - GET  http://localhost:%s/revenue-growth/test", port)
	log.Printf("   - GET  http://localhost:%s/revenue-growth/admin/health", port)
	log.Printf("   - GET  http://localhost:%s/revenue-growth/data/complete?startDate=2024-01-01&endDate=2024-01-31", port)
	log.Printf("   - GET  http://localhost:%s/revenue-growth/data/by-date?startDate=2024-01-01&endDate=2024-01-31&granularity=monthly", port)
	log.Printf("   - GET  http://localhost:%s/revenue-growth/data/period-comparison?currentStart=2024-01-01&currentEnd=2024-01-31&previousStart=2023-12-01&previousEnd=2023-12-31", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// The rest of the functions remain the same as in your original code...

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func loadConfig() Config {
	return Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   getEnv("DB_NAME", "dashboard_server"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func initializeDatabase(config Config) *pgxpool.Pool {
	connStr := buildConnectionString(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("❌ Failed to parse database config: %v", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = 30
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("❌ Failed to create connection pool: %v", err)
	}

	log.Println("✅ Database pool initialized successfully!")
	return pool
}

func testConnection(config Config) {
	connStr := buildConnectionString(config)
	log.Println("🧪 Testing PostgreSQL connection...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("❌ Connection test failed: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		log.Fatalf("❌ Ping test failed: %v", err)
	}

	log.Println("✅ Connection test successful!")
}

func buildConnectionString(config Config) string {
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.DBName, config.SSLMode)

	if config.Password != "" {
		connStr += fmt.Sprintf(" password=%s", config.Password)
	}

	return connStr
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
