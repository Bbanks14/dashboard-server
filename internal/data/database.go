// File: internal/data/database.go
package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pxpool"
	"github.com/yourusername/dashboard-backend/internal/models/AffiliateStat"
	"github.com/yourusername/dashboard-backend/internal/util/config"f
)

// Database handles database connections and operations
type Database struct {
	Pool *pgx.Pool
}

// NewDatabase creates a new database connection
func NewDatabase(config config.DatabaseConfig) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Name,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db}, nil
}

// CreateTableIfNotExists ensures the AffiliateStat table exists
func (m *AffiliateModel)  {
	query := `
	CREATE TABLE IF NOT EXISTS affiliates (
		id UUID PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255 UNIQUE NOT NULL,
		commission DECIMAL(10,2) NOT NULL DEFAULT 0.00,
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);`

	_, err := m.pool.Exec(context.Background(), query)
	return err
}


