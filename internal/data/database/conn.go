package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

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

// DBPool is the global database connection pool.
var (
	DBPool      *pgxpool.Pool
	initialized bool
	mu          sync.RWMutex
)

// ConnectDB establishes a connection to the database with retries.
func ConnectDB(ctx context.Context, dsn string, config Config) error {
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if initialized && DBPool != nil {
		return nil // Already connected
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

		DBPool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			lastErr = err
			log.Printf("Database connection attempt %d/%d failed: %v", attempt, config.MaxRetries, err)
			if attempt < config.MaxRetries {
				backoff := time.Duration(1<<attempt) * time.Second // Exponential backoff
				log.Printf("Retrying in %v...", backoff)
				time.Sleep(backoff)
			}
			continue
		}

		// Verify the connection
		pingCtx, cancel := context.WithTimeout(ctx, config.ConnectTimeout)
		defer cancel()

		if err := DBPool.Ping(pingCtx); err != nil {
			DBPool.Close()
			DBPool = nil
			lastErr = fmt.Errorf("database ping failed: %w", err)
			continue
		}

		initialized = true
		log.Println("Successfully connected to database")
		return nil
	}

	return fmt.Errorf("failed to connect to database after %d attempts: %w", config.MaxRetries, lastErr)
}

// validateConfig validates the database configuration.
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

// CloseDB closes the database connection pool with a timeout.
func CloseDB(timeout time.Duration) error {
	mu.Lock()
	defer mu.Unlock()

	if !initialized || DBPool == nil {
		return nil // No pool to close
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		DBPool.Close()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Database connection pool closed successfully")
	case <-ctx.Done():
		log.Println("Database connection pool closed with timeout")
	}

	DBPool = nil
	initialized = false
	return nil
}

// GetDB returns the database connection pool.
func GetDB() (*pgxpool.Pool, error) {
	mu.RLock()
	defer mu.RUnlock()

	if !initialized || DBPool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	return DBPool, nil
}

// IsInitialized checks if the database connection pool is initialized.
func IsInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()
	return initialized
}

// HealthCheck performs a health check on the database connection pool.
func HealthCheck(ctx context.Context) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	return db.Ping(ctx)
}

// Stats returns connection pool statistics.
func Stats() (pgxpool.Stat, error) {
	db, err := GetDB()
	if err != nil {
		return pgxpool.Stat{}, err
	}
	stat := db.Stat()
	return *stat, nil
}
