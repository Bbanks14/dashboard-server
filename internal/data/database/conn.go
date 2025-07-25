package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds database connection pool configuration.
type Config struct {
	MaxConns          int
	MinConns          int
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
	MaxRetries        int
}

// Database wraps the pgxpool.Pool and provides additional functionality
type Database struct {
	pool *pgxpool.Pool
	mu   sync.RWMutex
}

// NewDatabase creates a new Database instance with connection pool
func NewDatabase(ctx context.Context, config interface{}) (*Database, error) {
	// Handle config conversion - assuming it's from your main Config struct
	dbConfig, err := convertConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}

	// Build DSN from environment variables
	dsn, err := buildDSN()
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	db := &Database{}
	if err := db.connect(ctx, dsn, dbConfig); err != nil {
		return nil, err
	}

	return db, nil
}

// convertConfig converts the main Config to database Config
func convertConfig(config interface{}) (Config, error) {
	// Type assertion to handle the config from main.go
	// You'll need to adjust this based on your actual config struct
	type MainConfig struct {
		DBMaxConns          int
		DBMinConns          int
		DBMaxConnLifetime   time.Duration
		DBHealthCheckPeriod time.Duration
		DBConnectTimeout    time.Duration
		DBMaxRetries        int
	}

	var mainCfg MainConfig

	// Try to convert the config - this is a simplified approach
	// You might need to adjust this based on your actual config structure
	if cfg, ok := config.(*MainConfig); ok {
		mainCfg = *cfg
	} else {
		// Set reasonable defaults if conversion fails
		mainCfg = MainConfig{
			DBMaxConns:          10,
			DBMinConns:          2,
			DBMaxConnLifetime:   time.Hour,
			DBHealthCheckPeriod: 30 * time.Second,
			DBConnectTimeout:    10 * time.Second,
			DBMaxRetries:        3,
		}
	}

	return Config{
		MaxConns:          mainCfg.DBMaxConns,
		MinConns:          mainCfg.DBMinConns,
		MaxConnLifetime:   mainCfg.DBMaxConnLifetime,
		MaxConnIdleTime:   30 * time.Second, // Default value
		HealthCheckPeriod: mainCfg.DBHealthCheckPeriod,
		ConnectTimeout:    mainCfg.DBConnectTimeout,
		MaxRetries:        mainCfg.DBMaxRetries,
	}, nil
}

// buildDSN constructs the database connection string from environment variables
func buildDSN() (string, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "dashboard")
	sslmode := getEnv("DB_SSLMODE", "disable")

	if password == "" {
		return "", errors.New("DB_PASSWORD environment variable is required")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)

	return dsn, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// connect establishes a connection to the database with retries and logging
func (db *Database) connect(ctx context.Context, dsn string, config Config) error {
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if db.pool != nil {
		log.Println("Database connection already established")
		return nil
	}

	var lastErr error
	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		poolConfig, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			return fmt.Errorf("failed to parse DSN: %w", err)
		}

		poolConfig.MaxConns = int32(config.MaxConns)
		poolConfig.MinConns = int32(config.MinConns)
		poolConfig.MaxConnLifetime = config.MaxConnLifetime
		poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
		poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

		// Add connection logging
		poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			log.Printf("New database connection established to %s", conn.Config().Host)
			return nil
		}

		poolConfig.BeforeClose = func(conn *pgx.Conn) {
			log.Printf("Closing database connection to %s", conn.Config().Host)
		}

		db.pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			lastErr = fmt.Errorf("connection attempt %d/%d failed: %w", attempt, config.MaxRetries, err)
			log.Println(lastErr)
			if attempt < config.MaxRetries {
				backoff := time.Duration(1<<uint(attempt)) * time.Second
				log.Printf("Retrying in %v...", backoff)
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			continue
		}

		// Verify the connection
		pingCtx, cancel := context.WithTimeout(ctx, config.ConnectTimeout)
		if err := db.pool.Ping(pingCtx); err != nil {
			cancel()
			db.pool.Close()
			db.pool = nil
			lastErr = fmt.Errorf("database ping failed: %w", err)
			log.Println(lastErr)
			continue
		}
		cancel()

		log.Println("Successfully connected to database")

		// Log initial pool stats
		stats := db.pool.Stat()
		log.Printf("Database pool stats: "+
			"MaxConns=%d, TotalConns=%d, IdleConns=%d, AcquiredConns=%d",
			stats.MaxConns(), stats.TotalConns(), stats.IdleConns(), stats.AcquiredConns())

		return nil
	}

	return fmt.Errorf("failed to connect to database after %d attempts: %w", config.MaxRetries, lastErr)
}

// GetPool returns the database connection pool
func (db *Database) GetPool() *pgxpool.Pool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.pool
}

// Close closes the database connection pool
func (db *Database) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.pool == nil {
		log.Println("No database pool to close")
		return
	}

	// Log stats before closing
	stats := db.pool.Stat()
	log.Printf("Closing database pool. Current stats: "+
		"MaxConns=%d, TotalConns=%d, IdleConns=%d, AcquiredConns=%d",
		stats.MaxConns(), stats.TotalConns(), stats.IdleConns(), stats.AcquiredConns())

	db.pool.Close()
	db.pool = nil
	log.Println("Database connection pool closed successfully")
}

// HealthCheck performs a health check on the database connection pool
func (db *Database) HealthCheck(ctx context.Context) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	return db.pool.Ping(ctx)
}

// Stats returns connection pool statistics
func (db *Database) Stats() (pgxpool.Stat, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.pool == nil {
		return pgxpool.Stat{}, errors.New("database connection pool is not initialized")
	}

	stat := db.pool.Stat()
	return *stat, nil
}

// LogStats logs current connection pool statistics
func (db *Database) LogStats() {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.pool == nil {
		log.Println("Database pool not initialized")
		return
	}

	stats := db.pool.Stat()
	log.Printf("Database pool stats: "+
		"MaxConns=%d, TotalConns=%d, IdleConns=%d, AcquiredConns=%d",
		stats.MaxConns(), stats.TotalConns(), stats.IdleConns(), stats.AcquiredConns())
}

// validateConfig validates the database configuration
func validateConfig(config Config) error {
	if config.MaxConns <= 0 {
		return errors.New("MaxConns must be greater than 0")
	}
	if config.MinConns < 0 {
		return errors.New("MinConns cannot be negative")
	}
	if config.MaxConnLifetime <= 0 {
		return errors.New("MaxConnLifetime must be greater than 0")
	}
	if config.MaxConnIdleTime <= 0 {
		return errors.New("MaxConnIdleTime must be greater than 0")
	}
	if config.HealthCheckPeriod <= 0 {
		return errors.New("HealthCheckPeriod must be greater than 0")
	}
	if config.ConnectTimeout <= 0 {
		return errors.New("ConnectTimeout must be greater than 0")
	}
	if config.MaxRetries < 1 {
		return errors.New("MaxRetries must be at least 1")
	}
	return nil
}

// Legacy functions for backward compatibility (if needed)
var (
	globalDB *Database
	initOnce sync.Once
)

// ConnectDB establishes a connection to the database (legacy function)
func ConnectDB(ctx context.Context, dsn string, config Config) error {
	var err error
	initOnce.Do(func() {
		globalDB = &Database{}
		err = globalDB.connect(ctx, dsn, config)
	})
	return err
}

// GetDB returns the global database connection pool (legacy function)
func GetDB() (*pgxpool.Pool, error) {
	if globalDB == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	return globalDB.GetPool(), nil
}

// CloseDB closes the global database connection pool (legacy function)
func CloseDB(timeout time.Duration) error {
	if globalDB != nil {
		globalDB.Close()
	}
	return nil
}
