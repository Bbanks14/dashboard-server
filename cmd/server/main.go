package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Bbanks14/dashboard-server/docs"
	"github.com/Bbanks14/dashboard-server/internal/api"
	"github.com/Bbanks14/dashboard-server/internal/auth"
	"github.com/Bbanks14/dashboard-server/internal/db"
	"github.com/Bbanks14/dashboard-server/internal/repositories"
	"github.com/Bbanks14/dashboard-server/internal/seeds"
	"github.com/Bbanks14/dashboard-server/internal/services"
	"github.com/Bbanks14/dashboard-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	dbConnectTimeout = 30 * time.Second
	serverTimeout    = 5 * time.Second
)

// Config holds application-wide configurations
type Config struct {
	// Server configuration
	Port               string
	Environment        string
	CORSAllowedOrigins []string
	CORSEnabled        bool

	// JWT configuration
	JWTSecret string

	// Google OAuth configuration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Database configuration
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	DBMaxConns          int
	DBMinConns          int
	DBMaxConnLifetime   time.Duration
	DBHealthCheckPeriod time.Duration
	DBConnectTimeout    time.Duration
	DBMaxRetries        int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		CORSAllowedOrigins:  []string{"*"}, // Will be overridden by env
		DBConnectTimeout:    dbConnectTimeout,
		DBHealthCheckPeriod: 30 * time.Second,
		DBMaxConnLifetime:   time.Hour,
	}

	// Server configuration
	cfg.Port = util.GetEnv("SERVER_PORT", "8080")
	cfg.Environment = util.GetEnv("ENVIRONMENT", "development")

	// CORS configuration
	cfg.CORSEnabled = util.GetEnv("CORS_ENABLED", "true") == "true"
	if origins := util.GetEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"); origins != "" {
		cfg.CORSAllowedOrigins = []string{origins}
	}

	// JWT Secret is required
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWTSecret = secret
	} else {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	// Database configuration
	cfg.DBHost = util.GetEnv("DB_HOST", "localhost")
	cfg.DBPort = util.GetEnv("DB_PORT", "5432")
	cfg.DBUser = util.GetEnv("DB_USER", "postgres")
	cfg.DBPassword = util.GetEnv("DB_PASSWORD", "")
	cfg.DBName = util.GetEnv("DB_NAME", "dashboard_db")
	cfg.DBSSLMode = util.GetEnv("DB_SSLMODE", "disable")

	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	// Database connection pool configuration
	cfg.DBMaxConns = util.GetEnvInt("DB_MAX_CONNS", 25)
	cfg.DBMinConns = util.GetEnvInt("DB_MIN_CONNS", 5)
	cfg.DBMaxRetries = util.GetEnvInt("DB_MAX_RETRIES", 3)

	// Parse database connection lifetime
	if lifetime := os.Getenv("DB_MAX_CONN_LIFETIME"); lifetime != "" {
		if d, err := time.ParseDuration(lifetime); err == nil {
			cfg.DBMaxConnLifetime = d
		} else {
			return nil, fmt.Errorf("invalid DB_MAX_CONN_LIFETIME: %w", err)
		}
	}

	// Parse health check period
	if period := os.Getenv("DB_HEALTH_CHECK_PERIOD"); period != "" {
		if d, err := time.ParseDuration(period); err == nil {
			cfg.DBHealthCheckPeriod = d
		} else {
			return nil, fmt.Errorf("invalid DB_HEALTH_CHECK_PERIOD: %w", err)
		}
	}

	// Google OAuth configuration
	cfg.GoogleClientID = util.GetEnv("GOOGLE_CLIENT_ID", "")
	cfg.GoogleClientSecret = util.GetEnv("GOOGLE_CLIENT_SECRET", "")
	cfg.GoogleRedirectURL = util.GetEnv("GOOGLE_REDIRECT_URL", fmt.Sprintf("http://localhost:%s/auth/google/callback", cfg.Port))

	return cfg, nil
}

func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Couldn't load .env file: %v\n", err)
	}

	// Load configuration early for logger setup
	cfg, err := LoadConfig()
	if err != nil {
		fmt.FPrintf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// initialize logger with configuration
	var logger *zap.Logger
	if cfg.Environment == "development" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction(
			// Add stacktrace for errors in production environment
			zap.AddStacktrace(zap.ErrorLevel),
		)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initalize logger: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprint(os.Stderr, "Warning: Failed to sync logger: %v\n", err)
		}
	}()
	zap.ReplaceGlobals(logger)

	// Log the config loading if successful
	logger.Info("Application configuration loaded",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
		zap.Bool("cors_enabled", cfg.CORSEnabled),
		zap.Strings("cors_origins", cfg.CORSAllowedOrigins),
		zap.String("db_host", cfg.DBHost),
		zap.Int("db_max_conns", cfg.DBMaxConns),
	)

	// Add redacted sensitive to the logger
	safeLogger := logger.With(
		zap.String("db_host", cfg.DBHost),
		zap.String("db_port", cfg.DBPort),
		zap.String("db_name", cfg.DBName),
		zap.String("db_user", cfg.DBUser),
		zap.String("db_password", "[REDACTED]"),
	)
	safeLogger.Info("Database configuration")

	// Create context for database connection with retries
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DBConnectTimeout)
	defer cancel()

	// Connect to PostgreSQL database with retry logic
	var database *db.Database
	for i := 0; i < cfg.DBMaxRetries; i++ {
		database, err = db.NewDatabase(ctx, cfg)
		if err == nil {
			break
		}

		logger.Warn("Database connection failed, retrying...",
			zap.Int("attempt", i+1),
			zap.Int("max_retries", cfg.DBMaxRetries),
			zap.Error(err),
		)

		if i < cfg.DBMaxRetries-1 {
			time.Sleep(time.Duration(i+1) * 2 * time.Second) // Exponential backoff
		}
	}

	if err != nil {
		logger.Fatal("Failed to connect to database after retries", zap.Error(err))
	}
	defer database.Close()

	logger.Info("Database connected successfully")

	// Get database pool from the database instance
	dbPool := database.GetPool()

	// Initialize repository layer
	userRepository := repositories.NewUserRepository(dbPool)
	clientRepository := repositories.NewClientRepository(dbPool)
	productRepository := repositories.NewProductRepository(dbPool)
	dashboardRepository := repositories.NewDashboardRepository(dbPool)

	// Seed initial data
	if err := seeds.SeedInitialData(ctx, dbPool); err != nil {
		logger.Fatal("Failed to seed initial data", zap.Error(err))
	}

	// Create Google OAuth config
	googleConfig := auth.NewGoogleOAuthConfig(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectURL,
	)

	// Initialize services
	authService := services.NewAuthService(userRepository, userRepository, cfg.JWTSecret, 7*24*time.Hour, googleConfig)
	dashboardService := services.NewDashboardService(dashboardRepository)
	productService := services.NewProductService(productRepository)
	clientService := services.NewClientService(clientRepository)
	userService := services.NewUserService(userRepository)

	// Initialize API handlers
	authAPI := api.NewAuthAPI(authService)
	dashboardAPI := api.NewDashboardAPI(dashboardService)
	clientAPI := api.NewClientAPI(clientService)
	productAPI := api.NewProductAPI(productService)
	userAPI := api.NewUserAPI(userService)

	// Initialize router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Add CORS middleware if enabled
	if cfg.CORSEnabled {
		router.Use(corsMiddleware(cfg.CORSAllowedOrigins))
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Dashboard API server is running",
			"version": "1.0.0",
			"endpoints": []string{
				"/health",
				"/auth/google",
				"/auth/google/callback",
				"/auth/me",
				"/client",
				"/general",
				"/management",
				"/sales",
			},
		})
	})

	// Setup routes
	setupRoutes(router, authAPI, dashboardAPI, clientAPI, productAPI, userAPI)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Server shutting down...")

	// Create shutdown context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), serverTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupRoutes(
	router *gin.Engine,
	authAPI *api.AuthAPI,
	dashboardAPI *api.DashboardAPI,
	clientAPI *api.ClientAPI,
	productAPI *api.ProductAPI,
	userAPI *api.UserAPI,
) {
	// Public routes
	router.GET("/auth/google", func(c *gin.Context) {
		authAPI.GoogleLoginHandler(c.Writer, c.Request)
	})
	router.GET("/auth/google/callback", func(c *gin.Context) {
		authAPI.GoogleCallbackHandler(c.Writer, c.Request)
	})
	router.GET("/auth/me", func(c *gin.Context) {
		authAPI.GetCurrentUser(c.Writer, c.Request)
	})

	// Protected routes (using Gin middleware)
	protected := router.Group("/")
	protected.Use(auth.GinAuthMiddleware(authAPI.AuthService.JWTSecret))

	// Client routes
	client := protected.Group("/client")
	{
		client.GET("", func(c *gin.Context) {
			clientAPI.GetClients(c.Writer, c.Request)
		})
		client.GET("/:id", func(c *gin.Context) {
			clientAPI.GetClient(c.Writer, c.Request)
		})
		client.POST("", func(c *gin.Context) {
			clientAPI.CreateClient(c.Writer, c.Request)
		})
		client.PUT("/:id", func(c *gin.Context) {
			clientAPI.UpdateClient(c.Writer, c.Request)
		})
		client.DELETE("/:id", func(c *gin.Context) {
			clientAPI.DeleteClient(c.Writer, c.Request)
		})
	}

	// General routes (dashboard)
	general := protected.Group("/general")
	{
		general.GET("/users", func(c *gin.Context) {
			dashboardAPI.GetUsers(c.Writer, c.Request)
		})
		general.GET("/stats", func(c *gin.Context) {
			dashboardAPI.GetDashboardStats(c.Writer, c.Request)
		})
		general.GET("/summary", func(c *gin.Context) {
			dashboardAPI.GetDashboardSummary(c.Writer, c.Request)
		})
	}

	// Management routes
	management := protected.Group("/management")
	{
		management.GET("/products", func(c *gin.Context) {
			productAPI.GetProducts(c.Writer, c.Request)
		})
		management.GET("/products/:id", func(c *gin.Context) {
			productAPI.GetProduct(c.Writer, c.Request)
		})
		management.POST("/products", func(c *gin.Context) {
			productAPI.CreateProduct(c.Writer, c.Request)
		})
		management.PUT("/products/:id", func(c *gin.Context) {
			productAPI.UpdateProduct(c.Writer, c.Request)
		})
		management.DELETE("/products/:id", func(c *gin.Context) {
			productAPI.DeleteProduct(c.Writer, c.Request)
		})
	}

	// Sales routes
	sales := protected.Group("/sales")
	{
		// Add sales endpoints as needed
		sales.GET("/reports", func(c *gin.Context) {
			// Add sales reports endpoint
		})
	}
}
