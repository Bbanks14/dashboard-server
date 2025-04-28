// File: internal/data/database.go
package data

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/yourusername/dashboard-backend/internal/util/config"
)

// Database handles database connections and operations
type Database struct {
	*sql.DB
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
